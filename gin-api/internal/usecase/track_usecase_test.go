package usecase_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"
	"github.com/stretchr/testify/require"
)

type fakeTrackRepository struct {
	tracks map[string]*domain.Track
}

func newFakeTrackRepository() *fakeTrackRepository {
	return &fakeTrackRepository{tracks: make(map[string]*domain.Track)}
}

func (r *fakeTrackRepository) Create(_ context.Context, track *domain.Track) error {
	if _, exists := r.tracks[track.ID]; exists {
		return domain.ErrTrackAlreadyExists
	}
	copy := *track
	r.tracks[track.ID] = &copy
	return nil
}

func (r *fakeTrackRepository) GetByID(_ context.Context, id string) (*domain.Track, error) {
	track, ok := r.tracks[id]
	if !ok {
		return nil, domain.ErrTrackNotFound
	}
	copy := *track
	return &copy, nil
}

func (r *fakeTrackRepository) Exists(_ context.Context, id string) (bool, error) {
	_, ok := r.tracks[id]
	return ok, nil
}

func (r *fakeTrackRepository) Update(_ context.Context, track *domain.Track) error {
	if _, ok := r.tracks[track.ID]; !ok {
		return domain.ErrTrackNotFound
	}
	copy := *track
	r.tracks[track.ID] = &copy
	return nil
}

func (r *fakeTrackRepository) Delete(_ context.Context, id string) error {
	if _, ok := r.tracks[id]; !ok {
		return domain.ErrTrackNotFound
	}
	delete(r.tracks, id)
	return nil
}

func (r *fakeTrackRepository) GetAllTracks(_ context.Context) ([]*domain.Track, error) {
	result := make([]*domain.Track, 0, len(r.tracks))
	for _, track := range r.tracks {
		copy := *track
		result = append(result, &copy)
	}
	return result, nil
}

type fakeTrackStorage struct {
	objects      map[string][]byte
	presignedURL string
	deleted      []string
}

func newFakeTrackStorage() *fakeTrackStorage {
	return &fakeTrackStorage{objects: make(map[string][]byte)}
}

func (s *fakeTrackStorage) PutObject(
	_ context.Context,
	objectKey string,
	reader io.Reader,
	_ int64,
	_ string,
) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	s.objects[objectKey] = data
	return nil
}

func (s *fakeTrackStorage) DeleteObject(_ context.Context, objectKey string) error {
	delete(s.objects, objectKey)
	s.deleted = append(s.deleted, objectKey)
	return nil
}

func (s *fakeTrackStorage) PresignGetObject(
	_ context.Context,
	_ string,
	_ time.Duration,
) (string, error) {
	return s.presignedURL, nil
}

type fakeMusicBrainzClient struct {
	metadata domain.TrackMetadata
}

func (c fakeMusicBrainzClient) SearchRecordings(
	_ context.Context,
	_, _ string,
	_ int,
) ([]domain.TrackMetadataCandidate, error) {
	return []domain.TrackMetadataCandidate{
		{
			MBID:   c.metadata.MBID,
			Title:  c.metadata.Title,
			Artist: c.metadata.Artist,
			Score:  100,
		},
	}, nil
}

func (c fakeMusicBrainzClient) GetRecording(
	_ context.Context,
	_ string,
) (*domain.TrackMetadata, error) {
	metadata := c.metadata
	return &metadata, nil
}

func TestTrackUseCaseUploadAndContent(t *testing.T) {
	repo := newFakeTrackRepository()
	storage := newFakeTrackStorage()
	storage.presignedURL = "http://localhost:9000/tracks/signed"

	mb := fakeMusicBrainzClient{
		metadata: domain.TrackMetadata{
			MBID:                      "550e8400-e29b-41d4-a716-446655440000",
			Title:                     "Test Track",
			Artist:                    "Test Artist",
			AlbumTitle:                "Test Album",
			Genre:                     []string{"Electronic"},
			Year:                      2025,
			MusicBrainzReleaseID:      "550e8400-e29b-41d4-a716-446655440001",
			MusicBrainzReleaseGroupID: "550e8400-e29b-41d4-a716-446655440002",
		},
	}

	uc := usecase.NewTrackUseCase(repo, storage, mb)
	ctx := context.Background()

	track, err := uc.UploadTrack(ctx, usecase.TrackUpload{
		MusicBrainzID: mb.metadata.MBID,
		Reader:        bytes.NewBufferString("audio-bytes"),
		FileName:      "track.mp3",
		ContentType:   "audio/mpeg",
		Size:          int64(len("audio-bytes")),
	})
	require.NoError(t, err)
	require.Equal(t, mb.metadata.MBID, track.ID)
	require.Equal(t, mb.metadata.Title, track.Title)
	require.Equal(t, mb.metadata.Artist, track.Artist)
	require.NotEmpty(t, track.ObjectKey)
	require.Len(t, storage.objects, 1)

	access, err := uc.GetTrackContent(ctx, track.ID)
	require.NoError(t, err)
	require.Equal(t, storage.presignedURL, access.URL)
	require.Equal(t, 5*time.Minute, access.ExpiresIn)
}

func TestTrackUseCaseRejectsInvalidMBID(t *testing.T) {
	uc := usecase.NewTrackUseCase(
		newFakeTrackRepository(),
		newFakeTrackStorage(),
		fakeMusicBrainzClient{},
	)

	_, err := uc.UploadTrack(context.Background(), usecase.TrackUpload{
		MusicBrainzID: "not-a-uuid",
		Reader:        bytes.NewBufferString("audio"),
		FileName:      "track.mp3",
		ContentType:   "audio/mpeg",
		Size:          5,
	})
	require.ErrorIs(t, err, domain.ErrInvalidTrackID)
}
