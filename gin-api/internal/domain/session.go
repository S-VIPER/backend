package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefreshTokenSession struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}
