package http

import (
	"net/http"

	"github.com/S-VIPER/backend/gin-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type TrackHandler struct {
	useCase TrackUseCaseInterface
}

func NewTrackHandler(useCase TrackUseCaseInterface) *TrackHandler {
	return &TrackHandler{useCase: useCase}
}

func (h *TrackHandler) CreateTrack(c *gin.Context) {
	var track domain.Track
	if err := c.ShouldBindJSON(&track); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.CreateTrack(&track); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, track)
}

func (h *TrackHandler) GetTrackByID(c *gin.Context) {
	id := c.Param("id")
	track, err := h.useCase.GetTrackByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, track)
}

func (h *TrackHandler) UpdateTrack(c *gin.Context) {
	var track domain.Track
	if err := c.ShouldBindJSON(&track); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.UpdateTrack(&track); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, track)
}

func (h *TrackHandler) DeleteTrack(c *gin.Context) {
	id := c.Param("id")
	if err := h.useCase.DeleteTrack(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Track deleted"})
}
