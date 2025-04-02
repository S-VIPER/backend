package http

import "github.com/S-VIPER/backend/gin-api/internal/domain"

// TrackUseCaseInterface defines the interface for track use case
type TrackUseCaseInterface interface {
	CreateTrack(track *domain.Track) error
	GetTrackByID(id string) (*domain.Track, error)
	UpdateTrack(track *domain.Track) error
	DeleteTrack(id string) error
	GetAllTracks() ([]*domain.Track, error)
}
