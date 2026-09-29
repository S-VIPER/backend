package usecase

import (
	"context"
	"io"
	"time"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/google/uuid"
)

type TrackStorageInterface interface {
	PutObject(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error
	DeleteObject(ctx context.Context, objectKey string) error
	PresignGetObject(ctx context.Context, objectKey string, expires time.Duration) (string, error)
}

type MusicBrainzClientInterface interface {
	SearchRecordings(ctx context.Context, artist, title string, limit int) ([]domain.TrackMetadataCandidate, error)
	GetRecording(ctx context.Context, mbid string) (*domain.TrackMetadata, error)
}

type TrackRepositoryInterface interface {
	Create(ctx context.Context, track *domain.Track) error
	GetByID(ctx context.Context, id string) (*domain.Track, error)
	Exists(ctx context.Context, id string) (bool, error)
	Update(ctx context.Context, track *domain.Track) error
	Delete(ctx context.Context, id string) error
	GetAllTracks(ctx context.Context) ([]*domain.Track, error)
}

type PlaylistRepositoryInterface interface {
	Create(ctx context.Context, playlist *domain.Playlist) error
	GetByID(ctx context.Context, id string) (*domain.Playlist, error)
	Update(ctx context.Context, playlist *domain.Playlist) error
	Delete(ctx context.Context, id string) error
	AddTrack(ctx context.Context, playlistID, trackID string) error
	RemoveTrack(ctx context.Context, playlistID, trackID string) error
}

type UserRepositoryInterface interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Activate(ctx context.Context, userID uuid.UUID) error
}

type EmailVerificationRepositoryInterface interface {
	Create(ctx context.Context, verification *domain.EmailVerification) (*domain.EmailVerification, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.EmailVerification, error)
	UpdateCode(ctx context.Context, id uuid.UUID, codeHash string, expiresAt time.Time, createdAt time.Time) error
	IncrementAttempts(ctx context.Context, id uuid.UUID, maxAttempts int) error
	MarkVerified(ctx context.Context, id uuid.UUID, verifiedAt time.Time) error
}

type RefreshTokenRepositoryInterface interface {
	Create(ctx context.Context, session *domain.RefreshTokenSession) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshTokenSession, error)
	Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error
	Rotate(ctx context.Context, oldID uuid.UUID, newSession *domain.RefreshTokenSession, revokedAt time.Time) error
}
