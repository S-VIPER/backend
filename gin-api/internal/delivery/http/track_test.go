package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTrackUseCase is a mock implementation of TrackUseCase
type MockTrackUseCase struct {
	mock.Mock
}

func (m *MockTrackUseCase) CreateTrack(track *domain.Track) error {
	args := m.Called(track)
	return args.Error(0)
}

func (m *MockTrackUseCase) GetTrackByID(id string) (*domain.Track, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Track), args.Error(1)
}

func (m *MockTrackUseCase) UpdateTrack(track *domain.Track) error {
	args := m.Called(track)
	return args.Error(0)
}

func (m *MockTrackUseCase) DeleteTrack(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestCreateTrack(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockUseCase := new(MockTrackUseCase)
	handler := NewTrackHandler(mockUseCase)

	// Test data
	track := domain.Track{
		Title:    "Test Track",
		BPM:      120,
		Key:      "C",
		FilePath: "http://example.com/track.mp3",
	}
	jsonData, _ := json.Marshal(track)

	// Expectations
	mockUseCase.On("CreateTrack", mock.AnythingOfType("*domain.Track")).Return(nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/tracks", bytes.NewBuffer(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	handler.CreateTrack(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestGetTrackByID(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockUseCase := new(MockTrackUseCase)
	handler := NewTrackHandler(mockUseCase)

	// Test data
	trackID := "123"
	expectedTrack := &domain.Track{
		ID:       123,
		Title:    "Test Track",
		BPM:      120,
		Key:      "C",
		FilePath: "http://example.com/track.mp3",
	}

	// Expectations
	mockUseCase.On("GetTrackByID", trackID).Return(expectedTrack, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/tracks/"+trackID, nil)
	c.Params = []gin.Param{{Key: "id", Value: trackID}}

	// Execute
	handler.GetTrackByID(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)

	var response domain.Track
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTrack.ID, response.ID)
	assert.Equal(t, expectedTrack.Title, response.Title)
	assert.Equal(t, expectedTrack.BPM, response.BPM)
	assert.Equal(t, expectedTrack.Key, response.Key)
	assert.Equal(t, expectedTrack.FilePath, response.FilePath)
}
