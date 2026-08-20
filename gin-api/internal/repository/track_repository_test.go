package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TrackRepositoryTestSuite struct {
	suite.Suite

	ctx       context.Context
	container *mongodb.MongoDBContainer
	client    *mongo.Client
	db        *mongo.Database

	repository *TrackRepository
}

func (s *TrackRepositoryTestSuite) SetupSuite() {
	s.ctx = context.Background()

	container, err := mongodb.Run(
		s.ctx,
		"mongo:8",
	)
	require.NoError(s.T(), err)

	s.container = container

	connectionString, err := container.ConnectionString(s.ctx)
	require.NoError(s.T(), err)

	client, err := mongo.Connect(
		s.ctx,
		options.Client().ApplyURI(connectionString),
	)
	require.NoError(s.T(), err)

	err = client.Ping(s.ctx, nil)
	require.NoError(s.T(), err)

	s.client = client
	s.db = client.Database("track_repository_test")

	s.repository = NewTrackRepository(s.db)

	// Compile-time protection for interface implementation.
	var _ TrackRepositoryInterface = (*TrackRepository)(nil)
}

func (s *TrackRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		err := s.db.Drop(s.ctx)
		require.NoError(s.T(), err)
	}

	if s.client != nil {
		err := s.client.Disconnect(s.ctx)
		require.NoError(s.T(), err)
	}

	if s.container != nil {
		err := testcontainers.TerminateContainer(s.container)
		require.NoError(s.T(), err)
	}
}

func (s *TrackRepositoryTestSuite) SetupTest() {
	err := s.db.Collection("tracks").Drop(s.ctx)
	require.NoError(s.T(), err)
}

func testTrack(id string) *domain.Track {
	return &domain.Track{
		ID:          id,
		Title:       "Test Track",
		Artist:      "Test Artist",
		URL:         "https://example.com/track.mp3",
		AlbumTitle:  "Test Album",
		AlbumArtURL: "https://example.com/album.jpg",
		PreviewURL:  "https://example.com/preview.mp3",
		Genre:       []string{"Rock", "Alternative"},
		Year:        2025,
	}
}

// ------------------------------------------------------------
// Create
// ------------------------------------------------------------

func (s *TrackRepositoryTestSuite) TestCreate() {
	ctx := context.Background()
	track := testTrack("track-001")

	err := s.repository.Create(ctx, track)

	s.Require().NoError(err)

	result, err := s.repository.GetByID(ctx, track.ID)

	s.Require().NoError(err)
	s.Require().NotNil(result)

	s.Equal(track.ID, result.ID)
	s.Equal(track.Title, result.Title)
	s.Equal(track.Artist, result.Artist)
	s.Equal(track.URL, result.URL)
	s.Equal(track.AlbumTitle, result.AlbumTitle)
	s.Equal(track.AlbumArtURL, result.AlbumArtURL)
	s.Equal(track.PreviewURL, result.PreviewURL)
	s.Equal(track.Genre, result.Genre)
	s.Equal(track.Year, result.Year)
}

func (s *TrackRepositoryTestSuite) TestCreateDuplicateID() {
	ctx := context.Background()
	track := testTrack("track-001")

	err := s.repository.Create(ctx, track)
	s.Require().NoError(err)

	err = s.repository.Create(ctx, track)

	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrTrackAlreadyExists)
}

// ------------------------------------------------------------
// GetByID
// ------------------------------------------------------------

func insertTrackDocument(
	ctx context.Context,
	collection *mongo.Collection,
	track *domain.Track,
) error {
	document := trackDocumentFromDomain(track)

	_, err := collection.InsertOne(ctx, document)

	return err
}
func (s *TrackRepositoryTestSuite) TestGetByID() {
	ctx := context.Background()
	track := testTrack("track-001")

	err := insertTrackDocument(
		ctx,
		s.db.Collection("tracks"),
		track,
	)
	s.Require().NoError(err)

	result, err := s.repository.GetByID(
		ctx,
		track.ID,
	)

	s.Require().NoError(err)
	s.Require().NotNil(result)

	s.Equal(track.ID, result.ID)
	s.Equal(track.Title, result.Title)
	s.Equal(track.Artist, result.Artist)
}

func (s *TrackRepositoryTestSuite) TestGetByIDNotFound() {
	ctx := context.Background()

	result, err := s.repository.GetByID(ctx, "does-not-exist")

	s.Require().Error(err)
	s.Require().Nil(result)
	s.ErrorIs(err, domain.ErrTrackNotFound)
}

// ------------------------------------------------------------
// Exists
// ------------------------------------------------------------

func (s *TrackRepositoryTestSuite) TestExistsWhenTrackExists() {
	ctx := context.Background()
	track := testTrack("track-001")

	err := insertTrackDocument(
		ctx,
		s.db.Collection("tracks"),
		track,
	)
	s.Require().NoError(err)

	exists, err := s.repository.Exists(ctx, track.ID)

	s.Require().NoError(err)
	s.True(exists)
}

func (s *TrackRepositoryTestSuite) TestExistsWhenTrackDoesNotExist() {
	ctx := context.Background()

	exists, err := s.repository.Exists(ctx, "does-not-exist")

	s.Require().NoError(err)
	s.False(exists)
}

// ------------------------------------------------------------
// Update
// ------------------------------------------------------------

func (s *TrackRepositoryTestSuite) TestUpdate() {
	ctx := context.Background()

	track := testTrack("track-001")

	err := insertTrackDocument(
		ctx,
		s.db.Collection("tracks"),
		track,
	)
	s.Require().NoError(err)

	updated := testTrack("track-001")
	updated.Title = "Updated Track"
	updated.Artist = "Updated Artist"
	updated.Year = 2026
	updated.Genre = []string{"Electronic"}

	err = s.repository.Update(ctx, updated)

	s.Require().NoError(err)

	result, err := s.repository.GetByID(ctx, track.ID)

	s.Require().NoError(err)
	s.Require().NotNil(result)

	s.Equal("Updated Track", result.Title)
	s.Equal("Updated Artist", result.Artist)
	s.Equal(2026, result.Year)
	s.Equal([]string{"Electronic"}, result.Genre)
	s.Equal(track.ID, result.ID)
}

func (s *TrackRepositoryTestSuite) TestUpdateNotFound() {
	ctx := context.Background()

	track := testTrack("does-not-exist")

	err := s.repository.Update(ctx, track)

	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrTrackNotFound)
}

// ------------------------------------------------------------
// Delete
// ------------------------------------------------------------

func (s *TrackRepositoryTestSuite) TestDelete() {
	ctx := context.Background()

	track := testTrack("track-001")

	err := insertTrackDocument(
		ctx,
		s.db.Collection("tracks"),
		track,
	)
	s.Require().NoError(err)

	err = s.repository.Delete(ctx, track.ID)

	s.Require().NoError(err)

	result, err := s.repository.GetByID(
		ctx,
		track.ID,
	)

	s.Require().Error(err)
	s.Nil(result)
	s.ErrorIs(err, domain.ErrTrackNotFound)
}

func (s *TrackRepositoryTestSuite) TestDeleteNotFound() {
	ctx := context.Background()

	err := s.repository.Delete(ctx, "does-not-exist")

	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrTrackNotFound)
}

// ------------------------------------------------------------
// GetAllTracks
// ------------------------------------------------------------

func (s *TrackRepositoryTestSuite) TestGetAllTracksEmpty() {
	ctx := context.Background()

	result, err := s.repository.GetAllTracks(ctx)

	s.Require().NoError(err)
	s.Empty(result)
}

func (s *TrackRepositoryTestSuite) TestGetAllTracks() {
	ctx := context.Background()

	tracks := []*domain.Track{
		testTrack("track-001"),
		testTrack("track-002"),
		testTrack("track-003"),
	}

	for _, track := range tracks {
		err := insertTrackDocument(
			ctx,
			s.db.Collection("tracks"),
			track,
		)
		s.Require().NoError(err)
	}

	result, err := s.repository.GetAllTracks(ctx)

	s.Require().NoError(err)
	s.Require().Len(result, 3)

	ids := make(map[string]bool, len(result))

	for _, track := range result {
		ids[track.ID] = true
	}

	s.True(ids["track-001"])
	s.True(ids["track-002"])
	s.True(ids["track-003"])
}

// ------------------------------------------------------------
// Context cancellation
// ------------------------------------------------------------

func (s *TrackRepositoryTestSuite) TestCreateContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.repository.Create(
		ctx,
		testTrack("track-001"),
	)

	s.Require().Error(err)
	s.True(errors.Is(err, context.Canceled))
}

func (s *TrackRepositoryTestSuite) TestGetByIDContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := s.repository.GetByID(
		ctx,
		"track-001",
	)

	s.Require().Error(err)
	s.Nil(result)
	s.True(errors.Is(err, context.Canceled))
}

func (s *TrackRepositoryTestSuite) TestExistsContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	exists, err := s.repository.Exists(
		ctx,
		"track-001",
	)

	s.Require().Error(err)
	s.False(exists)
	s.True(errors.Is(err, context.Canceled))
}

func (s *TrackRepositoryTestSuite) TestUpdateContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.repository.Update(
		ctx,
		testTrack("track-001"),
	)

	s.Require().Error(err)
	s.True(errors.Is(err, context.Canceled))
}

func (s *TrackRepositoryTestSuite) TestDeleteContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.repository.Delete(
		ctx,
		"track-001",
	)

	s.Require().Error(err)
	s.True(errors.Is(err, context.Canceled))
}

func (s *TrackRepositoryTestSuite) TestGetAllTracksContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := s.repository.GetAllTracks(ctx)

	s.Require().Error(err)
	s.Nil(result)
	s.True(errors.Is(err, context.Canceled))
}

// ------------------------------------------------------------
// Test entrypoint
// ------------------------------------------------------------

func TestTrackRepository(t *testing.T) {
	suite.Run(t, new(TrackRepositoryTestSuite))
}
