package token

import (
	"fmt"
	"time"

	"github.com/S-VIPER/backend/gin-api/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
}

func NewJWTService(secret, issuer, audience string, ttl time.Duration) *JWTService {
	return &JWTService{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		ttl:      ttl,
	}
}

type claims struct {
	jwt.RegisteredClaims
}

var _ service.AccessTokenService = (*JWTService)(nil)

func (s *JWTService) Generate(userID, sessionID uuid.UUID, now time.Time) (string, error) {
	expiresAt := now.Add(s.ttl)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			Subject:   userID.String(),
			ID:        sessionID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	})

	return token.SignedString(s.secret)
}

func (s *JWTService) ExpiresIn() time.Duration {
	return s.ttl
}

func (s *JWTService) Parse(rawToken string) (service.AccessTokenClaims, error) {
	parsed, err := jwt.ParseWithClaims(
		rawToken,
		&claims{},
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return s.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
	)
	if err != nil {
		return service.AccessTokenClaims{}, fmt.Errorf("parse access token: %w", err)
	}

	c, ok := parsed.Claims.(*claims)
	if !ok || !parsed.Valid {
		return service.AccessTokenClaims{}, fmt.Errorf("invalid access token")
	}

	userID, err := uuid.Parse(c.Subject)
	if err != nil {
		return service.AccessTokenClaims{}, fmt.Errorf("parse user id: %w", err)
	}

	sessionID, err := uuid.Parse(c.ID)
	if err != nil {
		return service.AccessTokenClaims{}, fmt.Errorf("parse session id: %w", err)
	}

	return service.AccessTokenClaims{
		UserID:    userID,
		SessionID: sessionID,
	}, nil
}
