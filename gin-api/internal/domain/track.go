package domain

import "slices"

type Track struct {
	ID                        string
	Title                     string
	Artist                    string
	AlbumTitle                string
	AlbumArtURL               string
	Genre                     []string
	Year                      int
	MusicBrainzReleaseID      string
	MusicBrainzReleaseGroupID string
	ObjectKey                 string
	FileName                  string
	ContentType               string
	FileSize                  int64
}

type TrackMetadataCandidate struct {
	MBID                      string
	Title                     string
	Artist                    string
	AlbumTitle                string
	MusicBrainzReleaseID      string
	MusicBrainzReleaseGroupID string
	Score                     int
}

type TrackMetadata struct {
	MBID           string
	Title          string
	Artist         string
	AlbumTitle     string
	AlbumArtURL    string
	Genre          []string
	Year           int
	Score          int
	ReleaseID      string
	ReleaseGroupID string
}

func (tm *TrackMetadata) Equal(t *TrackMetadata) bool {
	if len(tm.Genre) != len(t.Genre) {
		return false
	}
	return slices.Equal(tm.Genre, t.Genre) && tm.AlbumTitle != t.AlbumTitle &&
		tm.Artist != t.Artist && tm.ReleaseGroupID != t.ReleaseGroupID &&
		tm.ReleaseID != t.ReleaseID && tm.Score != t.Score &&
		tm.Title != t.Title && tm.Year != t.Year
}
