package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/S-VIPER/backend/gin-api/gen/db"
	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Playlist struct {
	queries *db.Queries
}

func NewPlaylistRepository(pool *pgxpool.Pool) *Playlist {
	return &Playlist{queries: db.New(pool)}
}

var _ usecase.PlaylistRepositoryInterface = (*Playlist)(nil)

func (r *Playlist) Create(ctx context.Context, playlist *domain.Playlist) error {
	id := uuid.New()
	tracks := playlist.Tracks
	if tracks == nil {
		tracks = []string{}
	}
	if playlist.ID != "" {
		parsed, err := uuid.Parse(playlist.ID)
		if err != nil {
			return domain.ErrInvalidPlaylistID
		}
		id = parsed
	}

	visibility := playlist.Visibility
	if visibility == "" {
		visibility = domain.PlaylistVisibilityPrivate
	}

	now := time.Now()

	row, err := r.queries.CreatePlaylist(ctx, db.CreatePlaylistParams{
		ID:         uuidToPGUUID(id),
		OwnerID:    uuidToPGUUID(playlist.OwnerID),
		Name:       playlist.Name,
		Visibility: string(visibility),
		Tracks:     tracks,
		CreatedAt:  timeToPGTimestamptz(now),
		UpdatedAt:  timeToPGTimestamptz(now),
	})
	if err != nil {
		return fmt.Errorf("create playlist: %w", err)
	}

	*playlist = toDomainPlaylist(row)
	return nil
}

func (r *Playlist) GetByID(ctx context.Context, id string) (*domain.Playlist, error) {
	playlistID, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrInvalidPlaylistID
	}

	row, err := r.queries.GetPlaylistByID(ctx, uuidToPGUUID(playlistID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPlaylistNotFound
		}
		return nil, fmt.Errorf("get playlist: %w", err)
	}

	result := toDomainPlaylist(row)
	return &result, nil
}

func (r *Playlist) Update(ctx context.Context, playlist *domain.Playlist) error {
	playlistID, err := uuid.Parse(playlist.ID)
	if err != nil {
		return domain.ErrInvalidPlaylistID
	}

	visibility := playlist.Visibility
	if visibility == "" {
		visibility = domain.PlaylistVisibilityPrivate
	}

	if err := r.queries.UpdatePlaylist(ctx, db.UpdatePlaylistParams{
		ID:         uuidToPGUUID(playlistID),
		Name:       playlist.Name,
		Visibility: string(visibility),
		Tracks:     playlist.Tracks,
	}); err != nil {
		return fmt.Errorf("update playlist: %w", err)
	}

	return nil
}

func (r *Playlist) Delete(ctx context.Context, id string) error {
	playlistID, err := uuid.Parse(id)
	if err != nil {
		return domain.ErrInvalidPlaylistID
	}

	if err := r.queries.DeletePlaylist(ctx, uuidToPGUUID(playlistID)); err != nil {
		return fmt.Errorf("delete playlist: %w", err)
	}

	return nil
}

func (r *Playlist) AddTrack(ctx context.Context, playlistID, trackID string) error {
	id, err := uuid.Parse(playlistID)
	if err != nil {
		return domain.ErrInvalidPlaylistID
	}

	if err := r.queries.AddTrackToPlaylist(ctx, db.AddTrackToPlaylistParams{
		ID:      uuidToPGUUID(id),
		TrackID: trackID,
	}); err != nil {
		return fmt.Errorf("add track to playlist: %w", err)
	}

	return nil
}

func (r *Playlist) RemoveTrack(ctx context.Context, playlistID, trackID string) error {
	id, err := uuid.Parse(playlistID)
	if err != nil {
		return domain.ErrInvalidPlaylistID
	}

	if err := r.queries.RemoveTrackFromPlaylist(ctx, db.RemoveTrackFromPlaylistParams{
		ID:      uuidToPGUUID(id),
		TrackID: trackID,
	}); err != nil {
		return fmt.Errorf("remove track from playlist: %w", err)
	}

	return nil
}

func toDomainPlaylist(row db.Playlist) domain.Playlist {
	return domain.Playlist{
		ID:         uuid.UUID(row.ID.Bytes).String(),
		OwnerID:    uuid.UUID(row.OwnerID.Bytes),
		Name:       row.Name,
		Tracks:     row.Tracks,
		Visibility: domain.PlaylistVisibility(row.Visibility),
	}
}
