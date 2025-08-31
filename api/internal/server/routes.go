package server

import (
	"net/http"
)

// SetupRoutes configures all HTTP routes
// Each feature adds their routes in their designated section
func SetupRoutes(mux *http.ServeMux) {
	// === API ROUTES START ===
	// Each feature branch adds their routes in their designated section
	// Keep all sections when merging - these are additive changes
	
	// [POSTS-ROUTES-START]
	// Owner: feature/posts-core
	// Post-related routes
	// mux.HandleFunc("/api/posts", postsHandler)
	// mux.HandleFunc("/api/posts/{id}", postHandler)
	// [POSTS-ROUTES-END]
	
	// [COMMENTS-ROUTES-START]
	// Owner: feature/comments-system
	// Comment-related routes
	// mux.HandleFunc("/api/posts/{id}/comments", commentsHandler)
	// [COMMENTS-ROUTES-END]
	
	// [REACTIONS-ROUTES-START]
	// Owner: feature/reactions-system
	// Reaction-related routes
	// mux.HandleFunc("/api/posts/{id}/reactions", reactionsHandler)
	// [REACTIONS-ROUTES-END]
	
	// [FEED-ROUTES-START]
	// Owner: feature/location-feed
	// Feed-related routes
	// mux.HandleFunc("/api/feed/location", locationFeedHandler)
	// mux.HandleFunc("/api/feed/trending", trendingFeedHandler)
	// [FEED-ROUTES-END]
	
	// [MAP-ROUTES-START]
	// Owner: feature/map-pins
	// Map-related routes
	// mux.HandleFunc("/api/map/pins", mapPinsHandler)
	// mux.HandleFunc("/api/map/clusters", clustersHandler)
	// [MAP-ROUTES-END]
	
	// [MEDIA-ROUTES-START]
	// Owner: feature/media-upload
	// Media-related routes
	// mux.HandleFunc("/api/media/upload", uploadHandler)
	// [MEDIA-ROUTES-END]
	
	// === API ROUTES END ===
}

// SetupMiddleware configures middleware stack
// Each feature can add their middleware in order
func SetupMiddleware(handler http.Handler) http.Handler {
	// === MIDDLEWARE STACK START ===
	// Add middleware in execution order
	// Be careful about ordering when merging
	
	// [CORS-MIDDLEWARE-START]
	// Owner: base/main
	// handler = CORSMiddleware(handler)
	// [CORS-MIDDLEWARE-END]
	
	// [AUTH-MIDDLEWARE-START]
	// Owner: feature/auth
	// handler = AuthMiddleware(handler)
	// [AUTH-MIDDLEWARE-END]
	
	// [LOGGING-MIDDLEWARE-START]
	// Owner: base/main
	// handler = LoggingMiddleware(handler)
	// [LOGGING-MIDDLEWARE-END]
	
	// [RATELIMIT-MIDDLEWARE-START]
	// Owner: feature/ratelimit
	// handler = RateLimitMiddleware(handler)
	// [RATELIMIT-MIDDLEWARE-END]
	
	// === MIDDLEWARE STACK END ===
	
	return handler
}