package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/S-VIPER/backend/gin-api/gen/db"
	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Track struct {
	queries *db.Queries
}

func NewTrackRepository(pool *pgxpool.Pool) *Track {
	return &Track{queries: db.New(pool)}
}

var _ usecase.TrackRepositoryInterface = (*Track)(nil)

func (r *Track) Create(ctx context.Context, track *domain.Track) error {
	if track == nil {
		return domain.ErrInvalidTrack
	}

	now := time.Now().UTC()
	row, err := r.queries.CreateTrack(ctx, db.CreateTrackParams{
		ID:                        track.ID,
		Title:                     track.Title,
		Artist:                    track.Artist,
		AlbumTitle:                track.AlbumTitle,
		AlbumArtURL:               track.AlbumArtURL,
		Genre:                     track.Genre,
		Year:                      int32(track.Year),
		MusicBrainzReleaseID:      track.MusicBrainzReleaseID,
		MusicBrainzReleaseGroupID: track.MusicBrainzReleaseGroupID,
		ObjectKey:                 track.ObjectKey,
		FileName:                  track.FileName,
		ContentType:               track.ContentType,
		FileSize:                  track.FileSize,
		CreatedAt:                 timeToPGTimestamptz(now),
		UpdatedAt:                 timeToPGTimestamptz(now),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrTrackAlreadyExists
		}
		return fmt.Errorf("create track: %w", err)
	}

	*track = toDomainTrack(row)
	return nil
}

func (r *Track) GetByID(ctx context.Context, id string) (*domain.Track, error) {
	row, err := r.queries.GetTrackByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTrackNotFound
		}
		return nil, fmt.Errorf("get track: %w", err)
	}

	result := toDomainTrack(row)
	return &result, nil
}

func (r *Track) Update(ctx context.Context, track *domain.Track) error {
	if track == nil {
		return domain.ErrInvalidTrack
	}

	rowsAffected, err := r.queries.UpdateTrack(ctx, db.UpdateTrackParams{
		ID:                        track.ID,
		Title:                     track.Title,
		Artist:                    track.Artist,
		AlbumTitle:                track.AlbumTitle,
		AlbumArtURL:               track.AlbumArtURL,
		Genre:                     track.Genre,
		Year:                      int32(track.Year),
		MusicBrainzReleaseID:      track.MusicBrainzReleaseID,
		MusicBrainzReleaseGroupID: track.MusicBrainzReleaseGroupID,
		UpdatedAt:                 timeToPGTimestamptz(time.Now().UTC()),
	})
	if err != nil {
		return fmt.Errorf("update track: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrTrackNotFound
	}

	return nil
}

func (r *Track) Delete(ctx context.Context, id string) error {
	rowsAffected, err := r.queries.DeleteTrack(ctx, id)
	if err != nil {
		return fmt.Errorf("delete track: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrTrackNotFound
	}

	return nil
}

func (r *Track) GetAllTracks(ctx context.Context) ([]*domain.Track, error) {
	rows, err := r.queries.GetAllTracks(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all tracks: %w", err)
	}

	tracks := make([]*domain.Track, 0, len(rows))
	for _, row := range rows {
		track := toDomainTrack(row)
		tracks = append(tracks, &track)
	}

	return tracks, nil
}

func (r *Track) Exists(ctx context.Context, id string) (bool, error) {
	exists, err := r.queries.TrackExists(ctx, id)
	if err != nil {
		return false, fmt.Errorf("check track existence: %w", err)
	}
	return exists, nil
}

func toDomainTrack(row db.Track) domain.Track {
	return domain.Track{
		ID:                        row.ID,
		Title:                     row.Title,
		Artist:                    row.Artist,
		AlbumTitle:                row.AlbumTitle,
		AlbumArtURL:               row.AlbumArtURL,
		Genre:                     row.Genre,
		Year:                      int(row.Year),
		MusicBrainzReleaseID:      row.MusicBrainzReleaseID,
		MusicBrainzReleaseGroupID: row.MusicBrainzReleaseGroupID,
		ObjectKey:                 row.ObjectKey,
		FileName:                  row.FileName,
		ContentType:               row.ContentType,
		FileSize:                  row.FileSize,
	}
}
