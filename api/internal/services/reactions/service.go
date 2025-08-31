package reactions

// This file is owned by: feature/reactions-system
// Dependencies: posts service (must be merged first)
// Database tables: post_reactions, reaction_activity
// Proto files: proto/v1/service/reactions.proto

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

// Core reaction operations will go here:
// - AddReaction
// - RemoveReaction
// - ChangeReaction
// - GetPostReactions
// - GetUserReactionsForPost
// - GetUserReactionsForPosts (batch)
// - GetReactionActivity

// === IMPLEMENTATION END ===