package token_test

import (
	"testing"
	"time"

	"github.com/S-VIPER/backend/gin-api/internal/service/token"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestJWTServiceRoundTrip(t *testing.T) {
	service := token.NewJWTService(
		"test-secret",
		"sviper-api",
		"sviper-client",
		15*time.Minute,
	)

	userID := uuid.New()
	sessionID := uuid.New()
	now := time.Now()

	rawToken, err := service.Generate(userID, sessionID, now)
	require.NoError(t, err)
	require.NotEmpty(t, rawToken)

	claims, err := service.Parse(rawToken)
	require.NoError(t, err)

	require.Equal(t, userID, claims.UserID)
	require.Equal(t, sessionID, claims.SessionID)
}
