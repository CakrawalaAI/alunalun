package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

// OAuthState represents the state data for OAuth flows
type OAuthState struct {
	Nonce       string    `json:"n"`   // Random nonce for CSRF protection
	Provider    string    `json:"p"`   // OAuth provider (google, apple, etc)
	RedirectURI string    `json:"r"`   // Client redirect URI after auth
	SessionID   string    `json:"sid,omitempty"` // Optional session ID for migration
	CreatedAt   int64     `json:"iat"` // Unix timestamp
	ExpiresAt   int64     `json:"exp"` // Unix timestamp
}

// StateManager handles stateless OAuth state management
type StateManager struct {
	encryptionKey []byte // 32 bytes for AES-256
	stateTTL      time.Duration
}

// NewStateManager creates a new state manager with encryption
func NewStateManager(encryptionKey []byte, stateTTL time.Duration) (*StateManager, error) {
	if len(encryptionKey) != 32 {
		return nil, errors.New("encryption key must be 32 bytes for AES-256")
	}
	
	if stateTTL == 0 {
		stateTTL = 10 * time.Minute // Default 10 minutes
	}
	
	return &StateManager{
		encryptionKey: encryptionKey,
		stateTTL:      stateTTL,
	}, nil
}

// GenerateState creates an encrypted state token for OAuth flow
func (sm *StateManager) GenerateState(provider, redirectURI, sessionID string) (string, error) {
	// Generate random nonce
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	
	now := time.Now()
	state := OAuthState{
		Nonce:       base64.URLEncoding.EncodeToString(nonce),
		Provider:    provider,
		RedirectURI: redirectURI,
		SessionID:   sessionID,
		CreatedAt:   now.Unix(),
		ExpiresAt:   now.Add(sm.stateTTL).Unix(),
	}
	
	// Marshal state to JSON
	plaintext, err := json.Marshal(state)
	if err != nil {
		return "", fmt.Errorf("failed to marshal state: %w", err)
	}
	
	// Encrypt the state
	ciphertext, err := sm.encrypt(plaintext)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt state: %w", err)
	}
	
	// Encode to URL-safe base64
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// ValidateState decrypts and validates an OAuth state token
func (sm *StateManager) ValidateState(stateToken string) (*OAuthState, error) {
	// Decode from base64
	ciphertext, err := base64.URLEncoding.DecodeString(stateToken)
	if err != nil {
		return nil, fmt.Errorf("invalid state token format: %w", err)
	}
	
	// Decrypt the state
	plaintext, err := sm.decrypt(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt state: %w", err)
	}
	
	// Unmarshal state
	var state OAuthState
	if err := json.Unmarshal(plaintext, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}
	
	// Check expiration
	if time.Now().Unix() > state.ExpiresAt {
		return nil, errors.New("state token expired")
	}
	
	return &state, nil
}

// encrypt performs AES-256-GCM encryption
func (sm *StateManager) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(sm.encryptionKey)
	if err != nil {
		return nil, err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	
	// Encrypt and prepend nonce
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt performs AES-256-GCM decryption
func (sm *StateManager) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(sm.encryptionKey)
	if err != nil {
		return nil, err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	
	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	
	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	
	return plaintext, nil
}

// GenerateEncryptionKey generates a secure random encryption key
func GenerateEncryptionKey() ([]byte, error) {
	key := make([]byte, 32) // 32 bytes for AES-256
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}
	return key, nil
}

// ParseProvider extracts the provider from a state token without full validation
// Useful for routing before full validation
func (sm *StateManager) ParseProvider(stateToken string) (string, error) {
	state, err := sm.ValidateState(stateToken)
	if err != nil {
		return "", err
	}
	return state.Provider, nil
}