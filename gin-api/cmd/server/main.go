package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/api"
	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/handler"
	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/middleware"
	"github.com/S-VIPER/backend/gin-api/internal/repository/musicbrainz"
	"github.com/S-VIPER/backend/gin-api/internal/repository/postgres"
	"github.com/S-VIPER/backend/gin-api/internal/repository/rustfs"
	"github.com/S-VIPER/backend/gin-api/internal/service/email"
	"github.com/S-VIPER/backend/gin-api/internal/service/password"
	"github.com/S-VIPER/backend/gin-api/internal/service/token"
	"github.com/S-VIPER/backend/gin-api/internal/service/verification"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	postgresURL := requiredEnv("POSTGRES_URI")
	jwtSecret := requiredEnv("JWT_SECRET")

	pgPool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		log.Fatalf("failed to create PostgreSQL pool: %v", err)
	}
	defer pgPool.Close()

	if err := pgPool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping PostgreSQL: %v", err)
	}
	log.Println("connected to PostgreSQL")

	userRepo := postgres.NewUserRepository(pgPool)
	verificationRepo := postgres.NewEmailVerificationRepository(pgPool)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(pgPool)
	playlistRepo := postgres.NewPlaylistRepository(pgPool)
	trackRepo := postgres.NewTrackRepository(pgPool)

	trackStorage, err := rustfs.NewTrackStorage(rustfs.Config{
		Endpoint:       requiredEnv("RUSTFS_ENDPOINT"),
		PublicEndpoint: requiredEnv("RUSTFS_PUBLIC_ENDPOINT"),
		AccessKey:      requiredEnv("RUSTFS_ACCESS_KEY"),
		SecretKey:      requiredEnv("RUSTFS_SECRET_KEY"),
		Bucket:         requiredEnv("RUSTFS_BUCKET"),
		Region:         os.Getenv("RUSTFS_REGION"),
	})
	if err != nil {
		log.Fatalf("failed to configure RustFS: %v", err)
	}
	if err := trackStorage.EnsureBucket(ctx); err != nil {
		log.Fatalf("failed to initialize RustFS bucket: %v", err)
	}
	log.Println("connected to RustFS")

	musicBrainzClient, err := musicbrainz.NewClient(musicbrainz.Config{
		BaseURL:     os.Getenv("MUSICBRAINZ_BASE_URL"),
		UserAgent:   requiredEnv("MUSICBRAINZ_USER_AGENT"),
		MinInterval: time.Second,
	})
	if err != nil {
		log.Fatalf("failed to configure MusicBrainz client: %v", err)
	}

	passwordHasher := password.NewArgon2Hasher()
	codeGenerator := verification.NewCodeGenerator()
	accessTokenService := token.NewJWTService(
		jwtSecret,
		"sviper-api",
		"sviper-client",
		15*time.Minute,
	)
	refreshTokenService := token.NewRefreshTokenService()

	smtpPort, err := strconv.Atoi(requiredEnv("SMTP_PORT"))
	if err != nil {
		log.Fatalf("invalid SMTP_PORT: %v", err)
	}

	emailSender := email.NewSMTPSender(email.SMTPConfig{
		Host:     requiredEnv("SMTP_HOST"),
		Port:     smtpPort,
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     requiredEnv("SMTP_FROM"),
	})

	authUseCase := usecase.NewAuthUseCase(
		userRepo,
		verificationRepo,
		passwordHasher,
		codeGenerator,
		emailSender,
		refreshTokenRepo,
		accessTokenService,
		refreshTokenService,
	)

	trackMaxUploadSize := int64(usecase.DefaultTrackMaxUploadSize)
	if raw := os.Getenv("MAX_TRACK_UPLOAD_SIZE"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			log.Fatalf("invalid MAX_TRACK_UPLOAD_SIZE: %q", raw)
		}
		trackMaxUploadSize = parsed
	}

	trackContentURLTTL := usecase.DefaultTrackContentURLTTL
	if raw := os.Getenv("TRACK_CONTENT_URL_TTL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed <= 0 {
			log.Fatalf("invalid TRACK_CONTENT_URL_TTL: %q", raw)
		}
		trackContentURLTTL = parsed
	}

	trackUseCase := usecase.NewTrackUseCase(
		trackRepo,
		trackStorage,
		musicBrainzClient,
	).WithMaxUploadSize(trackMaxUploadSize).WithContentURLTTL(trackContentURLTTL)

	playlistUseCase := usecase.NewPlaylistUseCase(
		playlistRepo,
		trackRepo,
	)

	authHandler := handler.NewAuthHandler(authUseCase)
	trackHandler := handler.NewTrackHandler(trackUseCase).WithMaxUploadSize(trackMaxUploadSize)
	playlistHandler := handler.NewPlaylistHandler(playlistUseCase)

	httpHandler := handler.NewHandler(
		authHandler,
		playlistHandler,
		trackHandler,
	)

	router := gin.Default()
	router.MaxMultipartMemory = 32 << 20

	authMiddleware := middleware.NewJWTMiddleware(accessTokenService)

	strictHandler := api.NewStrictHandlerWithOptions(
		httpHandler,
		[]api.StrictMiddlewareFunc{authMiddleware.StrictMiddleware},
		api.StrictGinServerOptions{
			RequestErrorHandlerFunc:  handler.HandleRequestError,
			HandlerErrorFunc:         handler.HandleHandlerError,
			ResponseErrorHandlerFunc: handler.HandleResponseError,
		},
	)

	api.RegisterHandlersWithOptions(
		router,
		strictHandler,
		api.GinServerOptions{
			BaseURL: "",
		},
	)

	log.Println("server started on :8080")

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Minute,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func requiredEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s environment variable is not set", name)
	}
	return value
}
