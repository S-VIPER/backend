package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/api"
)

type contextKey string

const userIDContextKey contextKey = "userID"

type JWTMiddleware struct {
	secret []byte
}

func NewJWTMiddleware(secret string) *JWTMiddleware {
	return &JWTMiddleware{
		secret: []byte(secret),
	}
}

type Claims struct {
	jwt.RegisteredClaims
}

func (m *JWTMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			writeUnauthorized(
				c,
				"missing or invalid Authorization header",
			)
			return
		}

		token, err := jwt.ParseWithClaims(
			authHeader,
			&Claims{},
			func(token *jwt.Token) (any, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, jwt.ErrSignatureInvalid
				}

				return m.secret, nil
			},
			jwt.WithValidMethods([]string{
				jwt.SigningMethodHS256.Alg(),
			}),
		)

		if err != nil {
			writeUnauthorized(c, "invalid or expired token")
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok || !token.Valid {
			writeUnauthorized(c, "invalid token")
			return
		}

		if claims.Subject == "" {
			writeUnauthorized(c, "token subject is missing")
			return
		}

		if claims.ExpiresAt == nil {
			writeUnauthorized(c, "token expiration is missing")
			return
		}

		ctx := context.WithValue(
			c.Request.Context(),
			userIDContextKey,
			claims.Subject,
		)

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
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

func UserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)

	return userID, ok
}

func writeUnauthorized(
	c *gin.Context,
	message string,
) {
	response := api.ErrorResponse{}

	response.Error.Code = "UNAUTHORIZED"
	response.Error.Message = message

	c.AbortWithStatusJSON(
		http.StatusUnauthorized,
		response,
	)
}
