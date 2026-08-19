package repository

import (
	"context"
	"errors"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PlaylistRepository struct {
	collection *mongo.Collection
}

var _ PlaylistRepositoryInterface = (*PlaylistRepository)(nil)

func NewPlaylistRepository(db *mongo.Database) *PlaylistRepository {
	return &PlaylistRepository{
		collection: db.Collection("playlists"),
	}
}

func (r *PlaylistRepository) Create(
	ctx context.Context,
	playlist *domain.Playlist,
) error {
	objID, err := createObjectID(playlist.ID)
	if err != nil {
		return err
	}

	document := bson.M{
		"_id":    objID,
		"name":   playlist.Name,
		"tracks": playlist.Tracks,
	}

	_, err = r.collection.InsertOne(ctx, document)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrPlaylistAlreadyExists
		}

		return err
	}

	// Если ID не был задан, сохраняем сгенерированный ObjectID
	// обратно в domain entity.
	playlist.ID = objID.Hex()

	return nil
}

func (r *PlaylistRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.Playlist, error) {
	objID, err := parseObjectID(id)
	if err != nil {
		return nil, err
	}

	var document playlistDocument

	err = r.collection.
		FindOne(ctx, bson.M{"_id": objID}).
		Decode(&document)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrPlaylistNotFound
	}

	if err != nil {
		return nil, err
	}

	return document.toDomain(), nil
}

func (r *PlaylistRepository) Update(
	ctx context.Context,
	playlist *domain.Playlist,
) error {
	objID, err := parseObjectID(playlist.ID)
	if err != nil {
		return err
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"name":   playlist.Name,
				"tracks": playlist.Tracks,
			},
		},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrPlaylistNotFound
	}

	return nil
}

func (r *PlaylistRepository) Delete(
	ctx context.Context,
	id string,
) error {
	objID, err := parseObjectID(id)
	if err != nil {
		return err
	}

	result, err := r.collection.DeleteOne(
		ctx,
		bson.M{"_id": objID},
	)

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return domain.ErrPlaylistNotFound
	}

	return nil
}

func (r *PlaylistRepository) AddTrack(
	ctx context.Context,
	playlistID string,
	trackID string,
) error {
	objID, err := parseObjectID(playlistID)
	if err != nil {
		return err
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{
			"$addToSet": bson.M{
				"tracks": trackID,
			},
		},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrPlaylistNotFound
	}

	return nil
}

func (r *PlaylistRepository) RemoveTrack(
	ctx context.Context,
	playlistID string,
	trackID string,
) error {
	objID, err := parseObjectID(playlistID)
	if err != nil {
		return err
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{
			"$pull": bson.M{
				"tracks": trackID,
			},
		},
	)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrPlaylistNotFound
	}

	return nil
}

func parseObjectID(id string) (primitive.ObjectID, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, domain.ErrInvalidPlaylistID
	}

	return objID, nil
}

func createObjectID(id string) (primitive.ObjectID, error) {
	if id == "" {
		return primitive.NewObjectID(), nil
	}

	return parseObjectID(id)
}

// dto
type playlistDocument struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Name   string             `bson:"name"`
	Tracks []string           `bson:"tracks"`
}

func (d playlistDocument) toDomain() *domain.Playlist {
	return &domain.Playlist{
		ID:     d.ID.Hex(),
		Name:   d.Name,
		Tracks: d.Tracks,
	}
}
