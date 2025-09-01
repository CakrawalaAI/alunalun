package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenManager handles JWT token generation and validation
type TokenManager struct {
	privateKey    *rsa.PrivateKey
	publicKey     *rsa.PublicKey
	issuer        string
	audience      string
	refreshWindow time.Duration // How long after expiry can tokens be refreshed
}

// Claims represents the JWT claims
type Claims struct {
	jwt.RegisteredClaims
	UserID      string                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id"`
	Username    string                 `json:"username,omitempty"`
	Email       string                 `json:"email,omitempty"`
	Provider    string                 `json:"provider"`
	IsAnonymous bool                   `json:"is_anonymous"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewTokenManager creates a new token manager
func NewTokenManager(privateKeyPEM, publicKeyPEM []byte, issuer, audience string) (*TokenManager, error) {
	// Parse private key
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.New("failed to parse private key PEM")
	}
	
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not RSA")
		}
	}
	
	// Parse public key
	block, _ = pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, errors.New("failed to parse public key PEM")
	}
	
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	
	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not RSA")
	}
	
	return &TokenManager{
		privateKey:    privateKey,
		publicKey:     publicKey,
		issuer:        issuer,
		audience:      audience,
		refreshWindow: 30 * 24 * time.Hour, // 30 days default refresh window
	}, nil
}

// GenerateKeyPair generates a new RSA key pair for testing
func GenerateKeyPair() (privateKeyPEM, publicKeyPEM []byte, err error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	
	// Encode private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	
	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	publicKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	
	return privateKeyPEM, publicKeyPEM, nil
}

// GenerateToken generates a new JWT token
func (tm *TokenManager) GenerateToken(claims *Claims, expiry time.Duration) (string, error) {
	now := time.Now()
	
	// Set standard claims
	claims.Issuer = tm.issuer
	claims.Audience = []string{tm.audience}
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.NotBefore = jwt.NewNumericDate(now)
	
	// Set expiry only if duration is positive (anonymous tokens don't expire)
	if expiry > 0 {
		claims.ExpiresAt = jwt.NewNumericDate(now.Add(expiry))
	}
	
	// Generate JWT ID
	jti := make([]byte, 16)
	if _, err := rand.Read(jti); err != nil {
		return "", err
	}
	claims.ID = base64.URLEncoding.EncodeToString(jti)
	
	// Create and sign token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(tm.privateKey)
}

// ValidateToken validates and parses a JWT token
func (tm *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.publicKey, nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	
	// Additional validation
	if claims.Issuer != tm.issuer {
		return nil, errors.New("invalid token issuer")
	}
	
	if len(claims.Audience) > 0 && claims.Audience[0] != tm.audience {
		return nil, errors.New("invalid token audience")
	}
	
	return claims, nil
}

// RefreshToken generates a new token from an expired one
func (tm *TokenManager) RefreshToken(expiredToken string, newExpiry time.Duration) (string, error) {
	// Parse without validating expiry
	parser := jwt.NewParser(
		jwt.WithoutClaimsValidation(),
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}),
	)
	
	token, err := parser.ParseWithClaims(expiredToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return tm.publicKey, nil
	})
	
	if err != nil {
		return "", fmt.Errorf("failed to parse expired token: %w", err)
	}
	
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", errors.New("invalid token claims")
	}
	
	// Don't refresh anonymous tokens (they don't expire)
	if claims.IsAnonymous {
		return "", errors.New("anonymous tokens cannot be refreshed")
	}
	
	// Check if token is actually expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.After(time.Now()) {
		return "", errors.New("token is not expired yet")
	}
	
	// Check if within refresh window
	if claims.ExpiresAt != nil {
		expiredAt := claims.ExpiresAt.Time
		refreshDeadline := expiredAt.Add(tm.refreshWindow)
		if time.Now().After(refreshDeadline) {
			return "", errors.New("token refresh window has expired")
		}
	}
	
	// Generate new token with same claims but new expiry
	newClaims := &Claims{
		UserID:      claims.UserID,
		SessionID:   claims.SessionID,
		Username:    claims.Username,
		Email:       claims.Email,
		Provider:    claims.Provider,
		IsAnonymous: claims.IsAnonymous,
		Metadata:    claims.Metadata,
	}
	
	return tm.GenerateToken(newClaims, newExpiry)
}

// GetPublicKeyPEM returns the public key in PEM format
func (tm *TokenManager) GetPublicKeyPEM() ([]byte, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(tm.publicKey)
	if err != nil {
		return nil, err
	}
	
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}), nil
}