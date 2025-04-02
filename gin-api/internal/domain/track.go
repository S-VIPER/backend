package domain

type Track struct {
	ID          string   `json:"id" gorm:"primaryKey"`
	Title       string   `json:"title"`
	Artist      string   `json:"artist"`
	URL         string   `json:"url"`
	AlbumTitle  string   `json:"albumTitle"`
	AlbumArtURL string   `json:"albumArtUrl"`
	PreviewURL  string   `json:"previewUrl"`
	Genre       []string `json:"genre" gorm:"serializer:json"`
	Year        int      `json:"year"`
}
