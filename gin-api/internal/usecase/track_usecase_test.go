package usecase

import (
	"errors"
	"testing"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTrackRepository - мок для репозитория
type MockTrackRepository struct {
	mock.Mock
}

// Проверка, что MockTrackRepository реализует интерфейс TrackRepositoryInterface
var _ repository.TrackRepositoryInterface = (*MockTrackRepository)(nil)

func (m *MockTrackRepository) Create(track *domain.Track) error {
	args := m.Called(track)
	return args.Error(0)
}

func (m *MockTrackRepository) GetByID(id string) (*domain.Track, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Track), args.Error(1)
}

func (m *MockTrackRepository) Update(track *domain.Track) error {
	args := m.Called(track)
	return args.Error(0)
}

func (m *MockTrackRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTrackRepository) GetAllTracks() ([]*domain.Track, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Track), args.Error(1)
}

func TestCreateTrack(t *testing.T) {
	mockRepo := new(MockTrackRepository)
	uc := NewTrackUseCase(mockRepo)

	track := &domain.Track{
		ID:     "track001",
		Title:  "Test Track",
		Artist: "Test Artist",
	}

	// Тест успешного создания
	mockRepo.On("Create", track).Return(nil)
	err := uc.CreateTrack(track)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Тест с ошибкой
	mockRepo = new(MockTrackRepository)
	uc = NewTrackUseCase(mockRepo)
	expectedErr := errors.New("database error")
	mockRepo.On("Create", track).Return(expectedErr)
	err = uc.CreateTrack(track)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestGetTrackByID(t *testing.T) {
	mockRepo := new(MockTrackRepository)
	uc := NewTrackUseCase(mockRepo)

	track := &domain.Track{
		ID:     "track001",
		Title:  "Test Track",
		Artist: "Test Artist",
	}

	// Тест успешного получения
	mockRepo.On("GetByID", "track001").Return(track, nil)
	result, err := uc.GetTrackByID("track001")
	assert.NoError(t, err)
	assert.Equal(t, track, result)
	mockRepo.AssertExpectations(t)

	// Тест с ошибкой "не найдено"
	mockRepo = new(MockTrackRepository)
	uc = NewTrackUseCase(mockRepo)
	expectedErr := errors.New("track not found")
	mockRepo.On("GetByID", "nonexistent").Return(nil, expectedErr)
	result, err = uc.GetTrackByID("nonexistent")
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestUpdateTrack(t *testing.T) {
	mockRepo := new(MockTrackRepository)
	uc := NewTrackUseCase(mockRepo)

	track := &domain.Track{
		ID:     "track001",
		Title:  "Updated Track",
		Artist: "Test Artist",
	}

	// Тест успешного обновления
	mockRepo.On("Update", track).Return(nil)
	err := uc.UpdateTrack(track)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Тест с ошибкой
	mockRepo = new(MockTrackRepository)
	uc = NewTrackUseCase(mockRepo)
	expectedErr := errors.New("update error")
	mockRepo.On("Update", track).Return(expectedErr)
	err = uc.UpdateTrack(track)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteTrack(t *testing.T) {
	mockRepo := new(MockTrackRepository)
	uc := NewTrackUseCase(mockRepo)

	// Тест успешного удаления
	mockRepo.On("Delete", "track001").Return(nil)
	err := uc.DeleteTrack("track001")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Тест с ошибкой
	mockRepo = new(MockTrackRepository)
	uc = NewTrackUseCase(mockRepo)
	expectedErr := errors.New("delete error")
	mockRepo.On("Delete", "track001").Return(expectedErr)
	err = uc.DeleteTrack("track001")
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestGetAllTracks(t *testing.T) {
	mockRepo := new(MockTrackRepository)
	uc := NewTrackUseCase(mockRepo)

	tracks := []*domain.Track{
		{
			ID:     "track001",
			Title:  "Test Track 1",
			Artist: "Test Artist 1",
		},
		{
			ID:     "track002",
			Title:  "Test Track 2",
			Artist: "Test Artist 2",
		},
	}

	// Тест успешного получения всех треков
	mockRepo.On("GetAllTracks").Return(tracks, nil)
	result, err := uc.GetAllTracks()
	assert.NoError(t, err)
	assert.Equal(t, tracks, result)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)

	// Тест с ошибкой
	mockRepo = new(MockTrackRepository)
	uc = NewTrackUseCase(mockRepo)
	expectedErr := errors.New("database error")
	mockRepo.On("GetAllTracks").Return(nil, expectedErr)
	result, err = uc.GetAllTracks()
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}
