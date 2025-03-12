package repository

import (
	"context"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TrackRepository struct {
	db *mongo.Database
}

func NewTrackRepository(db *mongo.Database) *TrackRepository {
	return &TrackRepository{db: db}
}

func (r *TrackRepository) Create(track *domain.Track) error {
	_, err := r.db.Collection("tracks").InsertOne(context.TODO(), track)
	return err
}

func (r *TrackRepository) GetByID(id string) (*domain.Track, error) {
	var track domain.Track
	err := r.db.Collection("tracks").FindOne(context.TODO(), bson.M{"_id": id}).Decode(&track)
	if err != nil {
		return nil, err
	}
	return &track, nil
}

func (r *TrackRepository) Update(track *domain.Track) error {
	filter := bson.M{"_id": track.ID}
	update := bson.M{"$set": track}
	_, err := r.db.Collection("tracks").UpdateOne(context.TODO(), filter, update)
	return err
}

func (r *TrackRepository) Delete(id string) error {
	_, err := r.db.Collection("tracks").DeleteOne(context.TODO(), bson.M{"_id": id})
	return err
}
