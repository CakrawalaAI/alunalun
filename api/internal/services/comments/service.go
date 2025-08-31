package comments

// This file is owned by: feature/comments-system
// Dependencies: posts service (must be merged first)
// Database tables: post_comments
// Proto files: proto/v1/service/comments.proto

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

// Core comment operations will go here:
// - AddComment
// - UpdateComment
// - DeleteComment (soft delete)
// - GetComment
// - GetPostComments (with threading)
// - GetCommentReplies
// - GetUserComments

// === IMPLEMENTATION END ===