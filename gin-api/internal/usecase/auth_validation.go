package usecase

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
)

var (
	errInvalidEmail    = errors.New("invalid email")
	errInvalidPassword = errors.New("invalid password")
)

const (
	minPasswordLength = 8
	maxPasswordLength = 128
)

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateEmail(email string) error {
	if email == "" {
		return domain.ErrInvalidEmail
	}

	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return domain.ErrInvalidEmail
	}

	if len(email) > 320 {
		return domain.ErrInvalidEmail
	}

	return nil
}

func validatePassword(password string) error {
	length := utf8.RuneCountInString(password)

	if length < 8 || length > 128 {
		return domain.ErrInvalidPassword
	}

	return nil
}
