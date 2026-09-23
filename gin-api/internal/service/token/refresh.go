package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/S-VIPER/backend/gin-api/internal/service"
)

type RefreshTokenService struct{}

func NewRefreshTokenService() *RefreshTokenService {
	return &RefreshTokenService{}
}

var _ service.RefreshTokenService = (*RefreshTokenService)(nil)

func (s *RefreshTokenService) Generate() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}

	plain := base64.RawURLEncoding.EncodeToString(buf)
	return plain, s.Hash(plain), nil
}

func (s *RefreshTokenService) Hash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
