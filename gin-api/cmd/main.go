package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/S-VIPER/backend/gin-api/internal/delivery/http"
	"github.com/S-VIPER/backend/gin-api/internal/repository"
	"github.com/S-VIPER/backend/gin-api/internal/usecase"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Получаем URI MongoDB из переменной окружения
	mongodbURI := os.Getenv("MONGODB_URI")
	if mongodbURI == "" {
		log.Fatal("MONGODB_URI environment variable is not set")
	}

	// Подключение к MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("sviper")

	// Инициализация слоев
	trackRepo := repository.NewTrackRepository(db)
	playlistRepo := repository.NewPlaylistRepository(db)
	trackUseCase := usecase.NewTrackUseCase(trackRepo)
	playlistUseCase := usecase.NewPlaylistUseCase(playlistRepo, trackRepo)
	trackHandler := http.NewTrackHandler(trackUseCase)
	playlistHandler := http.NewPlaylistHandler(playlistUseCase)

	// Настройка Gin
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Маршруты для треков
	r.POST("/tracks", trackHandler.CreateTrack)
	r.GET("/tracks", trackHandler.GetAllTracks)
	r.GET("/tracks/:id", trackHandler.GetTrackByID)
	r.PUT("/tracks/:id", trackHandler.UpdateTrack)
	r.DELETE("/tracks/:id", trackHandler.DeleteTrack)

	// Маршруты для плейлистов
	r.GET("/playlists", playlistHandler.GetAllPlaylists)
	r.POST("/playlists", playlistHandler.CreatePlaylist)
	r.GET("/playlists/:id", playlistHandler.GetPlaylistByID)
	r.PUT("/playlists/:id", playlistHandler.UpdatePlaylist)
	r.DELETE("/playlists/:id", playlistHandler.DeletePlaylist)
	r.POST("/playlists/:id/tracks", playlistHandler.AddTrackToPlaylist)
	r.DELETE("/playlists/:id/tracks/:trackId", playlistHandler.RemoveTrackFromPlaylist)

	// Запуск сервера
	r.Run(":8080")
}
