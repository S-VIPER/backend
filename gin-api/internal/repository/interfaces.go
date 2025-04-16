package repository

import "github.com/S-VIPER/backend/gin-api/internal/domain"

// TrackRepositoryInterface определяет интерфейс для работы с репозиторием треков
type TrackRepositoryInterface interface {
	Create(track *domain.Track) error
	GetByID(id string) (*domain.Track, error)
	Update(track *domain.Track) error
	Delete(id string) error
	GetAllTracks() ([]*domain.Track, error)
}
