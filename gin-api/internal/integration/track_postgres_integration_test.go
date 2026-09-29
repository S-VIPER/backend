//go:build integration

package integration_test

import (
	"context"
	"testing"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/repository/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestTrackPostgresRepository(t *testing.T) {
	ctx := context.Background()

	container, err := postgres.Run(
		ctx,
		"postgres:18.6-alpine",
		postgres.WithDatabase("sviper-track-test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp"),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	connectionString, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connectionString)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	require.NoError(t, pool.Ping(ctx))
	runMigrations(t, pool)

	repo := postgres.NewTrackRepository(pool)

	track := &domain.Track{
		ID:                        "550e8400-e29b-41d4-a716-446655440000",
		Title:                     "Integration Track",
		Artist:                    "Integration Artist",
		AlbumTitle:                "Integration Album",
		AlbumArtURL:               "https://example.com/album.jpg",
		Genre:                     []string{"Electronic"},
		Year:                      2025,
		MusicBrainzReleaseID:      "550e8400-e29b-41d4-a716-446655440001",
		MusicBrainzReleaseGroupID: "550e8400-e29b-41d4-a716-446655440002",
		ObjectKey:                 "tracks/550e8400-e29b-41d4-a716-446655440000/file.mp3",
		FileName:                  "file.mp3",
		ContentType:               "audio/mpeg",
		FileSize:                  1024,
	}

	require.NoError(t, repo.Create(ctx, track))

	got, err := repo.GetByID(ctx, track.ID)
	require.NoError(t, err)
	require.Equal(t, track.ID, got.ID)
	require.Equal(t, track.ObjectKey, got.ObjectKey)
	require.Equal(t, track.MusicBrainzReleaseID, got.MusicBrainzReleaseID)

	track.Title = "Updated Track"
	require.NoError(t, repo.Update(ctx, track))

	got, err = repo.GetByID(ctx, track.ID)
	require.NoError(t, err)
	require.Equal(t, "Updated Track", got.Title)

	require.NoError(t, repo.Delete(ctx, track.ID))
	_, err = repo.GetByID(ctx, track.ID)
	require.ErrorIs(t, err, domain.ErrTrackNotFound)
}
