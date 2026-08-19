package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/api"
	"github.com/S-VIPER/backend/gin-api/internal/delivery/http/handler"
	"github.com/S-VIPER/backend/gin-api/internal/repository"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	mongodbURI := os.Getenv("MONGODB_URI")
	if mongodbURI == "" {
		log.Fatal("MONGODB_URI environment variable is not set")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	client, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI(mongodbURI),
	)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("failed to ping MongoDB: %v", err)
	}

	db := client.Database("sviper")

	trackRepo := repository.NewTrackRepository(db)
	playlistRepo := repository.NewPlaylistRepository(db)

	trackUseCase := usecase.NewTrackUseCase(trackRepo)
	playlistUseCase := usecase.NewPlaylistUseCase(
		playlistRepo,
		trackRepo,
	)

	trackHandler := handler.NewTrackHandler(trackUseCase)
	playlistHandler := handler.NewPlaylistHandler(playlistUseCase)

	httpHandler := handler.NewHandler(
		playlistHandler,
		trackHandler,
	)

	r := gin.Default()

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
		r,
		strictHandler,
		api.GinServerOptions{
			BaseURL: "/api/v1",
		},
	)

	log.Println("server started on :8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}

	_ = http.StatusOK
}
