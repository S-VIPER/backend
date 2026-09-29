package musicbrainz_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/S-VIPER/backend/gin-api/internal/repository/musicbrainz"
	"github.com/stretchr/testify/require"
)

func TestClientSearchAndLookup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "S-VIPER/test (+https://example.com/contact)", r.Header.Get("User-Agent"))
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/recording":
			_, _ = w.Write([]byte(`{
				"recordings": [{
					"id": "550e8400-e29b-41d4-a716-446655440000",
					"title": "Test Track",
					"score": 100,
					"artist-credit": [{"name":"Test Artist","joinphrase":""}]
				}]
			}`))
		case "/recording/550e8400-e29b-41d4-a716-446655440000":
			_, _ = w.Write([]byte(`{
				"id": "550e8400-e29b-41d4-a716-446655440000",
				"title": "Test Track",
				"artist-credit": [{"name":"Test Artist","joinphrase":""}],
				"genres": [{"name":"Electronic"}],
				"releases": [{
					"id":"550e8400-e29b-41d4-a716-446655440001",
					"title":"Test Album",
					"date":"2025-01-02",
					"release-group":{"id":"550e8400-e29b-41d4-a716-446655440002"}
				}]
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := musicbrainz.NewClient(musicbrainz.Config{
		BaseURL:     server.URL,
		UserAgent:   "S-VIPER/test (+https://example.com/contact)",
		MinInterval: 0,
	})
	require.NoError(t, err)

	ctx := context.Background()
	candidates, err := client.SearchRecordings(ctx, "Test Artist", "Test Track", 10)
	require.NoError(t, err)
	require.Len(t, candidates, 1)
	require.Equal(t, "550e8400-e29b-41d4-a716-446655440000", candidates[0].MBID)

	metadata, err := client.GetRecording(ctx, candidates[0].MBID)
	require.NoError(t, err)
	require.Equal(t, "Test Album", metadata.AlbumTitle)
	require.Equal(t, 2025, metadata.Year)
	require.Equal(t, []string{"Electronic"}, metadata.Genre)
}

func TestClientRateLimitRespectsContext(t *testing.T) {
	client, err := musicbrainz.NewClient(musicbrainz.Config{
		BaseURL:     "http://127.0.0.1",
		UserAgent:   "S-VIPER/test",
		MinInterval: time.Second,
	})
	require.NoError(t, err)

	ctx := context.Background()
	start := time.Now()
	_, _ = client.SearchRecordings(ctx, "a", "b", 1)
	elapsed := time.Since(start)

	require.Less(t, elapsed, 2*time.Second)
}
