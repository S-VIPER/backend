package rustfs

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Endpoint       string
	PublicEndpoint string
	AccessKey      string
	SecretKey      string
	Bucket         string
	Region         string
}

type TrackStorage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
}

func NewTrackStorage(cfg Config) (*TrackStorage, error) {
	if err := validateEndpoint("rustfs endpoint", cfg.Endpoint); err != nil {
		return nil, err
	}
	if err := validateEndpoint("rustfs public endpoint", cfg.PublicEndpoint); err != nil {
		return nil, err
	}
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("rustfs credentials are required")
	}
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, fmt.Errorf("rustfs bucket is required")
	}

	region := strings.TrimSpace(cfg.Region)
	if region == "" {
		region = "us-east-1"
	}

	creds := credentials.NewStaticCredentialsProvider(
		cfg.AccessKey,
		cfg.SecretKey,
		"",
	)

	client := newS3Client(cfg.Endpoint, region, creds)
	publicClient := newS3Client(cfg.PublicEndpoint, region, creds)

	return &TrackStorage{
		client:        client,
		presignClient: s3.NewPresignClient(publicClient),
		bucket:        cfg.Bucket,
	}, nil
}

func newS3Client(
	endpoint string,
	region string,
	creds aws.CredentialsProvider,
) *s3.Client {
	return s3.New(s3.Options{
		BaseEndpoint: aws.String(endpoint),
		Region:       region,
		Credentials:  creds,
		UsePathStyle: true,
	})
}

func (s *TrackStorage) EnsureBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err == nil {
		return nil
	}

	if _, createErr := s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	}); createErr == nil {
		return nil
	}

	// Another backend instance can create the bucket between HeadBucket and
	// CreateBucket. Re-check before returning the original failure.
	if _, headErr := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	}); headErr == nil {
		return nil
	}

	return fmt.Errorf("ensure rustfs bucket %q: %w", s.bucket, err)
}

func (s *TrackStorage) PutObject(
	ctx context.Context,
	objectKey string,
	reader io.Reader,
	size int64,
	contentType string,
) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(objectKey),
		Body:          reader,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put object %q: %w", objectKey, err)
	}

	return nil
}

func (s *TrackStorage) DeleteObject(
	ctx context.Context,
	objectKey string,
) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("delete object %q: %w", objectKey, err)
	}

	return nil
}

func (s *TrackStorage) PresignGetObject(
	ctx context.Context,
	objectKey string,
	expires time.Duration,
) (string, error) {
	if expires <= 0 {
		return "", fmt.Errorf("presign expiration must be positive")
	}

	presigned, err := s.presignClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(objectKey),
		},
		s3.WithPresignExpires(expires),
	)
	if err != nil {
		return "", fmt.Errorf("presign object %q: %w", objectKey, err)
	}

	return presigned.URL, nil
}

func validateEndpoint(name, raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("%s is required", name)
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid %s: %w", name, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%s must use http or https", name)
	}
	if parsed.Host == "" {
		return fmt.Errorf("%s host is required", name)
	}

	return nil
}

var _ interface {
	PutObject(context.Context, string, io.Reader, int64, string) error
	DeleteObject(context.Context, string) error
	PresignGetObject(context.Context, string, time.Duration) (string, error)
} = (*TrackStorage)(nil)
