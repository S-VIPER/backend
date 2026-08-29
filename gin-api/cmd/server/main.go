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
	"github.com/S-VIPER/backend/gin-api/internal/repository/mongodb"
	"github.com/S-VIPER/backend/gin-api/internal/repository/postgres"
	"github.com/S-VIPER/backend/gin-api/internal/service/email"
	"github.com/S-VIPER/backend/gin-api/internal/service/password"
	"github.com/S-VIPER/backend/gin-api/internal/service/verification"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"

	"github.com/gin-gonic/gin"

	"github.com/jackc/pgx/v5/pgxpool"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	postgresURL := os.Getenv("POSTGRES_URI")
	if postgresURL == "" {
		log.Fatal("POSTGRES_URI environment variable is not set")
	}

	mongodbURI := requiredEnv("MONGODB_URI")

	// -------------------------------------------------------------------------
	// PostgreSQL
	// -------------------------------------------------------------------------

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

	// -------------------------------------------------------------------------
	// MongoDB
	// -------------------------------------------------------------------------

	mongoClient, err := mongodriver.Connect(
		ctx,
		options.Client().ApplyURI(mongodbURI),
	)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer disconnectCancel()

		if err := mongoClient.Disconnect(disconnectCtx); err != nil {
			log.Printf("failed to disconnect MongoDB: %v", err)
		}
	}()

	if err := mongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("failed to ping MongoDB: %v", err)
	}

	mongoDB := mongoClient.Database("sviper")

	log.Println("connected to MongoDB")

	trackRepo := mongodb.NewTrackRepository(mongoDB)
	playlistRepo := mongodb.NewPlaylistRepository(mongoDB)

	// -------------------------------------------------------------------------
	// Auth dependencies
	// -------------------------------------------------------------------------

	passwordHasher := password.NewArgon2Hasher()

	codeGenerator := verification.NewCodeGenerator()

	// SMTP/provider.
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM")

	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		log.Fatalf("invalid SMTP_PORT: %v", err)
	}

	emailSender := email.NewSMTPSender(email.SMTPConfig{
		Host:     smtpHost,
		Port:     smtpPort,
		Username: smtpUsername,
		Password: smtpPassword,
		From:     smtpFrom,
	})

	authUseCase := usecase.NewAuthUseCase(
		userRepo,
		verificationRepo,
		passwordHasher,
		codeGenerator,
		emailSender,
	)

	// -------------------------------------------------------------------------
	// Use cases
	// -------------------------------------------------------------------------

	trackUseCase := usecase.NewTrackUseCase(trackRepo)

	playlistUseCase := usecase.NewPlaylistUseCase(
		playlistRepo,
		trackRepo,
	)

	// -------------------------------------------------------------------------
	// Handlers
	// -------------------------------------------------------------------------

	authHandler := handler.NewAuthHandler(authUseCase)
	trackHandler := handler.NewTrackHandler(trackUseCase)
	playlistHandler := handler.NewPlaylistHandler(playlistUseCase)

	httpHandler := handler.NewHandler(
		authHandler,
		playlistHandler,
		trackHandler,
	)

	// -------------------------------------------------------------------------
	// HTTP
	// -------------------------------------------------------------------------

	router := gin.Default()

	strictHandler := api.NewStrictHandlerWithOptions(
		httpHandler,
		nil,
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

	// -------------------------------------------------------------------------
	// Protected routes
	// -------------------------------------------------------------------------
	jwtSecret := requiredEnv("JWT_SECRET")

	authMiddleware := middleware.NewJWTMiddleware(jwtSecret)

	// Пока используем middleware на уровне защищённых group.
	protected := router.Group("")
	protected.Use(authMiddleware.Handler())

	log.Println("server started on :8080")

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}

	_ = http.StatusOK
	_ = protected
}

// -----------------------------------------------------------------------------
// Env
// -----------------------------------------------------------------------------

func requiredEnv(name string) string {
	value := os.Getenv(name)

	if value == "" {
		log.Fatalf(
			"%s environment variable is not set",
			name,
		)
	}

	return value
}
