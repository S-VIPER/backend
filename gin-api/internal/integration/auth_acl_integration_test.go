//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/api"
	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/handler"
	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/middleware"
	"github.com/S-VIPER/backend/gin-api/internal/domain"
	repo_postgres "github.com/S-VIPER/backend/gin-api/internal/repository/postgres"
	"github.com/S-VIPER/backend/gin-api/internal/service/password"
	"github.com/S-VIPER/backend/gin-api/internal/service/token"
	"github.com/S-VIPER/backend/gin-api/internal/service/verification"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	migratePgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type noopEmailSender struct{}

func (noopEmailSender) SendRegistrationCode(context.Context, string, string) error {
	return nil
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type playlistResponse struct {
	Data struct {
		ID         string   `json:"id"`
		Name       string   `json:"name"`
		Tracks     []string `json:"tracks"`
		Visibility string   `json:"visibility"`
	} `json:"data"`
}

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func TestAuthACLIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Docker")
	}

	ctx := context.Background()

	container, err := postgres.Run(
		ctx,
		"postgres:18.6-alpine",
		postgres.WithDatabase("sviper-integration-test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").
				WithStartupTimeout(40*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	connectionString, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connectionString)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	require.NoError(t, pool.Ping(ctx))
	runMigrations(t, pool)

	userRepo := repo_postgres.NewUserRepository(pool)
	verificationRepo := repo_postgres.NewEmailVerificationRepository(pool)
	refreshTokenRepo := repo_postgres.NewRefreshTokenRepository(pool)
	playlistRepo := repo_postgres.NewPlaylistRepository(pool)

	passwordHasher := password.NewArgon2Hasher()
	codeGenerator := verification.NewCodeGenerator()
	accessTokenService := token.NewJWTService(
		"integration-test-secret",
		"sviper-api",
		"sviper-client",
		15*time.Minute,
	)
	refreshTokenService := token.NewRefreshTokenService()

	authUseCase := usecase.NewAuthUseCase(
		userRepo,
		verificationRepo,
		passwordHasher,
		codeGenerator,
		noopEmailSender{},
		refreshTokenRepo,
		accessTokenService,
		refreshTokenService,
	)

	playlistUseCase := usecase.NewPlaylistUseCase(
		playlistRepo,
		nil,
	)

	authHandler := handler.NewAuthHandler(authUseCase)
	playlistHandler := handler.NewPlaylistHandler(playlistUseCase)

	// TrackHandler is intentionally nil: this integration scenario exercises
	// auth + playlist ACL only and never invokes track endpoints.
	httpHandler := handler.NewHandler(
		authHandler,
		playlistHandler,
		nil,
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	jwtMiddleware := middleware.NewJWTMiddleware(accessTokenService)
	strictHandler := api.NewStrictHandlerWithOptions(
		httpHandler,
		[]api.StrictMiddlewareFunc{jwtMiddleware.StrictMiddleware},
		api.StrictGinServerOptions{
			RequestErrorHandlerFunc:  handler.HandleRequestError,
			HandlerErrorFunc:         handler.HandleHandlerError,
			ResponseErrorHandlerFunc: handler.HandleResponseError,
		},
	)

	api.RegisterHandlersWithOptions(
		router,
		strictHandler,
		api.GinServerOptions{},
	)

	alice := seedVerifiedUser(
		t,
		ctx,
		userRepo,
		passwordHasher,
		"alice@example.com",
		"alice-password",
	)

	_ = seedVerifiedUser(
		t,
		ctx,
		userRepo,
		passwordHasher,
		"bob@example.com",
		"bob-password",
	)

	// -------------------------------------------------------------------------
	// Login -> access JWT
	// -------------------------------------------------------------------------

	aliceAuth := login(t, router, "alice@example.com", "alice-password")

	require.NotEmpty(t, aliceAuth.AccessToken)
	require.NotEmpty(t, aliceAuth.RefreshToken)
	require.Equal(t, "Bearer", aliceAuth.TokenType)
	require.Positive(t, aliceAuth.ExpiresIn)

	claims, err := accessTokenService.Parse(aliceAuth.AccessToken)
	require.NoError(
		t,
		err,
		"login returned access token that cannot be parsed",
	)
	require.Equal(t, alice.ID, claims.UserID)
	require.NotEqual(t, uuid.Nil, claims.SessionID)

	// Protected route must reject requests without access JWT.
	unauthorizedCreate := doJSON(
		t,
		router,
		http.MethodPost,
		"/playlists",
		map[string]any{"name": "no token"},
		"",
	)
	require.Equal(t, http.StatusUnauthorized, unauthorizedCreate.Code)
	assertErrorCode(t, unauthorizedCreate, "UNAUTHORIZED")

	// -------------------------------------------------------------------------
	// Alice creates private + public playlists.
	// -------------------------------------------------------------------------

	private := createPlaylist(t, router, aliceAuth.AccessToken, map[string]any{
		"name":       "Alice Private",
		"visibility": "private",
	})
	public := createPlaylist(t, router, aliceAuth.AccessToken, map[string]any{
		"name":       "Alice Public",
		"visibility": "public",
	})

	require.Equal(t, "private", private.Data.Visibility)
	require.Equal(t, "public", public.Data.Visibility)

	privateID := private.Data.ID
	publicID := public.Data.ID
	require.NotEmpty(t, privateID)
	require.NotEmpty(t, publicID)

	// -------------------------------------------------------------------------
	// Public/private ACL
	// -------------------------------------------------------------------------

	// Anonymous user cannot read a private playlist.
	anonymousPrivate := doJSON(
		t,
		router,
		http.MethodGet,
		"/playlists/"+privateID,
		nil,
		"",
	)
	require.Equal(t, http.StatusNotFound, anonymousPrivate.Code)
	assertErrorCode(t, anonymousPrivate, "PLAYLIST_NOT_FOUND")

	// Anonymous user can read a public playlist.
	anonymousPublic := doJSON(
		t,
		router,
		http.MethodGet,
		"/playlists/"+publicID,
		nil,
		"",
	)
	require.Equal(t, http.StatusOK, anonymousPublic.Code)
	assertPlaylist(t, anonymousPublic, publicID, "public")

	bobAuth := login(t, router, "bob@example.com", "bob-password")

	// Another authenticated user still cannot read Alice's private playlist.
	bobPrivate := doJSON(
		t,
		router,
		http.MethodGet,
		"/playlists/"+privateID,
		nil,
		bobAuth.AccessToken,
	)
	require.Equal(t, http.StatusNotFound, bobPrivate.Code)
	assertErrorCode(t, bobPrivate, "PLAYLIST_NOT_FOUND")

	// Another authenticated user can read Alice's public playlist.
	bobPublic := doJSON(
		t,
		router,
		http.MethodGet,
		"/playlists/"+publicID,
		nil,
		bobAuth.AccessToken,
	)
	require.Equal(t, http.StatusOK, bobPublic.Code)
	assertPlaylist(t, bobPublic, publicID, "public")

	// Owner can read their private playlist.
	alicePrivate := doJSON(
		t,
		router,
		http.MethodGet,
		"/playlists/"+privateID,
		nil,
		aliceAuth.AccessToken,
	)
	require.Equal(t, http.StatusOK, alicePrivate.Code)
	assertPlaylist(t, alicePrivate, privateID, "private")

	// Another user cannot modify or delete Alice's playlist, even with a valid JWT.
	bobUpdate := doJSON(
		t,
		router,
		http.MethodPut,
		"/playlists/"+privateID,
		map[string]any{"name": "hacked", "visibility": "private"},
		bobAuth.AccessToken,
	)
	require.Equal(t, http.StatusNotFound, bobUpdate.Code)
	assertErrorCode(t, bobUpdate, "PLAYLIST_NOT_FOUND")

	bobDelete := doJSON(
		t,
		router,
		http.MethodDelete,
		"/playlists/"+publicID,
		nil,
		bobAuth.AccessToken,
	)
	require.Equal(t, http.StatusNotFound, bobDelete.Code)
	assertErrorCode(t, bobDelete, "PLAYLIST_NOT_FOUND")

	// -------------------------------------------------------------------------
	// Refresh rotation: old refresh token becomes unusable after rotation.
	// -------------------------------------------------------------------------

	refreshed := doJSON(
		t,
		router,
		http.MethodPost,
		"/auth/refresh",
		map[string]any{"refresh_token": aliceAuth.RefreshToken},
		"",
	)
	require.Equal(t, http.StatusOK, refreshed.Code, refreshed.Body.String())

	var refreshedAuth authResponse
	require.NoError(t, json.Unmarshal(refreshed.Body.Bytes(), &refreshedAuth))
	require.NotEmpty(t, refreshedAuth.AccessToken)
	require.NotEmpty(t, refreshedAuth.RefreshToken)
	require.NotEqual(t, aliceAuth.RefreshToken, refreshedAuth.RefreshToken)

	oldRefreshAfterRotation := doJSON(
		t,
		router,
		http.MethodPost,
		"/auth/refresh",
		map[string]any{"refresh_token": aliceAuth.RefreshToken},
		"",
	)
	require.Equal(t, http.StatusUnauthorized, oldRefreshAfterRotation.Code)
	assertErrorCode(t, oldRefreshAfterRotation, "INVALID_CREDENTIALS")

	// -------------------------------------------------------------------------
	// Logout -> refresh must fail because the rotated refresh session is revoked.
	// -------------------------------------------------------------------------

	logout := doJSON(
		t,
		router,
		http.MethodPost,
		"/auth/logout",
		map[string]any{"refresh_token": refreshedAuth.RefreshToken},
		"",
	)
	require.Equal(t, http.StatusNoContent, logout.Code)

	refreshAfterLogout := doJSON(
		t,
		router,
		http.MethodPost,
		"/auth/refresh",
		map[string]any{"refresh_token": refreshedAuth.RefreshToken},
		"",
	)
	require.Equal(t, http.StatusUnauthorized, refreshAfterLogout.Code)
	assertErrorCode(t, refreshAfterLogout, "INVALID_CREDENTIALS")

	// Logout revokes the refresh session, not the already issued access JWT.
	// With the current design this access token remains valid until its 15m TTL.
	stillValidAccessToken := doJSON(
		t,
		router,
		http.MethodGet,
		"/playlists/"+privateID,
		nil,
		aliceAuth.AccessToken,
	)
	require.Equal(t, http.StatusOK, stillValidAccessToken.Code)
	assertPlaylist(t, stillValidAccessToken, privateID, "private")

}

func runMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	db := stdlib.OpenDB(*pool.Config().ConnConfig)
	defer db.Close()

	driver, err := migratePgx.WithInstance(db, &migratePgx.Config{})
	require.NoError(t, err)

	migrationsPath, err := filepath.Abs(filepath.Join("..", "..", "migrations"))
	require.NoError(t, err)

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"pgx",
		driver,
	)
	require.NoError(t, err)

	err = m.Up()
	require.True(t, errors.Is(err, migrate.ErrNoChange) || err == nil, "failed to run migrations: %v", err)
}

func seedVerifiedUser(
	t *testing.T,
	ctx context.Context,
	repo usecase.UserRepositoryInterface,
	hasher interface {
		Hash(string) (string, error)
	},
	email string,
	plainPassword string,
) *domain.User {
	t.Helper()

	hash, err := hasher.Hash(plainPassword)
	require.NoError(t, err)

	user := &domain.User{
		Email:         email,
		PasswordHash:  hash,
		EmailVerified: false,
	}
	require.NoError(t, repo.Create(ctx, user))
	require.NoError(t, repo.Activate(ctx, user.ID))

	user.EmailVerified = true
	return user
}

func login(
	t *testing.T,
	router http.Handler,
	email string,
	password string,
) authResponse {
	t.Helper()

	response := doJSON(
		t,
		router,
		http.MethodPost,
		"/auth/login",
		map[string]any{
			"email":    email,
			"password": password,
		},
		"",
	)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())

	var result authResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
	return result
}

func createPlaylist(
	t *testing.T,
	router http.Handler,
	accessToken string,
	body map[string]any,
) playlistResponse {
	t.Helper()

	response := doJSON(
		t,
		router,
		http.MethodPost,
		"/playlists",
		body,
		accessToken,
	)

	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())

	var result playlistResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
	return result
}

func doJSON(
	t *testing.T,
	httpHandler http.Handler,
	method string,
	path string,
	body any,
	accessToken string,
) *httptest.ResponseRecorder {
	t.Helper()

	var requestBody io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		require.NoError(t, err)
		requestBody = bytes.NewReader(payload)
	}

	request := httptest.NewRequest(method, path, requestBody)
	request.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)

		require.Equal(
			t,
			"Bearer "+accessToken,
			request.Header.Get("Authorization"),
		)
	}

	response := httptest.NewRecorder()
	httpHandler.ServeHTTP(response, request)
	return response
}

func assertErrorCode(t *testing.T, response *httptest.ResponseRecorder, expected string) {
	t.Helper()

	var body errorResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, expected, body.Error.Code)
}

func assertPlaylist(
	t *testing.T,
	response *httptest.ResponseRecorder,
	expectedID string,
	expectedVisibility string,
) {
	t.Helper()

	var body playlistResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, expectedID, body.Data.ID)
	require.Equal(t, expectedVisibility, body.Data.Visibility)
}
