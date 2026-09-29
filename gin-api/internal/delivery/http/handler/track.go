package handler

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/api"
	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type TrackHandler struct {
	useCase       *usecase.TrackUseCase
	maxUploadSize int64
}

func NewTrackHandler(useCase *usecase.TrackUseCase) *TrackHandler {
	return &TrackHandler{
		useCase:       useCase,
		maxUploadSize: usecase.DefaultTrackMaxUploadSize,
	}
}

func (h *TrackHandler) WithMaxUploadSize(size int64) *TrackHandler {
	if size > 0 {
		h.maxUploadSize = size
	}
	return h
}

func (h *TrackHandler) UploadTrack(
	ctx context.Context,
	request api.UploadTrackRequestObject,
) (api.UploadTrackResponseObject, error) {
	if request.Body == nil {
		return nil, domain.ErrInvalidTrack
	}

	upload, cleanup, err := parseTrackMultipart(request.Body, h.maxUploadSize)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	track, err := h.useCase.UploadTrack(ctx, upload)
	if err != nil {
		return nil, err
	}

	apiTrack, err := toAPITrack(track)
	if err != nil {
		return nil, err
	}

	return api.UploadTrack201JSONResponse{
		Data: apiTrack,
	}, nil
}

func (h *TrackHandler) SearchTrackMetadata(
	ctx context.Context,
	request api.SearchTrackMetadataRequestObject,
) (api.SearchTrackMetadataResponseObject, error) {
	limit := 10
	if request.Params.Limit != nil {
		limit = int(*request.Params.Limit)
	}

	candidates, err := h.useCase.SearchTrackMetadata(
		ctx,
		request.Params.Artist,
		request.Params.Title,
		limit,
	)
	if err != nil {
		return nil, err
	}

	response := api.SearchTrackMetadata200JSONResponse{}
	response.Data = make([]api.TrackMetadataCandidate, 0, len(candidates))

	for _, candidate := range candidates {
		mbid, err := parseRequiredUUID(candidate.MBID)
		if err != nil {
			return nil, fmt.Errorf("invalid MusicBrainz ID %q: %w", candidate.MBID, err)
		}

		item := api.TrackMetadataCandidate{
			MusicbrainzId: mbid,
			Title:         candidate.Title,
			Artist:        candidate.Artist,
			AlbumTitle:    optionalString(candidate.AlbumTitle),
			ReleaseId:     optionalUUID(candidate.MusicBrainzReleaseID),
			ReleaseGroupId: optionalUUID(
				candidate.MusicBrainzReleaseGroupID,
			),
			Score: int32(candidate.Score),
		}

		response.Data = append(response.Data, item)
	}

	return response, nil
}

func (h *TrackHandler) GetTrackContent(
	ctx context.Context,
	request api.GetTrackContentRequestObject,
) (api.GetTrackContentResponseObject, error) {
	trackID := trackIDToString(request.TrackId)
	access, err := h.useCase.GetTrackContent(ctx, trackID)
	if err != nil {
		return nil, err
	}

	response := api.GetTrackContent200JSONResponse{}
	response.Data.Url = access.URL
	response.Data.ExpiresIn = int32(access.ExpiresIn.Seconds())

	return response, nil
}

func parseTrackMultipart(
	reader *multipart.Reader,
	maxSize int64,
) (usecase.TrackUpload, func(), error) {
	tempFile, err := os.CreateTemp("", "sviper-track-*")
	if err != nil {
		return usecase.TrackUpload{}, func() {}, fmt.Errorf("create temporary upload file: %w", err)
	}

	cleanup := func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}

	var (
		musicBrainzID string
		fileName      string
		contentType   string
		fileSize      int64
		fileSeen      bool
	)

	for {
		part, nextErr := reader.NextPart()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			cleanup()
			return usecase.TrackUpload{}, func() {}, fmt.Errorf("read multipart body: %w", nextErr)
		}

		switch part.FormName() {
		case "musicbrainz_id":
			value, err := io.ReadAll(io.LimitReader(part, 1024))
			if err != nil {
				cleanup()
				return usecase.TrackUpload{}, func() {}, fmt.Errorf("read musicbrainz_id: %w", err)
			}
			musicBrainzID = strings.TrimSpace(string(value))

		case "file":
			if fileSeen {
				cleanup()
				return usecase.TrackUpload{}, func() {}, domain.ErrInvalidTrack
			}
			fileSeen = true
			fileName = filepath.Base(part.FileName())
			contentType = part.Header.Get("Content-Type")

			written, err := io.Copy(
				tempFile,
				io.LimitReader(part, maxSize+1),
			)
			if err != nil {
				cleanup()
				return usecase.TrackUpload{}, func() {}, fmt.Errorf("store temporary upload: %w", err)
			}
			fileSize = written
			if fileSize > maxSize {
				cleanup()
				return usecase.TrackUpload{}, func() {}, domain.ErrInvalidTrack
			}
		}
	}

	if musicBrainzID == "" || !fileSeen || fileSize == 0 {
		cleanup()
		return usecase.TrackUpload{}, func() {}, domain.ErrInvalidTrack
	}

	if _, err := tempFile.Seek(0, io.SeekStart); err != nil {
		cleanup()
		return usecase.TrackUpload{}, func() {}, fmt.Errorf("rewind temporary upload: %w", err)
	}

	return usecase.TrackUpload{
		MusicBrainzID: musicBrainzID,
		Reader:        tempFile,
		FileName:      fileName,
		ContentType:   contentType,
		Size:          fileSize,
	}, cleanup, nil
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
		apiTrack, err := toAPITrack(track)
		if err != nil {
			return nil, err
		}
		result = append(result, apiTrack)
	}

	return api.GetAllTracks200JSONResponse{
		Data: result,
	}, nil
}

func (h *TrackHandler) GetTrackByID(
	ctx context.Context,
	request api.GetTrackByIDRequestObject,
) (api.GetTrackByIDResponseObject, error) {
	trackID := trackIDToString(request.TrackId)
	track, err := h.useCase.GetTrackByID(ctx, trackID)
	if err != nil {
		return nil, err
	}

	apiTrack, err := toAPITrack(track)
	if err != nil {
		return nil, err
	}

	return api.GetTrackByID200JSONResponse{
		Data: apiTrack,
	}, nil
}

func (h *TrackHandler) UpdateTrack(
	ctx context.Context,
	request api.UpdateTrackRequestObject,
) (api.UpdateTrackResponseObject, error) {
	req := request.Body

	track := &domain.Track{
		ID:          trackIDToString(request.TrackId),
		Title:       req.Title,
		Artist:      req.Artist,
		AlbumTitle:  req.AlbumTitle,
		AlbumArtURL: req.AlbumArtURL,
		Genre:       req.Genre,
		Year:        req.Year,
	}

	if err := h.useCase.UpdateTrack(ctx, track); err != nil {
		return nil, err
	}

	apiTrack, err := toAPITrack(track)
	if err != nil {
		return nil, err
	}

	return api.UpdateTrack200JSONResponse{
		Data: apiTrack,
	}, nil
}

func (h *TrackHandler) DeleteTrack(
	ctx context.Context,
	request api.DeleteTrackRequestObject,
) (api.DeleteTrackResponseObject, error) {
	trackID := trackIDToString(request.TrackId)
	if err := h.useCase.DeleteTrack(ctx, trackID); err != nil {
		return nil, err
	}

	return api.DeleteTrack204Response{}, nil
}

func toAPITrack(track *domain.Track) (api.Track, error) {
	id, err := parseRequiredUUID(track.ID)
	if err != nil {
		return api.Track{}, fmt.Errorf("invalid track ID %q: %w", track.ID, err)
	}

	return api.Track{
		Id:          id,
		Title:       track.Title,
		Artist:      track.Artist,
		AlbumTitle:  track.AlbumTitle,
		AlbumArtURL: optionalString(track.AlbumArtURL),
		Genre:       track.Genre,
		Year:        optionalInt(track.Year),
	}, nil
}

func trackIDToString(id api.TrackId) string {
	return uuid.UUID(id).String()
}

func parseRequiredUUID(value string) (openapi_types.UUID, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return openapi_types.UUID{}, err
	}
	return openapi_types.UUID(parsed), nil
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func optionalInt(value int) *int {
	if value == 0 {
		return nil
	}
	return &value
}

func optionalUUID(value string) *openapi_types.UUID {
	result, err := parseRequiredUUID(value)
	if err != nil {
		return nil
	}
	return &result
}
