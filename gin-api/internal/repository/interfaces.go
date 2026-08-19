package repository

import (
	"context"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
)

// TrackRepositoryInterface определяет интерфейс для работы с репозиторием треков
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
