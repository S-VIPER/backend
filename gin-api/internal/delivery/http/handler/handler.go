package handler

import (
	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/api"
)

type Handler struct {
	*PlaylistHandler
	*TrackHandler
}

var _ api.StrictServerInterface = (*Handler)(nil)

func NewHandler(
	playlistHandler *PlaylistHandler,
	trackHandler *TrackHandler,
) *Handler {
	return &Handler{
		PlaylistHandler: playlistHandler,
		TrackHandler:    trackHandler,
	}
}
