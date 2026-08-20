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
	document := trackDocumentFromDomain(track)

	_, err := r.db.
		Collection("tracks").
		InsertOne(ctx, document)

	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrTrackAlreadyExists
		}

		return err
	}

	return nil
}

func (r *TrackRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.Track, error) {
	var document trackDocument

	err := r.db.
		Collection("tracks").
		FindOne(
			ctx,
			bson.M{"_id": id},
		).
		Decode(&document)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrTrackNotFound
	}

	if err != nil {
		return nil, err
	}

	return document.toDomain(), nil
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
			bson.M{
				"$set": bson.M{
					"title":       track.Title,
					"artist":      track.Artist,
					"url":         track.URL,
					"albumTitle":  track.AlbumTitle,
					"albumArtURL": track.AlbumArtURL,
					"previewURL":  track.PreviewURL,
					"genre":       track.Genre,
					"year":        track.Year,
				},
			},
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

	var documents []trackDocument

	if err := cursor.All(ctx, &documents); err != nil {
		return nil, err
	}

	tracks := make([]*domain.Track, 0, len(documents))

	for _, document := range documents {
		tracks = append(tracks, document.toDomain())
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

type trackDocument struct {
	ID          string   `bson:"_id,omitempty"`
	Title       string   `bson:"title"`
	Artist      string   `bson:"artist"`
	URL         string   `bson:"url"`
	AlbumTitle  string   `bson:"albumTitle"`
	AlbumArtURL string   `bson:"albumArtURL"`
	PreviewURL  string   `bson:"previewURL"`
	Genre       []string `bson:"genre"`
	Year        int      `bson:"year"`
}

func trackDocumentFromDomain(track *domain.Track) trackDocument {
	return trackDocument{
		ID:          track.ID,
		Title:       track.Title,
		Artist:      track.Artist,
		URL:         track.URL,
		AlbumTitle:  track.AlbumTitle,
		AlbumArtURL: track.AlbumArtURL,
		PreviewURL:  track.PreviewURL,
		Genre:       track.Genre,
		Year:        track.Year,
	}
}

func (d trackDocument) toDomain() *domain.Track {
	return &domain.Track{
		ID:          d.ID,
		Title:       d.Title,
		Artist:      d.Artist,
		URL:         d.URL,
		AlbumTitle:  d.AlbumTitle,
		AlbumArtURL: d.AlbumArtURL,
		PreviewURL:  d.PreviewURL,
		Genre:       d.Genre,
		Year:        d.Year,
	}
}
