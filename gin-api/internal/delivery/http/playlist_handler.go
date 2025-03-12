package http

import (
	"net/http"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type PlaylistHandler struct {
	useCase *usecase.PlaylistUseCase
}

func NewPlaylistHandler(useCase *usecase.PlaylistUseCase) *PlaylistHandler {
	return &PlaylistHandler{useCase: useCase}
}

func (h *PlaylistHandler) CreatePlaylist(c *gin.Context) {
	var playlist domain.Playlist
	if err := c.ShouldBindJSON(&playlist); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.CreatePlaylist(&playlist); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, playlist)
}

func (h *PlaylistHandler) GetPlaylistByID(c *gin.Context) {
	id := c.Param("id")
	playlist, err := h.useCase.GetPlaylistByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, playlist)
}

func (h *PlaylistHandler) UpdatePlaylist(c *gin.Context) {
	var playlist domain.Playlist
	if err := c.ShouldBindJSON(&playlist); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.UpdatePlaylist(&playlist); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, playlist)
}

func (h *PlaylistHandler) DeletePlaylist(c *gin.Context) {
	id := c.Param("id")
	if err := h.useCase.DeletePlaylist(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Playlist deleted"})
}

func (h *PlaylistHandler) AddTrackToPlaylist(c *gin.Context) {
	playlistID := c.Param("id")
	trackID := c.Param("trackId")

	if err := h.useCase.AddTrackToPlaylist(playlistID, trackID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Track added to playlist"})
}

func (h *PlaylistHandler) RemoveTrackFromPlaylist(c *gin.Context) {
	playlistID := c.Param("id")
	trackID := c.Param("trackId")

	if err := h.useCase.RemoveTrackFromPlaylist(playlistID, trackID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Track removed from playlist"})
}
