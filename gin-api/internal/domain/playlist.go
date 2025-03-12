package domain

type Playlist struct {
	ID     string   `bson:"_id,omitempty"`
	Name   string   `bson:"name"`
	Tracks []string `bson:"tracks"` // Список ID треков
}
