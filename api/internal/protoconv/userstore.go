package protoconv

import (
	"context"
	"encoding/json"
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
	if user.ID == "" {
		user.ID = "user-" + ksuid.New().String()
	}

	// Prepare metadata
	metadata := make(map[string]interface{})
	if user.Metadata != nil {
		metadata = user.Metadata
	}
	
	// Add auth-related metadata
	metadata["first_name"] = user.FirstName
	metadata["last_name"] = user.LastName
	metadata["picture"] = user.Picture
	metadata["email_verified"] = user.EmailVerified
	metadata["status"] = user.Status
	
	if user.EmailVerifiedAt != nil {
		metadata["email_verified_at"] = user.EmailVerifiedAt.Format(time.RFC3339)
	}
	if user.LastLoginAt != nil {
		metadata["last_login_at"] = user.LastLoginAt.Format(time.RFC3339)
	}

	// Serialize metadata to JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create user params
	params := &repository.CreateUserParams{
		ID:       user.ID,
		Username: user.Username,
		Metadata: metadataJSON,
		CreatedAt: pgtype.Timestamptz{
			Time:  user.CreatedAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  user.UpdatedAt,
			Valid: true,
		},
	}

	// Set email if not empty
	if user.Email != "" {
		params.Email = &user.Email
	}

	// Create user in database
	createdUser, err := s.queries.CreateUser(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Update user ID from database
	user.ID = createdUser.ID
	return nil
}

// GetUserByEmail retrieves a user by email
func (s *PostgresUserStore) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	repoUser, err := s.queries.GetUserByEmail(ctx, &email)
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
	repoUser, err := s.queries.GetUserByID(ctx, userID)
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
	// Prepare metadata
	metadata := make(map[string]interface{})
	if user.Metadata != nil {
		metadata = user.Metadata
	}
	
	// Update auth-related metadata
	metadata["first_name"] = user.FirstName
	metadata["last_name"] = user.LastName
	metadata["picture"] = user.Picture
	metadata["email_verified"] = user.EmailVerified
	metadata["status"] = user.Status
	
	if user.EmailVerifiedAt != nil {
		metadata["email_verified_at"] = user.EmailVerifiedAt.Format(time.RFC3339)
	}
	if user.LastLoginAt != nil {
		metadata["last_login_at"] = user.LastLoginAt.Format(time.RFC3339)
	}

	// Serialize metadata
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	params := &repository.UpdateUserParams{
		ID:       user.ID,
		Username: user.Username,
		Metadata: metadataJSON,
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}

	// Set email if not empty
	if user.Email != "" {
		params.Email = &user.Email
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
func (s *PostgresUserStore) CheckUsernameAvailable(ctx context.Context, username string) (bool, error) {
	_, err := s.queries.GetUserByUsername(ctx, username)
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
	// Parse metadata
	var metadata map[string]interface{}
	if len(repoUser.Metadata) > 0 {
		if err := json.Unmarshal(repoUser.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	// Extract fields from metadata
	firstName, _ := metadata["first_name"].(string)
	lastName, _ := metadata["last_name"].(string)
	picture, _ := metadata["picture"].(string)
	emailVerified, _ := metadata["email_verified"].(bool)
	status, _ := metadata["status"].(string)
	if status == "" {
		status = "active"
	}

	// Parse timestamps from metadata
	var emailVerifiedAt *time.Time
	if evStr, ok := metadata["email_verified_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, evStr); err == nil {
			emailVerifiedAt = &t
		}
	}

	var lastLoginAt *time.Time
	if llStr, ok := metadata["last_login_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, llStr); err == nil {
			lastLoginAt = &t
		}
	}

	// Get email (nullable)
	email := ""
	if repoUser.Email != nil {
		email = *repoUser.Email
	}

	return &auth.User{
		ID:              repoUser.ID,
		Email:           email,
		Username:        repoUser.Username,
		FirstName:       firstName,
		LastName:        lastName,
		Picture:         picture,
		EmailVerified:   emailVerified,
		EmailVerifiedAt: emailVerifiedAt,
		CreatedAt:       repoUser.CreatedAt.Time,
		UpdatedAt:       repoUser.UpdatedAt.Time,
		LastLoginAt:     lastLoginAt,
		Metadata:        metadata,
		Status:          status,
		// PasswordHash not used in OAuth-based auth
		PasswordHash: "",
	}, nil
}