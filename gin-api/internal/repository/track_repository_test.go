package repository

import (
	"context"
	"testing"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// TrackRepositoryTestSuite - тестовый набор для репозитория треков
type TrackRepositoryTestSuite struct {
	suite.Suite
	db         *mongo.Database
	client     *mongo.Client
	repository *TrackRepository
}

// SetupSuite - подготовка тестового окружения перед запуском всех тестов
func (suite *TrackRepositoryTestSuite) SetupSuite() {
	// Подключение к MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	ctx := context.Background()

	// Попытка подключения к MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		suite.T().Skip("MongoDB is not available, skipping suite")
		return
	}

	// Проверка соединения
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		suite.T().Skip("MongoDB is not available, skipping suite")
		return
	}

	// Создание тестовой базы данных
	db := client.Database("test_db")

	suite.client = client
	suite.db = db
	suite.repository = NewTrackRepository(db)
}

// TearDownSuite - очистка после выполнения всех тестов
func (suite *TrackRepositoryTestSuite) TearDownSuite() {
	if suite.client != nil {
		suite.db.Drop(context.Background())
		suite.client.Disconnect(context.Background())
	}
}

// SetupTest - подготовка перед каждым тестом
func (suite *TrackRepositoryTestSuite) SetupTest() {
	// Очистка коллекции перед каждым тестом
	if suite.db != nil {
		suite.db.Collection("tracks").Drop(context.Background())
	}
}

// TestCreate - тест создания трека
func (suite *TrackRepositoryTestSuite) TestCreate() {
	if suite.repository == nil {
		suite.T().Skip("Repository is not initialized, skipping test")
		return
	}

	// Тестовые данные
	track := &domain.Track{
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

	// Создание трека
	err := suite.repository.Create(track)
	suite.NoError(err)

	// Проверка, что трек создан
	var result domain.Track
	err = suite.db.Collection("tracks").FindOne(context.Background(), bson.M{"_id": track.ID}).Decode(&result)
	suite.NoError(err)
	suite.Equal(track.ID, result.ID)
	suite.Equal(track.Title, result.Title)
	suite.Equal(track.Artist, result.Artist)
}

// TestGetByID - тест получения трека по ID
func (suite *TrackRepositoryTestSuite) TestGetByID() {
	if suite.repository == nil {
		suite.T().Skip("Repository is not initialized, skipping test")
		return
	}

	// Вставка тестовых данных
	track := &domain.Track{
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
	_, err := suite.db.Collection("tracks").InsertOne(context.Background(), track)
	suite.NoError(err)

	// Получение трека по ID
	result, err := suite.repository.GetByID(track.ID)
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(track.ID, result.ID)
	suite.Equal(track.Title, result.Title)
	suite.Equal(track.Artist, result.Artist)

	// Получение несуществующего трека
	result, err = suite.repository.GetByID("nonexistent")
	suite.Error(err)
	suite.Nil(result)
}

// TestUpdate - тест обновления трека
func (suite *TrackRepositoryTestSuite) TestUpdate() {
	if suite.repository == nil {
		suite.T().Skip("Repository is not initialized, skipping test")
		return
	}

	// Вставка тестовых данных
	track := &domain.Track{
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
	_, err := suite.db.Collection("tracks").InsertOne(context.Background(), track)
	suite.NoError(err)

	// Обновление трека
	updatedTrack := &domain.Track{
		ID:          "track001",
		Title:       "Updated Track",
		Artist:      "Test Artist",
		URL:         "http://example.com/track.mp3",
		AlbumTitle:  "Test Album",
		AlbumArtURL: "http://example.com/album.jpg",
		PreviewURL:  "http://example.com/preview.mp3",
		Genre:       []string{"Rock", "Alternative"},
		Year:        2024,
	}
	err = suite.repository.Update(updatedTrack)
	suite.NoError(err)

	// Проверка, что трек обновлен
	var result domain.Track
	err = suite.db.Collection("tracks").FindOne(context.Background(), bson.M{"_id": track.ID}).Decode(&result)
	suite.NoError(err)
	suite.Equal(updatedTrack.Title, result.Title)
	suite.Equal(updatedTrack.Year, result.Year)
}

// TestDelete - тест удаления трека
func (suite *TrackRepositoryTestSuite) TestDelete() {
	if suite.repository == nil {
		suite.T().Skip("Repository is not initialized, skipping test")
		return
	}

	// Вставка тестовых данных
	track := &domain.Track{
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
	_, err := suite.db.Collection("tracks").InsertOne(context.Background(), track)
	suite.NoError(err)

	// Удаление трека
	err = suite.repository.Delete(track.ID)
	suite.NoError(err)

	// Проверка, что трек удален
	count, err := suite.db.Collection("tracks").CountDocuments(context.Background(), bson.M{"_id": track.ID})
	suite.NoError(err)
	suite.Equal(int64(0), count)
}

// TestGetAllTracks - тест получения всех треков
func (suite *TrackRepositoryTestSuite) TestGetAllTracks() {
	if suite.repository == nil {
		suite.T().Skip("Repository is not initialized, skipping test")
		return
	}

	// Вставка тестовых данных
	tracks := []interface{}{
		&domain.Track{
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
		&domain.Track{
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
	_, err := suite.db.Collection("tracks").InsertMany(context.Background(), tracks)
	suite.NoError(err)

	// Получение всех треков
	result, err := suite.repository.GetAllTracks()
	suite.NoError(err)
	suite.NotNil(result)
	suite.Len(result, 2)
}

// TestTrackRepository запускает набор тестов
func TestTrackRepository(t *testing.T) {
	suite.Run(t, new(TrackRepositoryTestSuite))
}
