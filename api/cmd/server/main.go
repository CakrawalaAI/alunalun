package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	// Import services as they are implemented
)

func main() {
	// === INITIALIZATION START ===
	// Each feature can add their initialization in their section
	
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://localhost/alunalun?sslmode=disable"
	}
	
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	
	// [POSTS-INIT-START]
	// Owner: feature/posts-core
	// Initialize posts service
	// postsService := posts.NewService(db)
	// [POSTS-INIT-END]
	
	// [COMMENTS-INIT-START]
	// Owner: feature/comments-system
	// Initialize comments service
	// commentsService := comments.NewService(db)
	// [COMMENTS-INIT-END]
	
	// [REACTIONS-INIT-START]
	// Owner: feature/reactions-system
	// Initialize reactions service
	// reactionsService := reactions.NewService(db)
	// [REACTIONS-INIT-END]
	
	// [FEED-INIT-START]
	// Owner: feature/location-feed
	// Initialize feed service
	// feedService := feed.NewService(db)
	// [FEED-INIT-END]
	
	// [MAP-INIT-START]
	// Owner: feature/map-pins
	// Initialize map service
	// mapService := map.NewService(db)
	// [MAP-INIT-END]
	
	// === INITIALIZATION END ===
	
	// Create server mux
	mux := http.NewServeMux()
	
	// === SERVICE REGISTRATION START ===
	// Services are registered here
	// See api/internal/server/server.go for the actual registration
	// === SERVICE REGISTRATION END ===
	
	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("Server failed:", err)
	}
}