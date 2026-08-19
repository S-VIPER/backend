package usecase

import (
	"context"
	"testing"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/repository"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TrackUseCaseIntegrationTestSuite struct {
	suite.Suite

	ctx       context.Context
	container *mongodb.MongoDBContainer

	client *mongo.Client
	db     *mongo.Database

	repository *repository.TrackRepository
	useCase    *TrackUseCase
}

func (s *TrackUseCaseIntegrationTestSuite) SetupSuite() {
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
	s.db = client.Database("track_usecase_integration_test")

	s.repository = repository.NewTrackRepository(s.db)
	s.useCase = NewTrackUseCase(s.repository)
}

func (s *TrackUseCaseIntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		err := s.db.Drop(s.ctx)
		require.NoError(s.T(), err)
	}

	if s.client != nil {
		err := s.client.Disconnect(s.ctx)
		require.NoError(s.T(), err)
	}

	if s.container != nil {
		err := testcontainers.TerminateContainer(
			s.container,
		)
		require.NoError(s.T(), err)
	}
}

func (s *TrackUseCaseIntegrationTestSuite) SetupTest() {
	err := s.db.Collection("tracks").Drop(s.ctx)
	require.NoError(s.T(), err)
}

func integrationTrack(id string) *domain.Track {
	return &domain.Track{
		ID:          id,
		Title:       "Integration Track",
		Artist:      "Integration Artist",
		URL:         "https://example.com/integration.mp3",
		AlbumTitle:  "Integration Album",
		AlbumArtURL: "https://example.com/album.jpg",
		PreviewURL:  "https://example.com/preview.mp3",
		Genre:       []string{"Rock"},
		Year:        2026,
	}
}

func (s *TrackUseCaseIntegrationTestSuite) TestCreateAndGetTrack() {
	ctx := context.Background()

	track := integrationTrack("integration-001")

	err := s.useCase.CreateTrack(ctx, track)
	s.Require().NoError(err)

	result, err := s.useCase.GetTrackByID(
		ctx,
		track.ID,
	)

	s.Require().NoError(err)
	s.Require().NotNil(result)

	s.Equal(track.ID, result.ID)
	s.Equal(track.Title, result.Title)
	s.Equal(track.Artist, result.Artist)
}

func (s *TrackUseCaseIntegrationTestSuite) TestUpdateTrack() {
	ctx := context.Background()

	track := integrationTrack("integration-001")

	err := s.useCase.CreateTrack(ctx, track)
	s.Require().NoError(err)

	track.Title = "Updated integration track"
	track.Year = 2027

	err = s.useCase.UpdateTrack(ctx, track)
	s.Require().NoError(err)

	result, err := s.useCase.GetTrackByID(
		ctx,
		track.ID,
	)

	s.Require().NoError(err)

	s.Equal("Updated integration track", result.Title)
	s.Equal(2027, result.Year)
}

func (s *TrackUseCaseIntegrationTestSuite) TestDeleteTrack() {
	ctx := context.Background()

	track := integrationTrack("integration-001")

	err := s.useCase.CreateTrack(ctx, track)
	s.Require().NoError(err)

	err = s.useCase.DeleteTrack(ctx, track.ID)
	s.Require().NoError(err)

	result, err := s.useCase.GetTrackByID(
		ctx,
		track.ID,
	)

	s.Require().Error(err)
	s.Nil(result)

	s.ErrorIs(err, domain.ErrTrackNotFound)
}

func (s *TrackUseCaseIntegrationTestSuite) TestGetAllTracks() {
	ctx := context.Background()

	tracks := []*domain.Track{
		integrationTrack("integration-001"),
		integrationTrack("integration-002"),
		integrationTrack("integration-003"),
	}

	for _, track := range tracks {
		err := s.useCase.CreateTrack(ctx, track)
		s.Require().NoError(err)
	}

	result, err := s.useCase.GetAllTracks(ctx)

	s.Require().NoError(err)
	s.Require().Len(result, 3)
}

func TestTrackUseCaseIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(TrackUseCaseIntegrationTestSuite))
}
