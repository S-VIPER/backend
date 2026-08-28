package handler

import (
	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/api"
)

type Handler struct {
	*AuthHandler
	*PlaylistHandler
	*TrackHandler
}

var _ api.StrictServerInterface = (*Handler)(nil)

func NewHandler(
	authHandler *AuthHandler,
	playlistHandler *PlaylistHandler,
	trackHandler *TrackHandler,
) *Handler {
	return &Handler{
		AuthHandler:     authHandler,
		PlaylistHandler: playlistHandler,
		TrackHandler:    trackHandler,
	}
}
