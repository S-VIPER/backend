package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/api"
	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/gin-gonic/gin"
)

func HandleRequestError(
	c *gin.Context,
	err error,
) {
	c.JSON(
		http.StatusBadRequest,
		newErrorResponse(
			api.INVALIDREQUEST,
			"invalid request",
		),
	)
}

func HandleHandlerError(
	c *gin.Context,
	err error,
) {
	status, code, message := mapDomainError(err)

	if status >= http.StatusInternalServerError {
		log.Printf("internal handler error: %v", err)
		message = "internal server error"
	}

	c.JSON(
		status,
		newErrorResponse(code, message),
	)
}

func HandleResponseError(
	c *gin.Context,
	err error,
) {
	log.Printf("response serialization error: %v", err)

	c.JSON(
		http.StatusInternalServerError,
		newErrorResponse(
			api.INTERNALERROR,
			"internal server error",
		),
	)
}

func mapDomainError(
	err error,
) (int, api.ErrorResponseErrorCode, string) {
	switch {
	case errors.Is(err, domain.ErrInvalidPlaylist):
		return http.StatusBadRequest, api.INVALIDPLAYLIST, err.Error()

	case errors.Is(err, domain.ErrInvalidPlaylistID):
		return http.StatusBadRequest, api.INVALIDPLAYLISTID, err.Error()

	case errors.Is(err, domain.ErrInvalidPlaylistName):
		return http.StatusBadRequest, api.INVALIDPLAYLISTNAME, err.Error()

	case errors.Is(err, domain.ErrInvalidTrack):
		return http.StatusBadRequest, api.INVALIDTRACK, err.Error()

	case errors.Is(err, domain.ErrInvalidTrackID):
		return http.StatusBadRequest, api.INVALIDTRACKID, err.Error()

	case errors.Is(err, domain.ErrInvalidTrackTitle):
		return http.StatusBadRequest, api.INVALIDTRACKTITLE, err.Error()

	case errors.Is(err, domain.ErrInvalidTrackArtist):
		return http.StatusBadRequest, api.INVALIDTRACKARTIST, err.Error()

	case errors.Is(err, domain.ErrInvalidTrackURL):
		return http.StatusBadRequest, api.INVALIDTRACKURL, err.Error()

	case errors.Is(err, domain.ErrInvalidTrackYear):
		return http.StatusBadRequest, api.INVALIDTRACKYEAR, err.Error()

	case errors.Is(err, domain.ErrPlaylistNotFound):
		return http.StatusNotFound, api.PLAYLISTNOTFOUND, err.Error()

	case errors.Is(err, domain.ErrTrackNotFound):
		return http.StatusNotFound, api.TRACKNOTFOUND, err.Error()

	case errors.Is(err, domain.ErrPlaylistAlreadyExists):
		return http.StatusConflict, api.PLAYLISTALREADYEXISTS, err.Error()

	case errors.Is(err, domain.ErrTrackAlreadyExists):
		return http.StatusConflict, api.TRACKALREADYEXISTS, err.Error()

	default:
		return http.StatusInternalServerError, api.INTERNALERROR, "internal server error"
	}
}

func newErrorResponse(
	code api.ErrorResponseErrorCode,
	message string,
) api.ErrorResponse {
	response := api.ErrorResponse{}
	response.Error.Code = code
	response.Error.Message = message

	return response
}
