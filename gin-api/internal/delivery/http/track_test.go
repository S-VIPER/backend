package http

import (
	"bytes"
	"encoding/json"
	"errors"
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

func (m *MockTrackUseCase) GetAllTracks() ([]*domain.Track, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Track), args.Error(1)
}

func TestCreateTrack(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockUseCase := new(MockTrackUseCase)
	handler := NewTrackHandler(mockUseCase)

	// Test data
	track := domain.Track{
		ID:          "track001",
		Title:       "Test Track",
		Artist:      "Test Artist",
		URL:         "http://example.com/track.mp3",
		AlbumTitle:  "Test Album",
		AlbumArtURL: "http://example.com/album.jpg",
		PreviewURL:  "http://example.com/preview.mp3",
		Genre:       []string{"Rock", "Alternative"},
		Year:        2023,
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
	assert.Equal(t, http.StatusCreated, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestGetTrackByID(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockUseCase := new(MockTrackUseCase)
	handler := NewTrackHandler(mockUseCase)

	// Test data
	trackID := "track001"
	expectedTrack := &domain.Track{
		ID:          trackID,
		Title:       "Test Track",
		Artist:      "Test Artist",
		URL:         "http://example.com/track.mp3",
		AlbumTitle:  "Test Album",
		AlbumArtURL: "http://example.com/album.jpg",
		PreviewURL:  "http://example.com/preview.mp3",
		Genre:       []string{"Rock", "Alternative"},
		Year:        2023,
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
	assert.Equal(t, expectedTrack.Artist, response.Artist)
	assert.Equal(t, expectedTrack.URL, response.URL)
	assert.Equal(t, expectedTrack.AlbumTitle, response.AlbumTitle)
}

func TestGetAllTracks(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockUseCase := new(MockTrackUseCase)
	handler := NewTrackHandler(mockUseCase)

	// Test data
	tracks := []*domain.Track{
		{
			ID:          "track001",
			Title:       "Test Track 1",
			Artist:      "Test Artist 1",
			URL:         "http://example.com/track1.mp3",
			AlbumTitle:  "Test Album 1",
			AlbumArtURL: "http://example.com/album1.jpg",
			PreviewURL:  "http://example.com/preview1.mp3",
			Genre:       []string{"Rock", "Alternative"},
			Year:        2023,
		},
		{
			ID:          "track002",
			Title:       "Test Track 2",
			Artist:      "Test Artist 2",
			URL:         "http://example.com/track2.mp3",
			AlbumTitle:  "Test Album 2",
			AlbumArtURL: "http://example.com/album2.jpg",
			PreviewURL:  "http://example.com/preview2.mp3",
			Genre:       []string{"Pop", "Electronic"},
			Year:        2022,
		},
	}

	// Expectations
	mockUseCase.On("GetAllTracks").Return(tracks, nil)

	// Create request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/tracks", nil)

	// Execute
	handler.GetAllTracks(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)

	var response struct {
		Tracks []*domain.Track `json:"tracks"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Tracks, 2)
	assert.Equal(t, tracks[0].ID, response.Tracks[0].ID)
	assert.Equal(t, tracks[1].ID, response.Tracks[1].ID)

	// Test case with error
	mockUseCase = new(MockTrackUseCase)
	handler = NewTrackHandler(mockUseCase)
	mockUseCase.On("GetAllTracks").Return(nil, errors.New("database error"))

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/tracks", nil)

	handler.GetAllTracks(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}
