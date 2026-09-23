package domain

import "github.com/google/uuid"

type PlaylistVisibility string

const (
	PlaylistVisibilityPrivate PlaylistVisibility = "private"
	PlaylistVisibilityPublic  PlaylistVisibility = "public"
)

type Playlist struct {
	ID         string
	OwnerID    uuid.UUID
	Name       string
	Tracks     []string
	Visibility PlaylistVisibility
}
