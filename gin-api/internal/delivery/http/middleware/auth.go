package middleware

import (
	"context"
	"strings"

	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/api"
	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const userIDContextKey contextKey = "userID"

type JWTMiddleware struct {
	accessTokenService service.AccessTokenService
}

func NewJWTMiddleware(accessTokenService service.AccessTokenService) *JWTMiddleware {
	return &JWTMiddleware{accessTokenService: accessTokenService}
}

func (m *JWTMiddleware) StrictMiddleware(
	next api.StrictHandlerFunc,
	operationID string,
) api.StrictHandlerFunc {
	return func(c *gin.Context, request any) (any, error) {
		if operationID == "GetPlaylistByID" {
			if err := m.authenticateOptional(c); err != nil {
				return nil, err
			}
		} else if requiresAuthentication(operationID) {
			if err := m.authenticate(c); err != nil {
				return nil, err
			}
		}

		return next(c, request)
	}
}

func requiresAuthentication(operationID string) bool {
	switch operationID {
	case "CreateTrack",
		"UpdateTrack",
		"DeleteTrack",
		"CreatePlaylist",
		"UpdatePlaylist",
		"DeletePlaylist",
		"AddTrackToPlaylist",
		"RemoveTrackFromPlaylist":
		return true
	default:
		return false
	}
}

func (m *JWTMiddleware) authenticate(c *gin.Context) error {
	rawToken, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		return domain.ErrUnauthorized
	}

	claims, err := m.accessTokenService.Parse(rawToken)
	if err != nil {
		return domain.ErrUnauthorized
	}

	setUserID(c, claims.UserID)
	return nil
}

func (m *JWTMiddleware) authenticateOptional(c *gin.Context) error {
	header := strings.TrimSpace(c.GetHeader("Authorization"))
	if header == "" {
		return nil
	}

	return m.authenticate(c)
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 {
		return "", false
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	if parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

func setUserID(c *gin.Context, id uuid.UUID) {
	// The strict OpenAPI handler passes *gin.Context to handlers as
	// context.Context. Gin's Value() does not necessarily delegate to
	// Request.Context(), so keep the value in Gin's context as well.
	c.Set(string(userIDContextKey), id)

	// Also keep it in the standard request context for code that receives
	// c.Request.Context() downstream.
	ctx := context.WithValue(
		c.Request.Context(),
		userIDContextKey,
		id,
	)
	c.Request = c.Request.WithContext(ctx)
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	// Strict handlers receive *gin.Context as context.Context.
	if c, ok := ctx.(*gin.Context); ok {
		if value, exists := c.Get(string(userIDContextKey)); exists {
			id, ok := value.(uuid.UUID)
			if ok {
				return id, true
			}
		}
	}

	// Fallback for ordinary context.Context callers.
	userID, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return userID, ok
}
