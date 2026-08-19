package domain

type Track struct {
	ID          string
	Title       string
	Artist      string
	URL         string
	AlbumTitle  string
	AlbumArtURL string
	PreviewURL  string
	Genre       []string
	Year        int
}
