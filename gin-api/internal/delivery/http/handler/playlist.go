package handler

import (
	"context"

	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/api"
	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/middleware"
	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"
	"github.com/google/uuid"
)

type PlaylistHandler struct {
	useCase *usecase.PlaylistUseCase
}

func NewPlaylistHandler(useCase *usecase.PlaylistUseCase) *PlaylistHandler {
	return &PlaylistHandler{useCase: useCase}
}

func (h *PlaylistHandler) CreatePlaylist(
	ctx context.Context,
	request api.CreatePlaylistRequestObject,
) (api.CreatePlaylistResponseObject, error) {
	ownerID, ok := middleware.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	playlist := &domain.Playlist{
		Name:       request.Body.Name,
		Visibility: domain.PlaylistVisibility(*request.Body.Visibility),
	}

	if err := h.useCase.CreatePlaylist(ctx, ownerID, playlist); err != nil {
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
	var viewerID *uuid.UUID
	if id, ok := middleware.UserID(ctx); ok {
		viewerID = &id
	}

	playlist, err := h.useCase.GetPlaylistByID(
		ctx,
		viewerID,
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
	ownerID, ok := middleware.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	playlist := &domain.Playlist{
		ID:         request.PlaylistId,
		Name:       request.Body.Name,
		Visibility: domain.PlaylistVisibility(*request.Body.Visibility),
	}

	if err := h.useCase.UpdatePlaylist(ctx, ownerID, playlist); err != nil {
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
	ownerID, ok := middleware.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	if err := h.useCase.DeletePlaylist(
		ctx,
		ownerID,
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
	ownerID, ok := middleware.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	if err := h.useCase.AddTrackToPlaylist(
		ctx,
		ownerID,
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
	ownerID, ok := middleware.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	if err := h.useCase.RemoveTrackFromPlaylist(
		ctx,
		ownerID,
		request.PlaylistId,
		request.TrackId,
	); err != nil {
		return nil, err
	}

	return api.RemoveTrackFromPlaylist204Response{}, nil
}

func toAPIPlaylist(playlist *domain.Playlist) api.Playlist {
	visibility := api.PlaylistVisibility(playlist.Visibility)
	return api.Playlist{
		Id:         playlist.ID,
		Name:       playlist.Name,
		Tracks:     playlist.Tracks,
		Visibility: &visibility,
	}
}
