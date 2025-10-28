package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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

	// --- Tracks list ---
	files := []string{
		"ANYACT - Grajdane.mp3",
		"Any Act - Lakad Matatag.mp3",
		"Artv - Alcohol Kvlt.mp3",
		"BERSERKRR - Karma.mp3",
		"Hermeth - Lone Wolf.mp3",
		"Locked Club - Chevo Chevo.mp3",
		"LOCKED CLUB - heart for whores.mp3",
		"LOCKED CLUB, Hofmannita - LAPKI.mp3",
		"Locked Club - It's My Rave.mp3",
		"PORNO - VES.mp3",
		"S.A.Y. - Marrakech.mp3",
		"Steals - im smoke.mp3",
		"Stelmakh - Naidy.mp3",
		"Trust True - EIR.mp3",
		"UNDERGROOVER - Merch.mp3",
		"UNDERGROOVER - Stock.mp3",
		"Unit Boy - ASAP.mp3",
		"WASTA - 100AB.mp3",
	}

	defaultArt := fmt.Sprintf("%s/tracks/images/default.jpg", objectStorage)
	defaultPreview := ""
	year := 2024

	var tracks []interface{}
	for i, fname := range files {
		base := strings.TrimSuffix(fname, filepath.Ext(fname))
		artist := "unknown"
		title := base

		if strings.Contains(base, " - ") {
			parts := strings.SplitN(base, " - ", 2)
			artist = strings.TrimSpace(parts[0])
			title = strings.TrimSpace(parts[1])
		}

		fileURL := fmt.Sprintf("%s/tracks/%s", objectStorage, url.PathEscape(fname))

		t := domain.Track{
			ID:          fmt.Sprintf("track%03d", i+1),
			Title:       title,
			Artist:      artist,
			URL:         fileURL,
			AlbumTitle:  "",
			AlbumArtURL: defaultArt,
			PreviewURL:  defaultPreview,
			Genre:       []string{},
			Year:        year,
		}
		tracks = append(tracks, t)
	}

	// --- Insert into MongoDB ---
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
		log.Printf("- %s by %s (%d) — %s", track.Title, track.Artist, track.Year, track.URL)
	}
}
