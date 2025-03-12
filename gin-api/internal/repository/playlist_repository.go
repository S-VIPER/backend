package repository

import (
	"context"

	"github.com/S-VIPER/backend/gin-api/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PlaylistRepository struct {
	collection *mongo.Collection
}

func NewPlaylistRepository(db *mongo.Database) *PlaylistRepository {
	return &PlaylistRepository{collection: db.Collection("playlists")}
}

func (r *PlaylistRepository) Create(playlist *domain.Playlist) error {
	_, err := r.collection.InsertOne(context.Background(), playlist)
	return err
}

func (r *PlaylistRepository) GetByID(id string) (*domain.Playlist, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var playlist domain.Playlist
	err = r.collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&playlist)
	return &playlist, err
}

func (r *PlaylistRepository) Update(playlist *domain.Playlist) error {
	objID, err := primitive.ObjectIDFromHex(playlist.ID)
	if err != nil {
		return err
	}

	_, err = r.collection.ReplaceOne(context.Background(), bson.M{"_id": objID}, playlist)
	return err
}

func (r *PlaylistRepository) Delete(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(context.Background(), bson.M{"_id": objID})
	return err
}

func (r *PlaylistRepository) AddTrack(playlistID, trackID string) error {
	objID, err := primitive.ObjectIDFromHex(playlistID)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$addToSet": bson.M{"tracks": trackID}},
	)
	return err
}

func (r *PlaylistRepository) RemoveTrack(playlistID, trackID string) error {
	objID, err := primitive.ObjectIDFromHex(playlistID)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$pull": bson.M{"tracks": trackID}},
	)
	return err
}
