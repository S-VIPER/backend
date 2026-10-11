package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/api"
	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"github.com/S-VIPER/backend/gin-api/internal/service"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const bearerAuthScheme = "bearerAuth"

type contextKey string

const userIDContextKey contextKey = "userID"

type securityPolicy uint8

const (
	policyPublic securityPolicy = iota
	policyRequired
	policyOptional
)

type JWTMiddleware struct {
	accessTokenService service.AccessTokenService
	policies           map[string]securityPolicy
}

func NewJWTMiddleware(
	accessTokenService service.AccessTokenService,
	spec *openapi3.T,
) (*JWTMiddleware, error) {
	if spec == nil {
		return nil, errors.New("OpenAPI specification is nil")
	}

	if spec.Paths == nil {
		return nil, errors.New("OpenAPI paths are not defined")
	}

	if len(spec.Security) == 0 {
		return nil, errors.New(
			"global OpenAPI security is not defined or empty",
		)
	}

	policies := make(map[string]securityPolicy)

	for _, pathItem := range spec.Paths.Map() {
		for _, operation := range pathItem.Operations() {
			if operation == nil {
				continue
			}

			if operation.OperationID == "" {
				return nil, errors.New(
					"OpenAPI operation has no operationId",
				)
			}

			if _, exists := policies[operation.OperationID]; exists {
				return nil, fmt.Errorf(
					"duplicate OpenAPI operationId %q",
					operation.OperationID,
				)
			}

			requirements := spec.Security

			if operation.Security != nil {
				requirements = *operation.Security
			}

			policy, err := policyFromSecurity(requirements)

			if err != nil {
				return nil, fmt.Errorf(
					"operation %q: %w",
					operation.OperationID,
					err,
				)
			}

			policies[operation.OperationID] = policy
		}
	}

	return &JWTMiddleware{
		accessTokenService: accessTokenService,
		policies:           policies,
	}, nil
}

func (m *JWTMiddleware) StrictMiddleware(
	next api.StrictHandlerFunc,
	operationID string,
) api.StrictHandlerFunc {
	return func(c *gin.Context, request any) (any, error) {
		policy, exists := m.policies[operationID]
		if !exists {
			return nil, fmt.Errorf(
				"security policy not found for operation %q",
				operationID,
			)
		}

		switch policy {
		case policyPublic:

		case policyRequired:
			if err := m.authenticate(c); err != nil {
				return nil, err
			}

		case policyOptional:
			if err := m.authenticateOptional(c); err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf(
				"unknown security policy for operation %q",
				operationID,
			)
		}

		return next(c, request)
	}
}

func (m *JWTMiddleware) RequireAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := m.authenticate(c); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "unauthorized",
					"message": "authentication required",
				},
			})
			return
		}

		c.Next()
	}
}

func policyFromSecurity(
	requirements openapi3.SecurityRequirements,
) (securityPolicy, error) {
	if len(requirements) == 0 {
		return policyPublic, nil
	}

	hasBearer := false
	allowsAnonymous := false

	for _, requirement := range requirements {
		if len(requirement) == 0 {
			allowsAnonymous = true
			continue
		}

		for schemeName := range requirement {
			if schemeName != bearerAuthScheme {
				return policyPublic, fmt.Errorf(
					"unsupported security scheme %q; only %q is supported",
					schemeName,
					bearerAuthScheme,
				)
			}

			hasBearer = true
		}
	}

	if !hasBearer {
		return policyPublic, nil
	}

	if allowsAnonymous {
		return policyOptional, nil
	}

	return policyRequired, nil
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
	c.Set(string(userIDContextKey), id)

	ctx := context.WithValue(
		c.Request.Context(),
		userIDContextKey,
		id,
	)
	c.Request = c.Request.WithContext(ctx)
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	if c, ok := ctx.(*gin.Context); ok {
		if value, exists := c.Get(string(userIDContextKey)); exists {
			id, ok := value.(uuid.UUID)
			if ok {
				return id, true
			}
		}
	}

	userID, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return userID, ok
}
