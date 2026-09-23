package usecase

import (
	"context"
	"strings"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/google/uuid"
)

type PlaylistUseCase struct {
	playlistRepo PlaylistRepositoryInterface
	trackRepo    TrackRepositoryInterface
}

func NewPlaylistUseCase(
	playlistRepo PlaylistRepositoryInterface,
	trackRepo TrackRepositoryInterface,
) *PlaylistUseCase {
	return &PlaylistUseCase{
		playlistRepo: playlistRepo,
		trackRepo:    trackRepo,
	}
}

func (uc *PlaylistUseCase) CreatePlaylist(
	ctx context.Context,
	ownerID uuid.UUID,
	playlist *domain.Playlist,
) error {
	if ownerID == uuid.Nil || playlist == nil {
		return domain.ErrInvalidPlaylist
	}

	normalizePlaylist(playlist)

	if playlist.Visibility == "" {
		playlist.Visibility = domain.PlaylistVisibilityPrivate
	}

	if err := validatePlaylistForCreate(playlist); err != nil {
		return err
	}

	playlist.OwnerID = ownerID

	return uc.playlistRepo.Create(ctx, playlist)
}

func (uc *PlaylistUseCase) GetPlaylistByID(
	ctx context.Context,
	viewerID *uuid.UUID,
	id string,
) (*domain.Playlist, error) {
	id = strings.TrimSpace(id)

	if id == "" {
		return nil, domain.ErrInvalidPlaylistID
	}

	playlist, err := uc.playlistRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if playlist.Visibility == domain.PlaylistVisibilityPublic {
		return playlist, nil
	}

	if viewerID != nil && *viewerID == playlist.OwnerID {
		return playlist, nil
	}

	// Do not disclose the existence of private playlists to other users.
	return nil, domain.ErrPlaylistNotFound
}

func (uc *PlaylistUseCase) UpdatePlaylist(
	ctx context.Context,
	ownerID uuid.UUID,
	playlist *domain.Playlist,
) error {
	if ownerID == uuid.Nil || playlist == nil {
		return domain.ErrInvalidPlaylist
	}

	playlist.ID = strings.TrimSpace(playlist.ID)
	if playlist.ID == "" {
		return domain.ErrInvalidPlaylistID
	}

	existing, err := uc.playlistRepo.GetByID(ctx, playlist.ID)
	if err != nil {
		return err
	}

	if existing.OwnerID != ownerID {
		return domain.ErrPlaylistNotFound
	}

	normalizePlaylist(playlist)
	if playlist.Visibility == "" {
		playlist.Visibility = existing.Visibility
	}

	if err := validatePlaylistForUpdate(playlist); err != nil {
		return err
	}

	playlist.OwnerID = existing.OwnerID
	playlist.Tracks = existing.Tracks

	return uc.playlistRepo.Update(ctx, playlist)
}

func (uc *PlaylistUseCase) DeletePlaylist(
	ctx context.Context,
	ownerID uuid.UUID,
	id string,
) error {
	if ownerID == uuid.Nil {
		return domain.ErrUnauthorized
	}

	id = strings.TrimSpace(id)

	if id == "" {
		return domain.ErrInvalidPlaylistID
	}

	playlist, err := uc.playlistRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if playlist.OwnerID != ownerID {
		return domain.ErrPlaylistNotFound
	}

	return uc.playlistRepo.Delete(ctx, id)
}

func (uc *PlaylistUseCase) AddTrackToPlaylist(
	ctx context.Context,
	ownerID uuid.UUID,
	playlistID string,
	trackID string,
) error {
	playlistID = strings.TrimSpace(playlistID)
	trackID = strings.TrimSpace(trackID)

	if ownerID == uuid.Nil {
		return domain.ErrUnauthorized
	}

	if playlistID == "" {
		return domain.ErrInvalidPlaylistID
	}

	if trackID == "" {
		return domain.ErrInvalidTrackID
	}

	playlist, err := uc.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return err
	}

	if playlist.OwnerID != ownerID {
		return domain.ErrPlaylistNotFound
	}

	if _, err := uc.trackRepo.GetByID(ctx, trackID); err != nil {
		return err
	}

	return uc.playlistRepo.AddTrack(ctx, playlistID, trackID)
}

func (uc *PlaylistUseCase) RemoveTrackFromPlaylist(
	ctx context.Context,
	ownerID uuid.UUID,
	playlistID string,
	trackID string,
) error {
	playlistID = strings.TrimSpace(playlistID)
	trackID = strings.TrimSpace(trackID)

	if ownerID == uuid.Nil {
		return domain.ErrUnauthorized
	}

	if playlistID == "" {
		return domain.ErrInvalidPlaylistID
	}

	if trackID == "" {
		return domain.ErrInvalidTrackID
	}

	playlist, err := uc.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return err
	}

	if playlist.OwnerID != ownerID {
		return domain.ErrPlaylistNotFound
	}

	return uc.playlistRepo.RemoveTrack(ctx, playlistID, trackID)
}

func normalizePlaylist(playlist *domain.Playlist) {
	playlist.Name = strings.TrimSpace(playlist.Name)

	for i := range playlist.Tracks {
		playlist.Tracks[i] = strings.TrimSpace(playlist.Tracks[i])
	}
}

func validatePlaylistForCreate(playlist *domain.Playlist) error {
	if playlist == nil {
		return domain.ErrInvalidPlaylist
	}

	if strings.TrimSpace(playlist.Name) == "" {
		return domain.ErrInvalidPlaylistName
	}

	if playlist.Visibility != domain.PlaylistVisibilityPrivate &&
		playlist.Visibility != domain.PlaylistVisibilityPublic {
		return domain.ErrInvalidPlaylist
	}

	return nil
}

func validatePlaylistForUpdate(playlist *domain.Playlist) error {
	if playlist == nil {
		return domain.ErrInvalidPlaylist
	}

	if strings.TrimSpace(playlist.ID) == "" {
		return domain.ErrInvalidPlaylistID
	}

	if strings.TrimSpace(playlist.Name) == "" {
		return domain.ErrInvalidPlaylistName
	}

	if playlist.Visibility != domain.PlaylistVisibilityPrivate &&
		playlist.Visibility != domain.PlaylistVisibilityPublic {
		return domain.ErrInvalidPlaylist
	}

	return nil
}
