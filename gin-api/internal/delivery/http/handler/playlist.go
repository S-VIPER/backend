package handler

import (
	"context"

	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/api"
	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"
)

type PlaylistHandler struct {
	useCase *usecase.PlaylistUseCase
}

func NewPlaylistHandler(
	useCase *usecase.PlaylistUseCase,
) *PlaylistHandler {
	return &PlaylistHandler{
		useCase: useCase,
	}
}

func (h *PlaylistHandler) CreatePlaylist(
	ctx context.Context,
	request api.CreatePlaylistRequestObject,
) (api.CreatePlaylistResponseObject, error) {
	playlist := &domain.Playlist{
		Name: request.Body.Name,
	}

	if err := h.useCase.CreatePlaylist(ctx, playlist); err != nil {
		return nil, err
	}

	return api.CreatePlaylist201JSONResponse{
		Data: toAPIPlaylist(playlist),
	}, nil
}

func (h *PlaylistHandler) GetPlaylistByID(
	ctx context.Context,
	request api.GetPlaylistByIDRequestObject,
) (api.GetPlaylistByIDResponseObject, error) {
	playlist, err := h.useCase.GetPlaylistByID(
		ctx,
		request.PlaylistId,
	)
	if err != nil {
		return nil, err
	}

	return api.GetPlaylistByID200JSONResponse{
		Data: toAPIPlaylist(playlist),
	}, nil
}

func (h *PlaylistHandler) UpdatePlaylist(
	ctx context.Context,
	request api.UpdatePlaylistRequestObject,
) (api.UpdatePlaylistResponseObject, error) {
	playlist := &domain.Playlist{
		ID:   request.PlaylistId,
		Name: request.Body.Name,
	}

	if err := h.useCase.UpdatePlaylist(ctx, playlist); err != nil {
		return nil, err
	}

	return api.UpdatePlaylist200JSONResponse{
		Data: toAPIPlaylist(playlist),
	}, nil
}

func (h *PlaylistHandler) DeletePlaylist(
	ctx context.Context,
	request api.DeletePlaylistRequestObject,
) (api.DeletePlaylistResponseObject, error) {
	if err := h.useCase.DeletePlaylist(
		ctx,
		request.PlaylistId,
	); err != nil {
		return nil, err
	}

	return api.DeletePlaylist204Response{}, nil
}

func (h *PlaylistHandler) AddTrackToPlaylist(
	ctx context.Context,
	request api.AddTrackToPlaylistRequestObject,
) (api.AddTrackToPlaylistResponseObject, error) {
	if err := h.useCase.AddTrackToPlaylist(
		ctx,
		request.PlaylistId,
		request.TrackId,
	); err != nil {
		return nil, err
	}

	return api.AddTrackToPlaylist200JSONResponse{
		Data: struct {
			Message string `json:"message"`
		}{
			Message: "track added to playlist",
		},
	}, nil
}

func (h *PlaylistHandler) RemoveTrackFromPlaylist(
	ctx context.Context,
	request api.RemoveTrackFromPlaylistRequestObject,
) (api.RemoveTrackFromPlaylistResponseObject, error) {
	if err := h.useCase.RemoveTrackFromPlaylist(
		ctx,
		request.PlaylistId,
		request.TrackId,
	); err != nil {
		return nil, err
	}

	return api.RemoveTrackFromPlaylist204Response{}, nil
}

func toAPIPlaylist(playlist *domain.Playlist) api.Playlist {
	return api.Playlist{
		Id:     playlist.ID,
		Name:   playlist.Name,
		Tracks: playlist.Tracks,
	}
}
