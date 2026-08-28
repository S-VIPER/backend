package domain

import "errors"

var (
	ErrInvalidTrack       = errors.New("invalid track")
	ErrInvalidTrackID     = errors.New("invalid track id")
	ErrInvalidTrackTitle  = errors.New("invalid track title")
	ErrInvalidTrackArtist = errors.New("invalid track artist")
	ErrInvalidTrackURL    = errors.New("invalid track url")
	ErrInvalidTrackYear   = errors.New("invalid track year")
	ErrTrackNotFound      = errors.New("track not found")
	ErrTrackAlreadyExists = errors.New("track already exists")
)

var (
	ErrInvalidPlaylist       = errors.New("invalid playlist")
	ErrInvalidPlaylistID     = errors.New("invalid playlist id")
	ErrInvalidPlaylistName   = errors.New("invalid playlist name")
	ErrPlaylistNotFound      = errors.New("playlist not found")
	ErrPlaylistAlreadyExists = errors.New("playlist already exists")
)

var (
	ErrUserAlreadyExists           = errors.New("user already exists")
	ErrUserNotFound                = errors.New("user not found")
	ErrVerificationNotFound        = errors.New("verification not found")
	ErrVerificationExpired         = errors.New("verification expired")
	ErrVerificationInvalid         = errors.New("verification invalid")
	ErrVerificationRateLimited     = errors.New("verification rate limited")
	ErrVerificationTooManyAttempts = errors.New("too many verification attempts")
	ErrInvalidEmail                = errors.New("invalid email")
	ErrInvalidPassword             = errors.New("invalid password")
)
