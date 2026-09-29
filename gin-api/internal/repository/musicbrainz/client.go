package musicbrainz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
)

const defaultBaseURL = "https://musicbrainz.org/ws/2"

// Client is an infrastructure adapter for the public MusicBrainz API.
type Client struct {
	baseURL     string
	userAgent   string
	httpClient  *http.Client
	minInterval time.Duration

	mu       sync.Mutex
	nextCall time.Time
}

type Config struct {
	BaseURL     string
	UserAgent   string
	HTTPClient  *http.Client
	MinInterval time.Duration
}

func NewClient(cfg Config) (*Client, error) {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if cfg.UserAgent == "" {
		return nil, fmt.Errorf("musicbrainz user-agent is required")
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	if cfg.MinInterval <= 0 {
		cfg.MinInterval = time.Second
	}

	return &Client{
		baseURL:     baseURL,
		userAgent:   cfg.UserAgent,
		httpClient:  cfg.HTTPClient,
		minInterval: cfg.MinInterval,
	}, nil
}

type searchResponse struct {
	Recordings []recording `json:"recordings"`
}

type recording struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	Score        int            `json:"score"`
	ArtistCredit []artistCredit `json:"artist-credit"`
	Releases     []release      `json:"releases"`
	Genres       []genre        `json:"genres"`
}

type artistCredit struct {
	Name       string `json:"name"`
	JoinPhrase string `json:"joinphrase"`
	Artist     artist `json:"artist"`
}

type artist struct {
	Name string `json:"name"`
}

type release struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Date         string       `json:"date"`
	ReleaseGroup releaseGroup `json:"release-group"`
}

type releaseGroup struct {
	ID string `json:"id"`
}

type genre struct {
	Name string `json:"name"`
}

func (c *Client) SearchRecordings(
	ctx context.Context,
	artistName string,
	title string,
	limit int,
) ([]domain.TrackMetadataCandidate, error) {
	query := fmt.Sprintf(
		`artist:"%s" AND recording:"%s"`,
		escapePhrase(artistName),
		escapePhrase(title),
	)

	values := url.Values{}
	values.Set("query", query)
	values.Set("fmt", "json")
	values.Set("limit", strconv.Itoa(limit))

	var response searchResponse
	if err := c.getJSON(ctx, "/recording", values, &response); err != nil {
		return nil, err
	}

	result := make([]domain.TrackMetadataCandidate, 0, len(response.Recordings))
	for _, item := range response.Recordings {
		candidate := toCandidate(item)
		if candidate.MBID == "" {
			continue
		}
		result = append(result, candidate)
	}

	return result, nil
}

func (c *Client) GetRecording(
	ctx context.Context,
	mbid string,
) (*domain.TrackMetadata, error) {
	values := url.Values{}
	values.Set("inc", "artists+releases+release-groups+genres")
	values.Set("fmt", "json")

	var recordingData recording
	if err := c.getJSON(ctx, "/recording/"+url.PathEscape(mbid), values, &recordingData); err != nil {
		return nil, err
	}

	metadata := toMetadata(recordingData)
	if metadata.MBID == "" || metadata.Title == "" || metadata.Artist == "" {
		return nil, domain.ErrInvalidTrack
	}

	return &metadata, nil
}

func (c *Client) getJSON(
	ctx context.Context,
	path string,
	query url.Values,
	out any,
) error {
	if err := c.waitRateLimit(ctx); err != nil {
		return err
	}

	reqURL := c.baseURL + path + "?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("create musicbrainz request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("musicbrainz request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return domain.ErrTrackNotFound
	}
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
		return domain.ErrMusicBrainzUnavailable
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("musicbrainz returned status %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode musicbrainz response: %w", err)
	}

	return nil
}

func (c *Client) waitRateLimit(ctx context.Context) error {
	c.mu.Lock()
	wait := time.Until(c.nextCall)
	if wait < 0 {
		wait = 0
	}
	c.nextCall = time.Now().Add(wait + c.minInterval)
	c.mu.Unlock()

	if wait == 0 {
		return nil
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func toCandidate(item recording) domain.TrackMetadataCandidate {
	candidate := domain.TrackMetadataCandidate{
		MBID:   item.ID,
		Title:  item.Title,
		Artist: artistCreditName(item.ArtistCredit),
		Score:  item.Score,
	}

	if len(item.Releases) > 0 {
		candidate.AlbumTitle = item.Releases[0].Title
		candidate.MusicBrainzReleaseID = item.Releases[0].ID
		candidate.MusicBrainzReleaseGroupID = item.Releases[0].ReleaseGroup.ID
	}

	return candidate
}

func toMetadata(item recording) domain.TrackMetadata {
	metadata := domain.TrackMetadata{
		MBID:   item.ID,
		Title:  item.Title,
		Artist: artistCreditName(item.ArtistCredit),
		Genre:  genreNames(item.Genres),
	}

	if len(item.Releases) > 0 {
		release := item.Releases[0]
		metadata.AlbumTitle = release.Title
		metadata.MusicBrainzReleaseID = release.ID
		metadata.MusicBrainzReleaseGroupID = release.ReleaseGroup.ID
		metadata.Year = releaseYear(release.Date)
		if release.ID != "" {
			metadata.AlbumArtURL = fmt.Sprintf(
				"https://coverartarchive.org/release/%s/front-500",
				release.ID,
			)
		}
	}

	return metadata
}

func artistCreditName(credits []artistCredit) string {
	var builder strings.Builder
	for _, credit := range credits {
		name := credit.Name
		if name == "" {
			name = credit.Artist.Name
		}
		builder.WriteString(name)
		builder.WriteString(credit.JoinPhrase)
	}
	return strings.TrimSpace(builder.String())
}

func genreNames(values []genre) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, item := range values {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	return result
}

func releaseYear(value string) int {
	if len(value) < 4 {
		return 0
	}
	year, err := strconv.Atoi(value[:4])
	if err != nil {
		return 0
	}
	return year
}

func escapePhrase(value string) string {
	return strings.ReplaceAll(value, `"`, `\"`)
}
