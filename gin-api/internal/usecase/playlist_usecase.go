package usecase

import (
	"context"
	"strings"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/repository"
)

type PlaylistUseCase struct {
	playlistRepo repository.PlaylistRepositoryInterface
	trackRepo    repository.TrackRepositoryInterface
}

func NewPlaylistUseCase(
	playlistRepo repository.PlaylistRepositoryInterface,
	trackRepo repository.TrackRepositoryInterface,
) *PlaylistUseCase {
	return &PlaylistUseCase{
		playlistRepo: playlistRepo,
		trackRepo:    trackRepo,
	}
}

func (uc *PlaylistUseCase) CreatePlaylist(
	ctx context.Context,
	playlist *domain.Playlist,
) error {
	if playlist == nil {
		return domain.ErrInvalidPlaylist
	}

	normalizePlaylist(playlist)

	if err := validatePlaylist(playlist); err != nil {
		return err
	}

	return uc.playlistRepo.Create(ctx, playlist)
}

func (uc *PlaylistUseCase) GetPlaylistByID(
	ctx context.Context,
	id string,
) (*domain.Playlist, error) {
	id = strings.TrimSpace(id)

	if id == "" {
		return nil, domain.ErrInvalidPlaylistID
	}

	return uc.playlistRepo.GetByID(ctx, id)
}

func (uc *PlaylistUseCase) UpdatePlaylist(
	ctx context.Context,
	playlist *domain.Playlist,
) error {
	if playlist == nil {
		return domain.ErrInvalidPlaylist
	}

	normalizePlaylist(playlist)

	if err := validatePlaylist(playlist); err != nil {
		return err
	}

	return uc.playlistRepo.Update(ctx, playlist)
}

func (uc *PlaylistUseCase) DeletePlaylist(
	ctx context.Context,
	id string,
) error {
	id = strings.TrimSpace(id)

	if id == "" {
		return domain.ErrInvalidPlaylistID
	}

	return uc.playlistRepo.Delete(ctx, id)
}

func (uc *PlaylistUseCase) AddTrackToPlaylist(
	ctx context.Context,
	playlistID string,
	trackID string,
) error {
	playlistID = strings.TrimSpace(playlistID)
	trackID = strings.TrimSpace(trackID)

	if playlistID == "" {
		return domain.ErrInvalidPlaylistID
	}

	if trackID == "" {
		return domain.ErrInvalidTrackID
	}

	// Проверяем существование трека.
	if _, err := uc.trackRepo.GetByID(ctx, trackID); err != nil {
		return err
	}

	// Проверяем существование playlist-а
	if _, err := uc.playlistRepo.GetByID(ctx, playlistID); err != nil {
		return err
	}

	// AddTrack в repository должен вернуть
	// ErrPlaylistNotFound, если playlist не существует.
	return uc.playlistRepo.AddTrack(
		ctx,
		playlistID,
		trackID,
	)
}

func (uc *PlaylistUseCase) RemoveTrackFromPlaylist(
	ctx context.Context,
	playlistID string,
	trackID string,
) error {
	playlistID = strings.TrimSpace(playlistID)
	trackID = strings.TrimSpace(trackID)

	if playlistID == "" {
		return domain.ErrInvalidPlaylistID
	}

	if trackID == "" {
		return domain.ErrInvalidTrackID
	}

	return uc.playlistRepo.RemoveTrack(
		ctx,
		playlistID,
		trackID,
	)
}

func normalizePlaylist(playlist *domain.Playlist) {
	playlist.Name = strings.TrimSpace(playlist.Name)

	for i := range playlist.Tracks {
		playlist.Tracks[i] = strings.TrimSpace(playlist.Tracks[i])
	}
}

func validatePlaylist(playlist *domain.Playlist) error {
	if playlist == nil {
		return domain.ErrInvalidPlaylist
	}

	if strings.TrimSpace(playlist.ID) == "" {
		return domain.ErrInvalidPlaylistID
	}

	if strings.TrimSpace(playlist.Name) == "" {
		return domain.ErrInvalidPlaylistName
	}

	return nil
}
