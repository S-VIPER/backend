package handler

import (
	"context"

	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/api"
	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"
)

type TrackHandler struct {
	useCase *usecase.TrackUseCase
}

func NewTrackHandler(
	useCase *usecase.TrackUseCase,
) *TrackHandler {
	return &TrackHandler{
		useCase: useCase,
	}
}

func (h *TrackHandler) CreateTrack(
	ctx context.Context,
	request api.CreateTrackRequestObject,
) (api.CreateTrackResponseObject, error) {
	req := request.Body

	track := &domain.Track{
		ID:          req.Id,
		Title:       req.Title,
		Artist:      req.Artist,
		URL:         req.Url,
		AlbumTitle:  req.AlbumTitle,
		AlbumArtURL: req.AlbumArtURL,
		PreviewURL:  req.PreviewURL,
		Genre:       req.Genre,
		Year:        req.Year,
	}

	if err := h.useCase.CreateTrack(ctx, track); err != nil {
		return nil, err
	}

	return api.CreateTrack201JSONResponse{
		Data: toAPITrack(track),
	}, nil
}

func (h *TrackHandler) GetAllTracks(
	ctx context.Context,
	request api.GetAllTracksRequestObject,
) (api.GetAllTracksResponseObject, error) {
	tracks, err := h.useCase.GetAllTracks(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]api.Track, 0, len(tracks))

	for _, track := range tracks {
		result = append(result, toAPITrack(track))
	}

	return api.GetAllTracks200JSONResponse{
		Data: result,
	}, nil
}

func (h *TrackHandler) GetTrackByID(
	ctx context.Context,
	request api.GetTrackByIDRequestObject,
) (api.GetTrackByIDResponseObject, error) {
	track, err := h.useCase.GetTrackByID(
		ctx,
		request.TrackId,
	)
	if err != nil {
		return nil, err
	}

	return api.GetTrackByID200JSONResponse{
		Data: toAPITrack(track),
	}, nil
}

func (h *TrackHandler) UpdateTrack(
	ctx context.Context,
	request api.UpdateTrackRequestObject,
) (api.UpdateTrackResponseObject, error) {
	req := request.Body

	track := &domain.Track{
		ID:          request.TrackId,
		Title:       req.Title,
		Artist:      req.Artist,
		URL:         req.Url,
		AlbumTitle:  req.AlbumTitle,
		AlbumArtURL: req.AlbumArtURL,
		PreviewURL:  req.PreviewURL,
		Genre:       req.Genre,
		Year:        req.Year,
	}

	if err := h.useCase.UpdateTrack(ctx, track); err != nil {
		return nil, err
	}

	return api.UpdateTrack200JSONResponse{
		Data: toAPITrack(track),
	}, nil
}

func (h *TrackHandler) DeleteTrack(
	ctx context.Context,
	request api.DeleteTrackRequestObject,
) (api.DeleteTrackResponseObject, error) {
	if err := h.useCase.DeleteTrack(
		ctx,
		request.TrackId,
	); err != nil {
		return nil, err
	}

	return api.DeleteTrack204Response{}, nil
}

func toAPITrack(track *domain.Track) api.Track {
	return api.Track{
		AlbumArtURL: track.AlbumArtURL,
		AlbumTitle:  track.AlbumTitle,
		Artist:      track.Artist,
		Genre:       track.Genre,
		Id:          track.ID,
		PreviewURL:  track.PreviewURL,
		Title:       track.Title,
		Url:         track.URL,
		Year:        track.Year,
	}
}
