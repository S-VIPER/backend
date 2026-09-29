package rustfs_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	trackstorage "github.com/S-VIPER/backend/gin-api/internal/repository/rustfs"
	"github.com/stretchr/testify/require"
)

func TestTrackStoragePresignUsesPublicEndpoint(t *testing.T) {
	storage, err := trackstorage.NewTrackStorage(trackstorage.Config{
		Endpoint:       "http://rustfs:9000",
		PublicEndpoint: "http://localhost:9000",
		AccessKey:      "RUSTFSDEV",
		SecretKey:      "secret",
		Bucket:         "tracks",
		Region:         "us-east-1",
	})
	require.NoError(t, err)

	signed, err := storage.PresignGetObject(
		context.Background(),
		"tracks/550e8400-e29b-41d4-a716-446655440000/file.mp3",
		5*time.Minute,
	)
	require.NoError(t, err)

	parsed, err := url.Parse(signed)
	require.NoError(t, err)
	require.Equal(t, "localhost:9000", parsed.Host)
	require.Contains(t, parsed.Path, "/tracks/tracks/550e8400-e29b-41d4-a716-446655440000/file.mp3")
	require.NotEmpty(t, parsed.Query().Get("X-Amz-Signature"))
}

func TestTrackStorageRejectsInvalidEndpoint(t *testing.T) {
	_, err := trackstorage.NewTrackStorage(trackstorage.Config{
		Endpoint:       "rustfs:9000",
		PublicEndpoint: "http://localhost:9000",
		AccessKey:      "RUSTFSDEV",
		SecretKey:      "secret",
		Bucket:         "tracks",
	})
	require.Error(t, err)
}

func TestTrackStorageRequiresCredentials(t *testing.T) {
	_, err := trackstorage.NewTrackStorage(trackstorage.Config{
		Endpoint:       "http://rustfs:9000",
		PublicEndpoint: "http://localhost:9000",
		Bucket:         "tracks",
	})
	require.Error(t, err)
}
