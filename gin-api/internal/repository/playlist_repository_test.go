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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlaylistRepositoryTestSuite struct {
	suite.Suite

	ctx       context.Context
	container *mongodb.MongoDBContainer
	client    *mongo.Client
	db        *mongo.Database

	repository *PlaylistRepository
}

func (s *PlaylistRepositoryTestSuite) SetupSuite() {
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
	s.db = client.Database("playlist_repository_test")

	s.repository = NewPlaylistRepository(s.db)

	var _ PlaylistRepositoryInterface = (*PlaylistRepository)(nil)
}

func (s *PlaylistRepositoryTestSuite) TearDownSuite() {
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

func (s *PlaylistRepositoryTestSuite) SetupTest() {
	err := s.db.Collection("playlists").Drop(s.ctx)
	require.NoError(s.T(), err)
}

func testPlaylist(id string) *domain.Playlist {
	return &domain.Playlist{
		ID:     id,
		Name:   "Test Playlist",
		Tracks: []string{"track-001", "track-002"},
	}
}

// ------------------------------------------------------------
// Create
// ------------------------------------------------------------

func (s *PlaylistRepositoryTestSuite) TestCreateWithProvidedID() {
	ctx := context.Background()

	id := primitive.NewObjectID().Hex()

	playlist := testPlaylist(id)

	err := s.repository.Create(ctx, playlist)

	s.Require().NoError(err)
	s.Equal(id, playlist.ID)

	objID, err := primitive.ObjectIDFromHex(id)
	s.Require().NoError(err)

	var stored struct {
		ID     primitive.ObjectID `bson:"_id"`
		Name   string             `bson:"name"`
		Tracks []string           `bson:"tracks"`
	}

	err = s.db.
		Collection("playlists").
		FindOne(ctx, bson.M{"_id": objID}).
		Decode(&stored)

	s.Require().NoError(err)

	s.Equal(id, stored.ID.Hex())
	s.Equal(playlist.Name, stored.Name)
	s.Equal(playlist.Tracks, stored.Tracks)
}

func (s *PlaylistRepositoryTestSuite) TestCreateGeneratesID() {
	ctx := context.Background()

	playlist := &domain.Playlist{
		Name:   "Test Playlist",
		Tracks: []string{"track-001"},
	}

	s.Empty(playlist.ID)

	err := s.repository.Create(ctx, playlist)

	s.Require().NoError(err)

	s.NotEmpty(playlist.ID)

	_, err = primitive.ObjectIDFromHex(playlist.ID)
	s.Require().NoError(err)

	result, err := s.repository.GetByID(ctx, playlist.ID)

	s.Require().NoError(err)
	s.Require().NotNil(result)

	s.Equal(playlist.ID, result.ID)
	s.Equal(playlist.Name, result.Name)
	s.Equal(playlist.Tracks, result.Tracks)
}

func (s *PlaylistRepositoryTestSuite) TestCreateDuplicateID() {
	ctx := context.Background()

	id := primitive.NewObjectID().Hex()

	first := testPlaylist(id)
	second := testPlaylist(id)

	err := s.repository.Create(ctx, first)
	s.Require().NoError(err)

	err = s.repository.Create(ctx, second)

	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrPlaylistAlreadyExists)
}

// ------------------------------------------------------------
// GetByID
// ------------------------------------------------------------

func (s *PlaylistRepositoryTestSuite) TestGetByID() {
	ctx := context.Background()

	id := primitive.NewObjectID()

	document := bson.M{
		"_id":    id,
		"name":   "Test Playlist",
		"tracks": []string{"track-001", "track-002"},
	}

	_, err := s.db.Collection("playlists").InsertOne(ctx, document)
	s.Require().NoError(err)

	result, err := s.repository.GetByID(ctx, id.Hex())

	s.Require().NoError(err)
	s.Require().NotNil(result)

	s.Equal(id.Hex(), result.ID)
	s.Equal("Test Playlist", result.Name)
	s.Equal([]string{"track-001", "track-002"}, result.Tracks)
}

func (s *PlaylistRepositoryTestSuite) TestGetByIDNotFound() {
	ctx := context.Background()

	id := primitive.NewObjectID().Hex()

	result, err := s.repository.GetByID(ctx, id)

	s.Require().Error(err)
	s.Nil(result)

	s.ErrorIs(err, domain.ErrPlaylistNotFound)
}

func (s *PlaylistRepositoryTestSuite) TestGetByIDInvalidID() {
	ctx := context.Background()

	result, err := s.repository.GetByID(ctx, "invalid-id")

	s.Require().Error(err)
	s.Nil(result)

	s.ErrorIs(err, domain.ErrInvalidPlaylistID)
}

// ------------------------------------------------------------
// Update
// ------------------------------------------------------------

func (s *PlaylistRepositoryTestSuite) TestUpdate() {
	ctx := context.Background()

	id := primitive.NewObjectID().Hex()

	playlist := testPlaylist(id)

	err := s.repository.Create(ctx, playlist)
	s.Require().NoError(err)

	playlist.Name = "Updated Playlist"
	playlist.Tracks = []string{
		"track-003",
		"track-004",
	}

	err = s.repository.Update(ctx, playlist)

	s.Require().NoError(err)

	result, err := s.repository.GetByID(ctx, id)

	s.Require().NoError(err)
	s.Require().NotNil(result)

	s.Equal(id, result.ID)
	s.Equal("Updated Playlist", result.Name)
	s.Equal(
		[]string{"track-003", "track-004"},
		result.Tracks,
	)
}

func (s *PlaylistRepositoryTestSuite) TestUpdateNotFound() {
	ctx := context.Background()

	playlist := testPlaylist(
		primitive.NewObjectID().Hex(),
	)

	err := s.repository.Update(ctx, playlist)

	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrPlaylistNotFound)
}

func (s *PlaylistRepositoryTestSuite) TestUpdateInvalidID() {
	ctx := context.Background()

	playlist := testPlaylist("invalid-id")

	err := s.repository.Update(ctx, playlist)

	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrInvalidPlaylistID)
}

// ------------------------------------------------------------
// Delete
// ------------------------------------------------------------

func (s *PlaylistRepositoryTestSuite) TestDelete() {
	ctx := context.Background()

	id := primitive.NewObjectID().Hex()

	playlist := testPlaylist(id)

	err := s.repository.Create(ctx, playlist)
	s.Require().NoError(err)

	err = s.repository.Delete(ctx, id)

	s.Require().NoError(err)

	objID, err := primitive.ObjectIDFromHex(id)
	s.Require().NoError(err)

	var result bson.M

	err = s.db.
		Collection("playlists").
		FindOne(ctx, bson.M{"_id": objID}).
		Decode(&result)

	s.ErrorIs(err, mongo.ErrNoDocuments)
}

func (s *PlaylistRepositoryTestSuite) TestDeleteNotFound() {
	ctx := context.Background()

	id := primitive.NewObjectID().Hex()

	err := s.repository.Delete(ctx, id)

	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrPlaylistNotFound)
}

func (s *PlaylistRepositoryTestSuite) TestDeleteInvalidID() {
	ctx := context.Background()

	err := s.repository.Delete(ctx, "invalid-id")

	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrInvalidPlaylistID)
}

// ------------------------------------------------------------
// AddTrack
// ------------------------------------------------------------

func (s *PlaylistRepositoryTestSuite) TestAddTrack() {
	ctx := context.Background()

	id := primitive.NewObjectID().Hex()

	playlist := &domain.Playlist{
		ID:     id,
		Name:   "Test Playlist",
		Tracks: []string{"track-001"},
	}

	err := s.repository.Create(ctx, playlist)
	s.Require().NoError(err)

	err = s.repository.AddTrack(
		ctx,
		id,
		"track-002",
	)

	s.Require().NoError(err)

	result, err := s.repository.GetByID(ctx, id)

	s.Require().NoError(err)
	s.Require().NotNil(result)

	s.Equal(
		[]string{"track-001", "track-002"},
		result.Tracks,
	)
}

func (s *PlaylistRepositoryTestSuite) TestAddTrackDoesNotCreateDuplicate() {
	ctx := context.Background()

	id := primitive.NewObjectID().Hex()

	playlist := &domain.Playlist{
		ID:     id,
		Name:   "Test Playlist",
		Tracks: []string{"track-001"},
	}

	err := s.repository.Create(ctx, playlist)
	s.Require().NoError(err)

	err = s.repository.AddTrack(ctx, id, "track-001")
	s.Require().NoError(err)

	result, err := s.repository.GetByID(ctx, id)

	s.Require().NoError(err)
	s.Require().Len(result.Tracks, 1)
	s.Equal("track-001", result.Tracks[0])
}

func (s *PlaylistRepositoryTestSuite) TestAddTrackPlaylistNotFound() {
	ctx := context.Background()

	id := primitive.NewObjectID().Hex()

	err := s.repository.AddTrack(
		ctx,
		id,
		"track-001",
	)

	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrPlaylistNotFound)
}

func (s *PlaylistRepositoryTestSuite) TestAddTrackInvalidPlaylistID() {
	ctx := context.Background()

	err := s.repository.AddTrack(
		ctx,
		"invalid-id",
		"track-001",
	)

	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrInvalidPlaylistID)
}

// ------------------------------------------------------------
// RemoveTrack
// ------------------------------------------------------------

func (s *PlaylistRepositoryTestSuite) TestRemoveTrack() {
	ctx := context.Background()

	id := primitive.NewObjectID().Hex()

	playlist := &domain.Playlist{
		ID:   id,
		Name: "Test Playlist",
		Tracks: []string{
			"track-001",
			"track-002",
			"track-003",
		},
	}

	err := s.repository.Create(ctx, playlist)
	s.Require().NoError(err)

	err = s.repository.RemoveTrack(
		ctx,
		id,
		"track-002",
	)

	s.Require().NoError(err)

	result, err := s.repository.GetByID(ctx, id)

	s.Require().NoError(err)
	s.Require().NotNil(result)

	s.Equal(
		[]string{
			"track-001",
			"track-003",
		},
		result.Tracks,
	)
}

func (s *PlaylistRepositoryTestSuite) TestRemoveTrackThatDoesNotExist() {
	ctx := context.Background()

	id := primitive.NewObjectID().Hex()

	playlist := &domain.Playlist{
		ID:     id,
		Name:   "Test Playlist",
		Tracks: []string{"track-001"},
	}

	err := s.repository.Create(ctx, playlist)
	s.Require().NoError(err)

	err = s.repository.RemoveTrack(
		ctx,
		id,
		"track-999",
	)

	s.Require().NoError(err)

	result, err := s.repository.GetByID(ctx, id)

	s.Require().NoError(err)
	s.Equal([]string{"track-001"}, result.Tracks)
}

func (s *PlaylistRepositoryTestSuite) TestRemoveTrackPlaylistNotFound() {
	ctx := context.Background()

	id := primitive.NewObjectID().Hex()

	err := s.repository.RemoveTrack(
		ctx,
		id,
		"track-001",
	)

	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrPlaylistNotFound)
}

func (s *PlaylistRepositoryTestSuite) TestRemoveTrackInvalidPlaylistID() {
	ctx := context.Background()

	err := s.repository.RemoveTrack(
		ctx,
		"invalid-id",
		"track-001",
	)

	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrInvalidPlaylistID)
}

// ------------------------------------------------------------
// Context cancellation
// ------------------------------------------------------------

func (s *PlaylistRepositoryTestSuite) TestCreateContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.repository.Create(
		ctx,
		testPlaylist(""),
	)

	s.Require().Error(err)
	s.True(errors.Is(err, context.Canceled))
}

func (s *PlaylistRepositoryTestSuite) TestGetByIDContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	id := primitive.NewObjectID().Hex()

	result, err := s.repository.GetByID(ctx, id)

	s.Require().Error(err)
	s.Nil(result)
	s.True(errors.Is(err, context.Canceled))
}

func (s *PlaylistRepositoryTestSuite) TestUpdateContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	playlist := testPlaylist(
		primitive.NewObjectID().Hex(),
	)

	err := s.repository.Update(ctx, playlist)

	s.Require().Error(err)
	s.True(errors.Is(err, context.Canceled))
}

func (s *PlaylistRepositoryTestSuite) TestDeleteContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	id := primitive.NewObjectID().Hex()

	err := s.repository.Delete(ctx, id)

	s.Require().Error(err)
	s.True(errors.Is(err, context.Canceled))
}

func (s *PlaylistRepositoryTestSuite) TestAddTrackContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	id := primitive.NewObjectID().Hex()

	err := s.repository.AddTrack(
		ctx,
		id,
		"track-001",
	)

	s.Require().Error(err)
	s.True(errors.Is(err, context.Canceled))
}

func (s *PlaylistRepositoryTestSuite) TestRemoveTrackContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	id := primitive.NewObjectID().Hex()

	err := s.repository.RemoveTrack(
		ctx,
		id,
		"track-001",
	)

	s.Require().Error(err)
	s.True(errors.Is(err, context.Canceled))
}

// ------------------------------------------------------------
// Test entrypoint
// ------------------------------------------------------------

func TestPlaylistRepository(t *testing.T) {
	suite.Run(t, new(PlaylistRepositoryTestSuite))
}
