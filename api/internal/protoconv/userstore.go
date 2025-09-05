package protoconv

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/radjathaher/alunalun/api/internal/repository"
	"github.com/radjathaher/alunalun/api/internal/utils/auth"
	"github.com/segmentio/ksuid"
)

// PostgresUserStore implements auth.UserStore interface
type PostgresUserStore struct {
	queries *repository.Queries
}

// NewPostgresUserStore creates a new PostgreSQL user store
func NewPostgresUserStore(queries *repository.Queries) *PostgresUserStore {
	return &PostgresUserStore{
		queries: queries,
	}
}

// CreateUser creates a new user in the database
func (s *PostgresUserStore) CreateUser(ctx context.Context, user *auth.User) error {
	// Generate ID if not provided
	idStr := user.ID
	if idStr == "" {
		idStr = ksuid.New().String()
	}

	// Parse UUID
	var userID pgtype.UUID
	err := userID.Scan(idStr)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Create user params
	params := &repository.CreateUserParams{
		ID: userID,
		Email: user.Email, // Email is string in DB, not pointer
		CreatedAt: pgtype.Timestamptz{
			Time:  user.CreatedAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  user.UpdatedAt,
			Valid: true,
		},
	}

	// Set display name from username or first/last name
	if user.Username != "" {
		params.DisplayName = &user.Username
	} else if user.FirstName != "" || user.LastName != "" {
		displayName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
		params.DisplayName = &displayName
	}

	// Create user in database
	createdUser, err := s.queries.CreateUser(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Update user ID from database
	user.ID = createdUser.ID.String()
	return nil
}

// GetUserByEmail retrieves a user by email
func (s *PostgresUserStore) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	repoUser, err := s.queries.GetUserByEmail(ctx, email) // Email is string, not pointer
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return s.repoUserToAuthUser(repoUser)
}

// GetUserByID retrieves a user by ID
func (s *PostgresUserStore) GetUserByID(ctx context.Context, userID string) (*auth.User, error) {
	// Parse UUID
	var id pgtype.UUID
	err := id.Scan(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	repoUser, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return s.repoUserToAuthUser(repoUser)
}

// UpdateUser updates user information
func (s *PostgresUserStore) UpdateUser(ctx context.Context, user *auth.User) error {
	// Parse UUID
	var userID pgtype.UUID
	err := userID.Scan(user.ID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	params := &repository.UpdateUserParams{
		ID: userID,
		Email: user.Email, // Email is string in DB, not pointer
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}

	// Set display name from username or first/last name
	if user.Username != "" {
		params.DisplayName = &user.Username
	} else if user.FirstName != "" || user.LastName != "" {
		displayName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
		params.DisplayName = &displayName
	}

	_, err = s.queries.UpdateUser(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// CheckUsernameAvailable checks if a username is available
// Since we don't have a dedicated username field, check display_name instead
func (s *PostgresUserStore) CheckUsernameAvailable(ctx context.Context, username string) (bool, error) {
	_, err := s.queries.GetUserByDisplayName(ctx, &username)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Username not found, so it's available
			return true, nil
		}
		return false, fmt.Errorf("failed to check username: %w", err)
	}
	// Username found, so it's not available
	return false, nil
}

// repoUserToAuthUser converts repository.User to auth.User
func (s *PostgresUserStore) repoUserToAuthUser(repoUser *repository.User) (*auth.User, error) {
	// Use display name as username
	username := "Anonymous"
	if repoUser.DisplayName != nil {
		username = *repoUser.DisplayName
	}

	return &auth.User{
		ID:        repoUser.ID.String(),
		Email:     repoUser.Email, // Email is string in DB
		Username:  username,
		CreatedAt: repoUser.CreatedAt.Time,
		UpdatedAt: repoUser.UpdatedAt.Time,
		Status:    "active", // Default status
		// Fields not available in current schema
		FirstName:       "",
		LastName:        "",
		Picture:         "",
		EmailVerified:   false,
		EmailVerifiedAt: nil,
		LastLoginAt:     nil,
		Metadata:        nil,
		PasswordHash:    "",
	}, nil
}