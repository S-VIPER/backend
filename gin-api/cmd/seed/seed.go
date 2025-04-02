package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/S-VIPER/backend/gin-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Get MongoDB URI from environment variable or use default
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://root:example@mongodb:27017"
		log.Println("MONGODB_URI environment variable is not set. Using default:", mongoURI)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(ctx)

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}
	log.Println("Connected to MongoDB!")

	// Get database reference
	db := client.Database("sviper")
	collection := db.Collection("tracks")

	// Drop existing tracks if any
	if err := collection.Drop(ctx); err != nil {
		log.Println("Warning: Failed to drop collection:", err)
	}

	// Sample tracks data
	tracks := []interface{}{
		domain.Track{
			ID:          "track001",
			Title:       "krovostok",
			Artist:      "arkadich",
			URL:         "http://192.168.0.164/tracks/arkadich/krovostok.mp3",
			AlbumTitle:  "",
			AlbumArtURL: "http://192.168.0.164/tracks/images/queen_night_at_the_opera.jpg",
			PreviewURL:  "http://192.168.0.164/tracks/previews/bohemian_rhapsody_preview.mp3",
			Genre:       []string{"Electro punk", "Rave"},
			Year:        2024,
		},
		domain.Track{
			ID:          "track002",
			Title:       "sosat",
			Artist:      "arkadich",
			URL:         "http://192.168.0.164/tracks/arkadich/sosat.mp3",
			AlbumTitle:  "",
			AlbumArtURL: "http://192.168.0.164/tracks/images/queen_night_at_the_opera.jpg",
			PreviewURL:  "http://192.168.0.164/tracks/previews/bohemian_rhapsody_preview.mp3",
			Genre:       []string{"Electro punk", "Rave"},
			Year:        2024,
		},
	}

	// Insert tracks
	result, err := collection.InsertMany(ctx, tracks)
	if err != nil {
		log.Fatal("Failed to insert tracks:", err)
	}

	log.Printf("Successfully inserted %d tracks with IDs: %v", len(result.InsertedIDs), result.InsertedIDs)

	// Print all inserted tracks for verification
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
