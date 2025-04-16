package usecase

import (
	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/repository"
)

type TrackUseCase struct {
	repo repository.TrackRepositoryInterface
}

func NewTrackUseCase(repo repository.TrackRepositoryInterface) *TrackUseCase {
	return &TrackUseCase{repo: repo}
}

func (uc *TrackUseCase) CreateTrack(track *domain.Track) error {
	return uc.repo.Create(track)
}

func (uc *TrackUseCase) GetTrackByID(id string) (*domain.Track, error) {
	return uc.repo.GetByID(id)
}

func (uc *TrackUseCase) UpdateTrack(track *domain.Track) error {
	return uc.repo.Update(track)
}

func (uc *TrackUseCase) DeleteTrack(id string) error {
	return uc.repo.Delete(id)
}

// GetAllTracks retrieves all tracks from the repository
func (uc *TrackUseCase) GetAllTracks() ([]*domain.Track, error) {
	return uc.repo.GetAllTracks()
}
