package repository

import (
	"context"
	"errors"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TrackRepository struct {
	db *mongo.Database
}

func NewTrackRepository(db *mongo.Database) *TrackRepository {
	return &TrackRepository{db: db}
}

var _ TrackRepositoryInterface = (*TrackRepository)(nil)

func (r *TrackRepository) Create(
	ctx context.Context,
	track *domain.Track,
) error {
	_, err := r.db.
		Collection("tracks").
		InsertOne(ctx, track)

	if err != nil {
		return err
	}

	return nil
}

func (r *TrackRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.Track, error) {
	var track domain.Track

	err := r.db.
		Collection("tracks").
		FindOne(ctx, bson.M{"_id": id}).
		Decode(&track)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrTrackNotFound
	}

	if err != nil {
		return nil, err
	}

	return &track, nil
}

func (r *TrackRepository) Update(
	ctx context.Context,
	track *domain.Track,
) error {
	result, err := r.db.
		Collection("tracks").
		UpdateOne(
			ctx,
			bson.M{"_id": track.ID},
			bson.M{"$set": track},
		)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrTrackNotFound
	}

	return nil
}

func (r *TrackRepository) Delete(
	ctx context.Context,
	id string,
) error {
	result, err := r.db.
		Collection("tracks").
		DeleteOne(ctx, bson.M{"_id": id})

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return domain.ErrTrackNotFound
	}

	return nil
}

func (r *TrackRepository) GetAllTracks(
	ctx context.Context,
) ([]*domain.Track, error) {
	cursor, err := r.db.
		Collection("tracks").
		Find(ctx, bson.M{})

	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tracks []*domain.Track

	if err := cursor.All(ctx, &tracks); err != nil {
		return nil, err
	}

	return tracks, nil
}

func (r *TrackRepository) Exists(ctx context.Context, id string) (bool, error) {
	count, err := r.db.Collection("tracks").CountDocuments(
		ctx,
		bson.M{"_id": id},
		options.Count().SetLimit(1),
	)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
