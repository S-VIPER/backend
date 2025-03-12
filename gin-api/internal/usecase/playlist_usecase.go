package usecase

import (
	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/repository"
)

type PlaylistUseCase struct {
	playlistRepo *repository.PlaylistRepository
	trackRepo    *repository.TrackRepository
}

func NewPlaylistUseCase(playlistRepo *repository.PlaylistRepository, trackRepo *repository.TrackRepository) *PlaylistUseCase {
	return &PlaylistUseCase{playlistRepo: playlistRepo, trackRepo: trackRepo}
}

func (uc *PlaylistUseCase) CreatePlaylist(playlist *domain.Playlist) error {
	return uc.playlistRepo.Create(playlist)
}

func (uc *PlaylistUseCase) GetPlaylistByID(id string) (*domain.Playlist, error) {
	return uc.playlistRepo.GetByID(id)
}

func (uc *PlaylistUseCase) UpdatePlaylist(playlist *domain.Playlist) error {
	return uc.playlistRepo.Update(playlist)
}

func (uc *PlaylistUseCase) DeletePlaylist(id string) error {
	return uc.playlistRepo.Delete(id)
}

func (uc *PlaylistUseCase) AddTrackToPlaylist(playlistID, trackID string) error {
	// Проверяем, существует ли трек
	_, err := uc.trackRepo.GetByID(trackID)
	if err != nil {
		return err
	}

	return uc.playlistRepo.AddTrack(playlistID, trackID)
}

func (uc *PlaylistUseCase) RemoveTrackFromPlaylist(playlistID, trackID string) error {
	return uc.playlistRepo.RemoveTrack(playlistID, trackID)
}
