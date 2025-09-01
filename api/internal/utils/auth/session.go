package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

// SessionStore defines the interface for session storage
type SessionStore interface {
	// Create creates a new session
	Create(ctx context.Context, session *Session) error
	
	// Get retrieves a session by ID
	Get(ctx context.Context, sessionID string) (*Session, error)
	
	// Update updates an existing session
	Update(ctx context.Context, session *Session) error
	
	// Delete removes a session
	Delete(ctx context.Context, sessionID string) error
	
	// FindByUserID finds all sessions for a user
	FindByUserID(ctx context.Context, userID string) ([]*Session, error)
	
	// DeleteExpired removes all expired sessions
	DeleteExpired(ctx context.Context) error
}

// SessionManager handles session lifecycle
type SessionManager struct {
	store SessionStore
	idGen func() string
}

// NewSessionManager creates a new session manager
func NewSessionManager(store SessionStore) *SessionManager {
	return &SessionManager{
		store: store,
		idGen: generateSessionID,
	}
}

// CreateAnonymous creates a new anonymous session
func (sm *SessionManager) CreateAnonymous(ctx context.Context, username string) (*Session, error) {
	if username == "" {
		return nil, errors.New("username is required for anonymous session")
	}
	
	session := &Session{
		ID:          sm.idGen(),
		Username:    username,
		IsAnonymous: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		// Anonymous sessions don't expire
		ExpiresAt:   nil,
	}
	
	if err := sm.store.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create anonymous session: %w", err)
	}
	
	return session, nil
}

// CreateAuthenticated creates a new authenticated session
func (sm *SessionManager) CreateAuthenticated(ctx context.Context, userID string, ttl time.Duration) (*Session, error) {
	if userID == "" {
		return nil, errors.New("userID is required for authenticated session")
	}
	
	now := time.Now()
	expiresAt := now.Add(ttl)
	
	session := &Session{
		ID:          sm.idGen(),
		UserID:      userID,
		IsAnonymous: false,
		CreatedAt:   now,
		UpdatedAt:   now,
		ExpiresAt:   &expiresAt,
	}
	
	if err := sm.store.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create authenticated session: %w", err)
	}
	
	return session, nil
}

// MigrateToUser converts an anonymous session to an authenticated one
func (sm *SessionManager) MigrateToUser(ctx context.Context, sessionID, userID string) error {
	if sessionID == "" || userID == "" {
		return errors.New("sessionID and userID are required")
	}
	
	// Get existing session
	session, err := sm.store.Get(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	
	if !session.IsAnonymous {
		return errors.New("session is already authenticated")
	}
	
	// Update session
	session.UserID = userID
	session.IsAnonymous = false
	session.UpdatedAt = time.Now()
	
	// Set expiry for migrated session (1 hour by default)
	expiresAt := time.Now().Add(time.Hour)
	session.ExpiresAt = &expiresAt
	
	if err := sm.store.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to migrate session: %w", err)
	}
	
	return nil
}

// Validate checks if a session is valid
func (sm *SessionManager) Validate(ctx context.Context, sessionID string) (*Session, error) {
	session, err := sm.store.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	
	// Check expiry
	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now()) {
		return nil, &AuthError{
			Code:    ErrTokenExpired,
			Message: "session has expired",
		}
	}
	
	return session, nil
}

// Refresh extends the expiry of an authenticated session
func (sm *SessionManager) Refresh(ctx context.Context, sessionID string, ttl time.Duration) (*Session, error) {
	session, err := sm.Validate(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	
	if session.IsAnonymous {
		return nil, errors.New("anonymous sessions cannot be refreshed")
	}
	
	// Update expiry
	expiresAt := time.Now().Add(ttl)
	session.ExpiresAt = &expiresAt
	session.UpdatedAt = time.Now()
	
	if err := sm.store.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to refresh session: %w", err)
	}
	
	return session, nil
}

// Revoke deletes a session
func (sm *SessionManager) Revoke(ctx context.Context, sessionID string) error {
	return sm.store.Delete(ctx, sessionID)
}

// RevokeAllForUser deletes all sessions for a user
func (sm *SessionManager) RevokeAllForUser(ctx context.Context, userID string) error {
	sessions, err := sm.store.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user sessions: %w", err)
	}
	
	for _, session := range sessions {
		if err := sm.store.Delete(ctx, session.ID); err != nil {
			// Log error but continue
			fmt.Printf("failed to delete session %s: %v\n", session.ID, err)
		}
	}
	
	return nil
}

// CleanupExpired removes all expired sessions
func (sm *SessionManager) CleanupExpired(ctx context.Context) error {
	return sm.store.DeleteExpired(ctx)
}

// generateSessionID generates a cryptographically secure session ID
func generateSessionID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("session_%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(b)
}

// InMemorySessionStore provides an in-memory implementation of SessionStore for testing
type InMemorySessionStore struct {
	sessions map[string]*Session
}

// NewInMemorySessionStore creates a new in-memory session store
func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		sessions: make(map[string]*Session),
	}
}

func (s *InMemorySessionStore) Create(ctx context.Context, session *Session) error {
	if _, exists := s.sessions[session.ID]; exists {
		return errors.New("session already exists")
	}
	s.sessions[session.ID] = session
	return nil
}

func (s *InMemorySessionStore) Get(ctx context.Context, sessionID string) (*Session, error) {
	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, &AuthError{
			Code:    ErrSessionNotFound,
			Message: "session not found",
		}
	}
	return session, nil
}

func (s *InMemorySessionStore) Update(ctx context.Context, session *Session) error {
	if _, exists := s.sessions[session.ID]; !exists {
		return errors.New("session not found")
	}
	s.sessions[session.ID] = session
	return nil
}

func (s *InMemorySessionStore) Delete(ctx context.Context, sessionID string) error {
	delete(s.sessions, sessionID)
	return nil
}

func (s *InMemorySessionStore) FindByUserID(ctx context.Context, userID string) ([]*Session, error) {
	var sessions []*Session
	for _, session := range s.sessions {
		if session.UserID == userID {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (s *InMemorySessionStore) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	for id, session := range s.sessions {
		if session.ExpiresAt != nil && session.ExpiresAt.Before(now) {
			delete(s.sessions, id)
		}
	}
	return nil
}