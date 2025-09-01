# Claude Guide: Authentication Service

Guidelines for Claude AI assistants working on the authentication service.

## Understanding the Architecture

### Core Principle: Full Server-Side OAuth
**Important**: This is NOT a hybrid/SPA OAuth flow. The server handles EVERYTHING:
- Server initiates OAuth
- Server handles redirects
- Server exchanges codes
- Client only knows about redirects, never sees OAuth details

### Mental Model
```
Think of it like a restaurant:
- Client = Customer (just orders)
- Server = Waiter (handles everything)
- OAuth Provider = Kitchen (does the work)

Customer never goes to kitchen, waiter handles all interactions.
```

## Common Tasks

### 1. Adding a New OAuth Provider

**DO:**
```go
// 1. Create provider in internal/utils/oauth/
type AppleProvider struct {
    BaseProvider
    // Apple-specific fields
}

// 2. Implement required methods
func (p *AppleProvider) Authenticate(ctx, credential) (*auth.UserInfo, error)
func (p *AppleProvider) Name() string { return "apple" }

// 3. Register in oauth_handler.go routes
mux.HandleFunc("/auth/oauth/apple", h.handleOAuthInitiate)
mux.HandleFunc("/auth/oauth/apple/callback", h.handleOAuthCallback)

// 4. Add to example main.go registration
```

**DON'T:**
- Don't make client handle OAuth codes
- Don't store state in memory/Redis (use encrypted JWT)
- Don't create separate refresh tokens

### 2. Modifying Authentication Flow

**ALWAYS CHECK:**
1. Does this maintain stateless design?
2. Does this work for both web AND mobile?
3. Is the client still OAuth-agnostic?

**Example Decision Tree:**
```
Need to add MFA?
├─ Is it after initial auth? → Add to session validation
├─ Is it during OAuth? → Add after user creation in callback
└─ Is it for specific providers? → Add provider-specific logic
```

### 3. Debugging Authentication Issues

**Step 1: Identify Layer**
```go
// Check where it fails:
1. OAuth initiation? → Check state generation
2. OAuth callback? → Check state validation, code exchange
3. Token generation? → Check JWT signing, claims
4. Token validation? → Check expiry, signature
```

**Step 2: Common Issues**
```go
// State validation fails
→ Check encryption key is consistent
→ Check state hasn't expired (10 min TTL)

// Code exchange fails  
→ Check redirect_uri matches exactly
→ Check client_id/secret are correct
→ Check provider endpoints are correct

// Token refresh fails
→ Check if within 30-day window
→ Check if token is actually expired
→ Check if anonymous token (can't refresh)
```

## Security Considerations

### NEVER DO:
```go
// ❌ Store plain text state
state := "random_string" // WRONG

// ❌ Skip state validation
// Always validate state in callback

// ❌ Use long-lived tokens for authenticated users
expiry := 365 * 24 * time.Hour // WRONG

// ❌ Store secrets in code
clientSecret := "abc123" // WRONG - use env vars
```

### ALWAYS DO:
```go
// ✅ Encrypt state with AES-256
state := stateManager.GenerateState(...)

// ✅ Validate state on every callback
state, err := stateManager.ValidateState(stateParam)

// ✅ Use short-lived tokens (1 hour)
expiry := time.Hour

// ✅ Use environment variables
clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
```

## Testing Approach

### 1. Unit Testing Priorities
```go
// High Priority (stateless, critical path):
- Token generation/validation
- State encryption/decryption
- Provider authentication logic

// Medium Priority (integration points):
- OAuth callback handling
- Session migration
- User creation/update

// Low Priority (external dependencies):
- OAuth provider endpoints
- Database operations
```

### 2. Integration Testing
```bash
# Test full OAuth flow
1. Start server with test credentials
2. Use test OAuth provider (Google has test users)
3. Verify token generation
4. Verify session migration
5. Verify token refresh
```

### 3. Test Data Setup
```go
// Always use in-memory stores for testing
sessionStore := auth.NewInMemorySessionStore()
userStore := NewInMemoryUserStore()

// Generate test keys
privateKey, publicKey, _ := auth.GenerateKeyPair()

// Use test encryption key
stateKey, _ := auth.GenerateEncryptionKey()
```

## Adding Features

### Before Adding Any Feature, Ask:

1. **Does it break statelessness?**
   - If it needs storage → Can it use encrypted JWT?
   - If it needs session → Can it use existing session?

2. **Does it work for all clients?**
   - Web browsers?
   - Mobile apps?
   - API clients?

3. **Does it maintain backward compatibility?**
   - Existing tokens still work?
   - Existing endpoints unchanged?

### Feature Implementation Checklist

- [ ] Read the current flow in README.md
- [ ] Identify integration points
- [ ] Maintain stateless design
- [ ] Support both web and mobile
- [ ] Add to provider registry if new provider
- [ ] Update environment variables
- [ ] Add tests
- [ ] Update documentation

## Common Patterns

### Pattern 1: Provider-Specific Logic
```go
// Use type assertion for provider-specific features
if googleProvider, ok := provider.(*oauth.GoogleProvider); ok {
    // Google-specific logic
    googleProvider.VerifyIDToken(idToken)
}
```

### Pattern 2: Client Detection
```go
// Always support both response types
if h.isMobileClient(r) {
    h.respondJSON(w, data)  // Mobile gets JSON
} else {
    http.Redirect(w, r, url) // Web gets redirect
}
```

### Pattern 3: Token Claims Extension
```go
// Add new claims without breaking existing
claims := &auth.Claims{
    // Standard claims
    UserID: user.ID,
    // New feature claims
    Metadata: map[string]interface{}{
        "feature_flag": true,
    },
}
```

## Performance Optimization

### Areas to Focus:
1. **State Encryption** - Cache cipher if bottleneck
2. **Token Validation** - Cache public key parsing
3. **User Lookup** - Add caching layer
4. **OAuth Exchange** - Can't optimize (external)

### DON'T Optimize:
- JWT signing (already fast)
- State generation (needs randomness)
- Session creation (needs uniqueness)

## Troubleshooting Guide

### Problem: "State validation failed"
```go
// Check 1: Encryption key
log.Printf("State key: %x", stateKey)

// Check 2: State expiry
log.Printf("State age: %v", time.Since(state.CreatedAt))

// Check 3: State tampering
log.Printf("Raw state: %s", stateParam)
```

### Problem: "Token refresh not working"
```go
// Check 1: Token is expired
if claims.ExpiresAt.After(time.Now()) {
    log.Println("Token not expired yet")
}

// Check 2: Within refresh window
refreshDeadline := claims.ExpiresAt.Add(30 * 24 * time.Hour)
if time.Now().After(refreshDeadline) {
    log.Println("Refresh window expired")
}

// Check 3: Anonymous token
if claims.IsAnonymous {
    log.Println("Anonymous tokens can't refresh")
}
```

### Problem: "OAuth callback fails"
```go
// Check 1: Provider configuration
log.Printf("Redirect URI: %s", provider.GetOAuth2Config().RedirectURL)

// Check 2: State parameter
log.Printf("State present: %v", stateParam != "")

// Check 3: Error from provider
log.Printf("OAuth error: %s", r.URL.Query().Get("error"))
```

## Important Files to Understand

### Priority 1 (Core Logic):
- `oauth_handler.go` - HTTP OAuth endpoints
- `internal/utils/auth/state.go` - State encryption
- `internal/utils/auth/token.go` - JWT management

### Priority 2 (Providers):
- `internal/utils/oauth/google.go` - OAuth example
- `internal/utils/auth/email_password.go` - Internal auth example

### Priority 3 (Integration):
- `service.go` - gRPC endpoints
- `router.go` - HTTP/gRPC combination
- `cmd/auth-example/main.go` - Full setup example

## Design Philosophy

### Stateless > Stateful
Always prefer stateless solutions. If you think you need Redis/database:
1. Can it be encoded in JWT?
2. Can it be encrypted in state?
3. Can it be derived from existing data?

### Simple > Complex
- One token type (no separate refresh tokens)
- One state mechanism (encrypted JWT)
- One response pattern (redirect or JSON)

### Secure by Default
- Short token expiry (1 hour)
- State expiry (10 minutes)
- Everything encrypted
- CSRF protection built-in

## When to Ask for Help

Ask for clarification when:
1. Breaking stateless design seems necessary
2. OAuth provider has non-standard flow
3. Security implications are unclear
4. Backward compatibility might break

## Remember

This auth system is designed to be:
- **Stateless**: No shared state between servers
- **Scalable**: Any server handles any request
- **Simple**: Client knows nothing about OAuth
- **Secure**: Multiple layers of protection

Keep these principles in mind when making any changes.