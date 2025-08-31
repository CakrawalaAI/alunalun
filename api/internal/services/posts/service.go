package posts

// This file is owned by: feature/posts-core
// Dependencies: none (base feature)
// Database tables: posts, post_locations, post_media
// Proto files: proto/v1/service/posts.proto

import (
	"context"
	"database/sql"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// === IMPLEMENTATION START ===
// Add service methods below this line
// Do not modify the struct or constructor above

// Core CRUD operations will go here:
// - CreatePost
// - GetPost
// - UpdatePost
// - DeletePost
// - GetMyPosts
// - GetUserPosts
// - GetPostsByLocation
// - GetPostsByRadius

// === IMPLEMENTATION END ===