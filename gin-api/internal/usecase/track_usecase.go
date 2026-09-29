package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/google/uuid"
)

const (
	DefaultTrackMaxUploadSize int64         = 100 << 20
	DefaultTrackContentURLTTL time.Duration = 5 * time.Minute
)

type TrackUpload struct {
	MusicBrainzID string
	Reader        io.Reader
	FileName      string
	ContentType   string
	Size          int64
}

type TrackContentAccess struct {
	URL       string
	ExpiresIn time.Duration
}

type TrackUseCase struct {
	repo          TrackRepositoryInterface
	storage       TrackStorageInterface
	musicBrainz   MusicBrainzClientInterface
	maxUploadSize int64
	contentURLTTL time.Duration
}

func NewTrackUseCase(
	repo TrackRepositoryInterface,
	storage TrackStorageInterface,
	musicBrainz MusicBrainzClientInterface,
) *TrackUseCase {
	return &TrackUseCase{
		repo:          repo,
		storage:       storage,
		musicBrainz:   musicBrainz,
		maxUploadSize: DefaultTrackMaxUploadSize,
		contentURLTTL: DefaultTrackContentURLTTL,
	}
}

func (uc *TrackUseCase) WithMaxUploadSize(size int64) *TrackUseCase {
	if size > 0 {
		uc.maxUploadSize = size
	}
	return uc
}

func (uc *TrackUseCase) WithContentURLTTL(ttl time.Duration) *TrackUseCase {
	if ttl > 0 {
		uc.contentURLTTL = ttl
	}

	return uc
}

func (uc *TrackUseCase) SearchTrackMetadata(
	ctx context.Context,
	artist string,
	title string,
	limit int,
) ([]domain.TrackMetadataCandidate, error) {
	artist = strings.TrimSpace(artist)
	title = strings.TrimSpace(title)

	if artist == "" || title == "" {
		return nil, domain.ErrInvalidTrack
	}

	if limit <= 0 || limit > 25 {
		limit = 10
	}

	return uc.musicBrainz.SearchRecordings(ctx, artist, title, limit)
}

func (uc *TrackUseCase) UploadTrack(
	ctx context.Context,
	input TrackUpload,
) (*domain.Track, error) {
	if input.Reader == nil {
		return nil, domain.ErrInvalidTrack
	}

	mbid := strings.TrimSpace(input.MusicBrainzID)
	if _, err := uuid.Parse(mbid); err != nil {
		return nil, domain.ErrInvalidTrackID
	}

	if input.Size <= 0 || input.Size > uc.maxUploadSize {
		return nil, domain.ErrInvalidTrack
	}

	fileName := strings.TrimSpace(filepath.Base(input.FileName))
	if fileName == "." || fileName == "" {
		fileName = "track.bin"
	}

	contentType := strings.TrimSpace(input.ContentType)
	if contentType == "" {
		contentType = mime.TypeByExtension(strings.ToLower(filepath.Ext(fileName)))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	metadata, err := uc.musicBrainz.GetRecording(ctx, mbid)
	if err != nil {
		return nil, err
	}

	objectKey := buildObjectKey(mbid, fileName)

	if err := uc.storage.PutObject(
		ctx,
		objectKey,
		input.Reader,
		input.Size,
		contentType,
	); err != nil {
		return nil, fmt.Errorf("store track object: %w", err)
	}

	track := &domain.Track{
		ID:                        metadata.MBID,
		Title:                     metadata.Title,
		Artist:                    metadata.Artist,
		AlbumTitle:                metadata.AlbumTitle,
		AlbumArtURL:               metadata.AlbumArtURL,
		Genre:                     append([]string(nil), metadata.Genre...),
		Year:                      metadata.Year,
		MusicBrainzReleaseID:      metadata.MusicBrainzReleaseID,
		MusicBrainzReleaseGroupID: metadata.MusicBrainzReleaseGroupID,
		ObjectKey:                 objectKey,
		FileName:                  fileName,
		ContentType:               contentType,
		FileSize:                  input.Size,
	}

	if err := validateStoredTrack(track); err != nil {
		_ = uc.storage.DeleteObject(ctx, objectKey)
		return nil, err
	}

	if err := uc.repo.Create(ctx, track); err != nil {
		cleanupErr := uc.storage.DeleteObject(ctx, objectKey)
		if cleanupErr != nil {
			return nil, errors.Join(
				fmt.Errorf("create track: %w", err),
				fmt.Errorf("cleanup uploaded object: %w", cleanupErr),
			)
		}
		return nil, err
	}

	return track, nil
}

func (uc *TrackUseCase) GetTrackByID(
	ctx context.Context,
	id string,
) (*domain.Track, error) {
	id = strings.TrimSpace(id)

	if id == "" {
		return nil, domain.ErrInvalidTrackID
	}

	return uc.repo.GetByID(ctx, id)
}

func (uc *TrackUseCase) GetTrackContent(
	ctx context.Context,
	id string,
) (*TrackContentAccess, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, domain.ErrInvalidTrackID
	}

	track, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if track.ObjectKey == "" {
		return nil, domain.ErrTrackContentUnavailable
	}

	expires := uc.contentURLTTL
	if expires <= 0 {
		expires = DefaultTrackContentURLTTL
	}

	url, err := uc.storage.PresignGetObject(ctx, track.ObjectKey, expires)
	if err != nil {
		return nil, fmt.Errorf("presign track object: %w", err)
	}

	return &TrackContentAccess{
		URL:       url,
		ExpiresIn: expires,
	}, nil
}

func (uc *TrackUseCase) UpdateTrack(
	ctx context.Context,
	track *domain.Track,
) error {
	if track == nil {
		return domain.ErrInvalidTrack
	}

	normalizeTrack(track)

	if err := validateTrackMetadata(track); err != nil {
		return err
	}

	existing, err := uc.repo.GetByID(ctx, track.ID)
	if err != nil {
		return err
	}

	// Object identity and the MusicBrainz recording identity are server-owned.
	track.ObjectKey = existing.ObjectKey
	track.FileName = existing.FileName
	track.ContentType = existing.ContentType
	track.FileSize = existing.FileSize
	track.MusicBrainzReleaseID = existing.MusicBrainzReleaseID
	track.MusicBrainzReleaseGroupID = existing.MusicBrainzReleaseGroupID

	return uc.repo.Update(ctx, track)
}

func (uc *TrackUseCase) DeleteTrack(
	ctx context.Context,
	id string,
) error {
	id = strings.TrimSpace(id)

	if id == "" {
		return domain.ErrInvalidTrackID
	}

	track, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	if track.ObjectKey != "" {
		if err := uc.storage.DeleteObject(ctx, track.ObjectKey); err != nil {
			return fmt.Errorf("delete track object: %w", err)
		}
	}

	return nil
}

func (uc *TrackUseCase) GetAllTracks(
	ctx context.Context,
) ([]*domain.Track, error) {
	return uc.repo.GetAllTracks(ctx)
}

func normalizeTrack(track *domain.Track) {
	track.ID = strings.TrimSpace(track.ID)
	track.Title = strings.TrimSpace(track.Title)
	track.Artist = strings.TrimSpace(track.Artist)
	track.AlbumTitle = strings.TrimSpace(track.AlbumTitle)
	track.AlbumArtURL = strings.TrimSpace(track.AlbumArtURL)
	track.MusicBrainzReleaseID = strings.TrimSpace(track.MusicBrainzReleaseID)
	track.MusicBrainzReleaseGroupID = strings.TrimSpace(track.MusicBrainzReleaseGroupID)

	for i := range track.Genre {
		track.Genre[i] = strings.TrimSpace(track.Genre[i])
	}
}

func validateTrackMetadata(track *domain.Track) error {
	if track == nil {
		return domain.ErrInvalidTrack
	}

	if track.ID == "" {
		return domain.ErrInvalidTrackID
	}
	if _, err := uuid.Parse(track.ID); err != nil {
		return domain.ErrInvalidTrackID
	}
	if track.Title == "" {
		return domain.ErrInvalidTrackTitle
	}
	if track.Artist == "" {
		return domain.ErrInvalidTrackArtist
	}
	if track.Year < 0 || track.Year > 2100 {
		return domain.ErrInvalidTrackYear
	}

	return nil
}

func validateStoredTrack(track *domain.Track) error {
	if err := validateTrackMetadata(track); err != nil {
		return err
	}
	if track.ObjectKey == "" || track.FileName == "" || track.FileSize <= 0 {
		return domain.ErrInvalidTrack
	}
	return nil
}

func buildObjectKey(mbid, fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	if len(ext) > 10 || strings.ContainsAny(ext, "/\\") {
		ext = ""
	}

	return fmt.Sprintf(
		"tracks/%s/%s%s",
		mbid,
		uuid.NewString(),
		ext,
	)
}
