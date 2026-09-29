package domain

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
	MBID                      string
	Title                     string
	Artist                    string
	AlbumTitle                string
	AlbumArtURL               string
	Genre                     []string
	Year                      int
	MusicBrainzReleaseID      string
	MusicBrainzReleaseGroupID string
}
