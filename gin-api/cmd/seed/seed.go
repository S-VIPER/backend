package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func main() {
	// --- Read env vars ---
	objectStorage := getEnv("OBJECT_STORAGE_SOCKET", "http://localhost:9000")

	mongoHost := getEnv("MONGO_HOST", "mongodb")
	mongoPort := getEnv("MONGO_PORT", "27017")
	mongoUser := getEnv("MONGO_USER", "root")
	mongoPass := getEnv("MONGO_PASS", "example")
	mongoDB := getEnv("MONGO_DB", "sviper")
	mongoCollection := getEnv("MONGO_COLLECTION", "tracks")

	mongoURI := getEnv("MONGODB_URI", fmt.Sprintf("mongodb://%s:%s@%s:%s", mongoUser, mongoPass, mongoHost, mongoPort))
	log.Println("Connecting to MongoDB at:", mongoURI)

	// --- Connect to MongoDB ---
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}
	log.Println("Connected to MongoDB!")

	db := client.Database(mongoDB)
	collection := db.Collection(mongoCollection)

	// --- Reset collection ---
	if err := collection.Drop(ctx); err != nil {
		log.Println("Warning: Failed to drop collection:", err)
	}

	// --- Seed sample data ---
	tracks := []interface{}{
		domain.Track{
			ID:          "track001",
			Title:       "krovostok",
			Artist:      "arkadich",
			URL:         fmt.Sprintf("%s/tracks/arkadich/krovostok.mp3", objectStorage),
			AlbumTitle:  "",
			AlbumArtURL: fmt.Sprintf("%s/tracks/images/queen_night_at_the_opera.jpg", objectStorage),
			PreviewURL:  fmt.Sprintf("%s/tracks/previews/bohemian_rhapsody_preview.mp3", objectStorage),
			Genre:       []string{"Electro punk", "Rave"},
			Year:        2024,
		},
		domain.Track{
			ID:          "track002",
			Title:       "sosat",
			Artist:      "arkadich",
			URL:         fmt.Sprintf("%s/tracks/arkadich/sosat.mp3", objectStorage),
			AlbumTitle:  "",
			AlbumArtURL: fmt.Sprintf("%s/tracks/images/queen_night_at_the_opera.jpg", objectStorage),
			PreviewURL:  fmt.Sprintf("%s/tracks/previews/bohemian_rhapsody_preview.mp3", objectStorage),
			Genre:       []string{"Electro punk", "Rave"},
			Year:        2024,
		},
	}

	result, err := collection.InsertMany(ctx, tracks)
	if err != nil {
		log.Fatal("Failed to insert tracks:", err)
	}

	log.Printf("Successfully inserted %d tracks with IDs: %v", len(result.InsertedIDs), result.InsertedIDs)

	// --- Verification ---
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal("Failed to retrieve tracks:", err)
	}
	defer cursor.Close(ctx)

	var retrievedTracks []domain.Track
	if err = cursor.All(ctx, &retrievedTracks); err != nil {
		log.Fatal("Failed to decode tracks:", err)
	}

	log.Println("Tracks in database:")
	for _, track := range retrievedTracks {
		log.Printf("- %s by %s (%d)", track.Title, track.Artist, track.Year)
	}
}
