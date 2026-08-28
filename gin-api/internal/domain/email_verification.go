package domain

import (
	"time"

	"github.com/google/uuid"
)

type EmailVerification struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	CodeHash   string
	ExpiresAt  time.Time
	Attempts   int
	CreatedAt  time.Time
	VerifiedAt *time.Time
}
