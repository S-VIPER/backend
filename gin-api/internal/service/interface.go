package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type VerificationCodeGenerator interface {
	Generate() (plainCode string, codeHash string, err error)
	Compare(codeHash string, code string) bool
}

type VerificationCodeHasher interface {
	Hash(code string) (string, error)
	Compare(hash, code string) error
}

type EmailSender interface {
	SendRegistrationCode(
		ctx context.Context,
		email string,
		code string,
	) error
}

type AccessTokenClaims struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
}

type AccessTokenService interface {
	Generate(userID, sessionID uuid.UUID, now time.Time) (string, error)
	Parse(token string) (AccessTokenClaims, error)
	ExpiresIn() time.Duration
}

type RefreshTokenService interface {
	Generate() (plainToken string, tokenHash string, err error)
	Hash(token string) string
}
