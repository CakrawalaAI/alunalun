# **Anonymous-to-Authenticated Social Media Auth System - Implementation Spec**

## **Project Overview**

Build a secure authentication system that allows users to start posting immediately as anonymous users with just a username, then seamlessly upgrade to verified accounts without losing any content. The system prioritizes zero-friction onboarding while maintaining security through stateless JWT tokens.

## **Tech Stack**

### **Backend**
- **Go** with ConnectRPC for type-safe protobuf APIs
- **SQLC** for compile-time safe SQL queries  
- **GOOSE** for database migrations
- **PostgreSQL** database
- **JWT** tokens (non-expiring for anonymous, 1hr for authenticated)
- **Google OAuth 2.0** for account verification

### **Frontend**
- **React** with Feature-Sliced Design architecture
- **ConnectRPC** generated client
- **TanStack Query** for API state management
- **Zustand** for auth state management
- **localStorage** for anonymous token persistence

---

## **Database Schema**

### **Core Design Principles**
- Single `usernames` table enforces uniqueness across all user types
- Anonymous sessions have NO expiry (indefinite)
- Stateless auth - no session storage for authenticated users
- Event tracking for abuse detection

### **GOOSE Migration Files**

```sql
-- migrations/001_users_and_sessions.sql

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Single table for ALL usernames (enforces uniqueness across tiers)
CREATE TABLE usernames (
    username VARCHAR(100) PRIMARY KEY,
    user_id UUID,           -- NULL for anonymous
    session_id UUID,        -- NULL for authenticated
    claimed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Either user_id OR session_id must be set, not both
    CONSTRAINT one_owner CHECK (
        (user_id IS NOT NULL AND session_id IS NULL) OR
        (user_id IS NULL AND session_id IS NOT NULL)
    ),
    
    -- Username format validation
    CONSTRAINT username_format CHECK (
        username ~ '^[a-zA-Z0-9_]{3,100}$'
    )
);

-- Users table for authenticated accounts
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255),
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Email format validation
    CONSTRAINT email_format CHECK (
        email ~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
    )
);

-- OAuth and auth provider tracking
CREATE TABLE user_auth_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,           -- 'google', 'magic_link', etc
    provider_user_id VARCHAR(255) NOT NULL,  -- google_id, etc
    email_verified BOOLEAN DEFAULT FALSE,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(provider, provider_user_id)
);

-- Event tracking for abuse detection
CREATE TABLE user_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,                   -- NULL for anonymous
    session_id UUID,                 -- For anonymous tracking
    event_type VARCHAR(50) NOT NULL, -- 'signup', 'login', 'pin_create', etc
    ip_address INET,
    x_forwarded_for TEXT,            -- Proxy detection
    x_real_ip INET,                  -- CDN/proxy real IP
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB                   -- Event-specific data
);

-- Future: Pins (location-based content)
CREATE TABLE pins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content TEXT NOT NULL,
    author_user_id UUID REFERENCES users(id),
    author_session_id UUID,
    longitude DECIMAL(10, 7) NOT NULL,
    latitude DECIMAL(10, 7) NOT NULL,
    geohash VARCHAR(12) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Exactly one author type must be set
    CHECK ((author_user_id IS NULL) != (author_session_id IS NULL))
);

-- Indexes
CREATE INDEX idx_usernames_user_id ON usernames(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_usernames_session_id ON usernames(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_user_auth_providers_user ON user_auth_providers(user_id);
CREATE INDEX idx_user_auth_providers_provider ON user_auth_providers(provider, provider_user_id);
CREATE INDEX idx_user_events_user ON user_events(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_user_events_session ON user_events(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX idx_user_events_created ON user_events(created_at);
CREATE INDEX idx_pins_author_user ON pins(author_user_id) WHERE author_user_id IS NOT NULL;
CREATE INDEX idx_pins_author_session ON pins(author_session_id) WHERE author_session_id IS NOT NULL;
CREATE INDEX idx_pins_geohash ON pins(geohash);
```

---

## **SQLC Query Definitions**

```sql
-- queries/usernames.sql
-- name: CheckUsername :one
SELECT COUNT(*) > 0 AS taken
FROM usernames
WHERE username = $1;

-- name: ClaimUsernameForAnonymous :one
INSERT INTO usernames (username, session_id)
VALUES ($1, $2)
ON CONFLICT (username) DO NOTHING
RETURNING username, session_id;

-- name: ClaimUsernameForUser :one
INSERT INTO usernames (username, user_id)
VALUES ($1, $2)
ON CONFLICT (username) DO NOTHING
RETURNING username, user_id;

-- name: MigrateUsernameToUser :exec
UPDATE usernames
SET user_id = $2, session_id = NULL
WHERE session_id = $1;

-- queries/users.sql
-- name: CreateUser :one
INSERT INTO users (email, display_name, avatar_url)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT u.*, un.username
FROM users u
JOIN usernames un ON un.user_id = u.id
WHERE u.id = $1;

-- name: GetUserByEmail :one
SELECT u.*, un.username
FROM users u
JOIN usernames un ON un.user_id = u.id
WHERE u.email = $1;

-- queries/auth.sql
-- name: CreateAuthProvider :one
INSERT INTO user_auth_providers (
    user_id, provider, provider_user_id, email_verified, is_primary
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetUserByAuthProvider :one
SELECT u.*, un.username FROM users u
JOIN user_auth_providers uap ON u.id = uap.user_id
JOIN usernames un ON un.user_id = u.id
WHERE uap.provider = $1 AND uap.provider_user_id = $2;

-- queries/events.sql
-- name: CreateUserEvent :one
INSERT INTO user_events (
    user_id, session_id, event_type, ip_address, 
    x_forwarded_for, x_real_ip, user_agent, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- queries/pins.sql (future)
-- name: CreatePin :one
INSERT INTO pins (content, author_user_id, author_session_id, longitude, latitude, geohash)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: MigratePinsToUser :exec
UPDATE pins 
SET author_user_id = $2, author_session_id = NULL
WHERE author_session_id = $1;
```

---

## **ConnectRPC Proto Structure**

### **Directory Structure**
```
proto/
├── v1/
│   ├── service/      # Service definitions
│   │   ├── auth.proto
│   │   └── user.proto
│   ├── entities/     # Data models
│   │   └── user.proto
│   └── common/       # Truly common types
│       └── pagination.proto
```

### **Service Definitions**

```protobuf
// v1/service/auth.proto
syntax = "proto3";

package api.v1.service;

import "v1/entities/user.proto";

service AuthService {
  // Check if username is available
  rpc CheckUsername(CheckUsernameRequest) returns (CheckUsernameResponse);
  
  // Initialize anonymous session (server generates session_id)
  rpc InitAnonymous(InitAnonymousRequest) returns (InitAnonymousResponse);
  
  // Provider-agnostic authentication
  rpc Authenticate(AuthenticateRequest) returns (AuthenticateResponse);
  
  // Refresh expired JWT (using expired token as refresh token)
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
}

message CheckUsernameRequest {
  string username = 1;
}

message CheckUsernameResponse {
  bool available = 1;
  string message = 2;
}

message InitAnonymousRequest {
  string username = 1;
}

message InitAnonymousResponse {
  string token = 1;       // JWT with no expiration
  string session_id = 2;  // Server-generated UUID
  string username = 3;
}

message AuthenticateRequest {
  string provider = 1;              // "google", "magic_link", etc
  string credential = 2;            // ID token, magic token, etc
  optional string session_id = 3;  // For migration from anonymous
}

message AuthenticateResponse {
  string token = 1;                // JWT with 1hr expiration
  api.v1.entities.User user = 2;
  bool session_migrated = 3;
}

message RefreshTokenRequest {
  string expired_token = 1;
}

message RefreshTokenResponse {
  string token = 1;
}
```

### **Endpoint Mapping**
```
POST /api.v1.service.AuthService/CheckUsername
POST /api.v1.service.AuthService/InitAnonymous
POST /api.v1.service.AuthService/Authenticate
POST /api.v1.service.AuthService/RefreshToken
POST /api.v1.service.UserService/GetCurrentUser
POST /api.v1.service.UserService/UpdateProfile
```

---

## **Authentication Flow Documentation**

### **Anonymous User Journey (Server-Generated)**

```mermaid
sequenceDiagram
    participant Client
    participant Backend
    participant Database
    
    Client->>Client: User attempts first action (e.g., create pin)
    Client->>Client: Check localStorage for token
    Client->>Client: No token found - show username modal
    Client->>Backend: CheckUsername(username)
    Backend->>Database: Check usernames table
    Database-->>Backend: Available: true
    Backend-->>Client: Available: true
    Client->>Backend: InitAnonymous(username)
    Backend->>Backend: Generate session_id (UUID)
    Backend->>Database: INSERT INTO usernames
    Backend->>Backend: Generate non-expiring JWT
    Backend-->>Client: {token, session_id, username}
    Client->>Client: localStorage.setItem('anonymous_token', token)
    Client->>Backend: CreatePin(content) + Bearer token
    Backend->>Backend: Parse token, identify as anonymous
    Backend->>Database: INSERT pins with author_session_id
    Backend-->>Client: Pin created
```

### **Authentication Upgrade Journey**

```mermaid
sequenceDiagram
    participant Client
    participant Backend
    participant Google
    participant Database
    
    Client->>Client: User clicks "Sign in with Google"
    Client->>Google: OAuth flow
    Google-->>Client: Credential (ID token)
    Client->>Client: Get session_id from localStorage
    Client->>Backend: Authenticate(provider: "google", credential, session_id?)
    Backend->>Google: Validate ID token
    Google-->>Backend: {email, google_id, verified}
    Backend->>Backend: Generate user_id (UUID)
    Backend->>Database: INSERT users
    Backend->>Database: ClaimUsernameForUser (migrate from session)
    Backend->>Database: INSERT user_auth_providers
    Backend->>Database: UPDATE pins SET author_user_id WHERE author_session_id
    Backend->>Backend: Generate JWT (1hr expiry)
    Backend-->>Client: {token, user, session_migrated: true}
    Client->>Client: Clear localStorage (anonymous token)
    Client->>Client: All future requests use new JWT
```

---

## **Token Strategy**

### **Token Types**

1. **Anonymous Token (Never Expires)**
```json
{
  "type": "anonymous",
  "session_id": "uuid-here",
  "username": "cooluser42",
  "iat": 1704067200
  // No exp claim - token never expires
}
```

2. **Authenticated Token (1hr Expiry, 30-day Refresh)**
```json
{
  "type": "authenticated",
  "sub": "user-uuid",
  "username": "cooluser42",
  "iat": 1704067200,
  "exp": 1704070800,        // 1 hour expiry
  "refresh_until": 1706659200  // 30 days refresh window
}
```

### **Middleware Token Routing**

```go
// middleware/auth.go
func AuthInterceptor(config AuthConfig) connect.UnaryInterceptorFunc {
    return func(next connect.UnaryFunc) connect.UnaryFunc {
        return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
            token := extractBearerToken(req.Header())
            
            if token == "" {
                // Check if endpoint requires auth
                if isProtectedEndpoint(req.Spec().Procedure) {
                    return nil, connect.NewError(connect.CodeUnauthenticated, nil)
                }
                return next(ctx, req)
            }
            
            // Parse token to determine type
            claims := parseTokenClaims(token, config.JWTSecret)
            
            switch claims["type"] {
            case "anonymous":
                // Anonymous tokens never expire
                ctx = context.WithValue(ctx, "auth_type", "anonymous")
                ctx = context.WithValue(ctx, "session_id", claims["session_id"])
                ctx = context.WithValue(ctx, "username", claims["username"])
                
            case "authenticated":
                // Check expiry and refresh window
                if isExpired(claims["exp"]) {
                    if canRefresh(claims["refresh_until"]) {
                        // Allow RefreshToken endpoint only
                        if req.Spec().Procedure == "/api.v1.service.AuthService/RefreshToken" {
                            ctx = context.WithValue(ctx, "expired_token", token)
                            return next(ctx, req)
                        }
                    }
                    return nil, connect.NewError(connect.CodeUnauthenticated, nil)
                }
                ctx = context.WithValue(ctx, "auth_type", "authenticated")
                ctx = context.WithValue(ctx, "user_id", claims["sub"])
            }
            
            return next(ctx, req)
        }
    }
}
```

---

## **Frontend Implementation (Feature-Sliced Design)**

### **Project Structure**

```
web/src/
├── features/
│   ├── auth/
│   │   ├── components/
│   │   │   ├── username-modal.tsx
│   │   │   ├── auth-prompt.tsx
│   │   │   └── google-auth-button.tsx
│   │   ├── hooks/
│   │   │   ├── use-auth.ts
│   │   │   ├── use-check-username.ts
│   │   │   └── use-init-anonymous.ts
│   │   ├── store/
│   │   │   └── auth-store.ts      # Zustand store
│   │   ├── lib/
│   │   │   ├── auth-client.ts
│   │   │   └── token-utils.ts
│   │   ├── types/
│   │   │   └── index.ts
│   │   └── index.ts
│   └── user/
│       ├── components/
│       │   ├── user-profile.tsx
│       │   └── user-avatar.tsx
│       ├── hooks/
│       │   ├── use-current-user.ts
│       │   └── use-update-profile.ts
│       └── index.ts
├── common/
│   ├── services/
│   │   └── connectrpc/
│   │       ├── [generated files]
│   │       └── client.ts
│   └── lib/
│       └── transport.ts
└── blocks/
    └── layout/
        └── auth-guard.tsx
```

### **Zustand Auth Store**

```typescript
// features/auth/store/auth-store.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

type AuthState = {
  type: 'none' | 'anonymous' | 'authenticated';
  token: string | null;
  sessionId: string | null;
  user: User | null;
  
  // Actions
  setAnonymousAuth: (token: string, sessionId: string, username: string) => void;
  setAuthenticatedAuth: (token: string, user: User) => void;
  clearAuth: () => void;
  getAuthHeaders: () => Record<string, string>;
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      type: 'none',
      token: null,
      sessionId: null,
      user: null,
      
      setAnonymousAuth: (token, sessionId, username) => {
        set({
          type: 'anonymous',
          token,
          sessionId,
          user: null,
        });
      },
      
      setAuthenticatedAuth: (token, user) => {
        // Clear anonymous data
        localStorage.removeItem('anonymous_token');
        set({
          type: 'authenticated',
          token,
          sessionId: null,
          user,
        });
      },
      
      clearAuth: () => {
        set({
          type: 'none',
          token: null,
          sessionId: null,
          user: null,
        });
      },
      
      getAuthHeaders: () => {
        const state = get();
        if (state.token) {
          return { Authorization: `Bearer ${state.token}` };
        }
        return {};
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        // Only persist anonymous token
        type: state.type,
        token: state.type === 'anonymous' ? state.token : null,
        sessionId: state.type === 'anonymous' ? state.sessionId : null,
      }),
    }
  )
);
```

### **ConnectRPC Client Setup**

```typescript
// common/services/connectrpc/client.ts
import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { AuthService } from "./api/v1/service/auth_connect";
import { UserService } from "./api/v1/service/user_connect";
import { useAuthStore } from "@/features/auth/store/auth-store";

const createTransport = () => {
  return createConnectTransport({
    baseUrl: import.meta.env.VITE_API_URL,
    interceptors: [
      (next) => async (req) => {
        // Add auth headers
        const headers = useAuthStore.getState().getAuthHeaders();
        Object.entries(headers).forEach(([key, value]) => {
          req.header.set(key, value);
        });
        
        // Auto-refresh logic for authenticated tokens
        const authType = useAuthStore.getState().type;
        if (authType === 'authenticated') {
          const token = useAuthStore.getState().token;
          if (token && isTokenExpiringSoon(token)) {
            await refreshToken();
          }
        }
        
        return next(req);
      },
    ],
  });
};

export const authClient = createPromiseClient(AuthService, createTransport());
export const userClient = createPromiseClient(UserService, createTransport());
```

### **Auth Hooks**

```typescript
// features/auth/hooks/use-init-anonymous.ts
import { useMutation } from "@tanstack/react-query";
import { authClient } from "@/common/services/connectrpc/client";
import { useAuthStore } from "../store/auth-store";

export const useInitAnonymous = () => {
  const setAnonymousAuth = useAuthStore((state) => state.setAnonymousAuth);
  
  return useMutation({
    mutationFn: async (username: string) => {
      const response = await authClient.initAnonymous({ username });
      return response;
    },
    onSuccess: (data) => {
      setAnonymousAuth(data.token, data.sessionId, data.username);
    },
  });
};

// features/auth/hooks/use-authenticate.ts
export const useAuthenticate = () => {
  const { sessionId, setAuthenticatedAuth } = useAuthStore();
  
  return useMutation({
    mutationFn: async ({ provider, credential }: AuthenticateParams) => {
      const response = await authClient.authenticate({
        provider,
        credential,
        sessionId, // Include for migration
      });
      return response;
    },
    onSuccess: (data) => {
      setAuthenticatedAuth(data.token, data.user);
      // Show migration success toast if data.sessionMigrated
    },
  });
};
```

### **Username Modal Component**

```tsx
// features/auth/components/username-modal.tsx
import { useState } from "react";
import { useCheckUsername, useInitAnonymous } from "../hooks";

export const UsernameModal = ({ isOpen, onClose, onSuccess }) => {
  const [username, setUsername] = useState("");
  const checkUsername = useCheckUsername();
  const initAnonymous = useInitAnonymous();
  
  const handleSubmit = async () => {
    // Check availability
    const { available } = await checkUsername.mutateAsync({ username });
    
    if (!available) {
      // Show error
      return;
    }
    
    // Create anonymous session
    await initAnonymous.mutateAsync(username);
    onSuccess();
    onClose();
  };
  
  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Choose your username</DialogTitle>
          <DialogDescription>
            Start posting immediately with just a username. 
            You can verify your account later.
          </DialogDescription>
        </DialogHeader>
        
        <Input
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          placeholder="Enter username"
          pattern="^[a-zA-Z0-9_]{3,100}$"
        />
        
        <DialogFooter>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={!username}>
            Continue
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
```

---

## **Security Considerations**

### **Token Security**
- Anonymous tokens: Stored in localStorage, never expire
- Authenticated tokens: Short-lived (1hr), stored in memory
- Refresh mechanism: Use expired JWT as refresh token (stateless)
- All tokens signed with HS256

### **Username Security**
- Single `usernames` table prevents conflicts
- Server-side validation on all operations
- Race condition handling with DB constraints

### **Rate Limiting** (Redis-based)
```javascript
// In-memory or Redis
const rateLimits = {
  anonymous: {
    pin_create: { limit: 5, window: '1h' },
    comment: { limit: 20, window: '1h' },
  },
  authenticated: {
    pin_create: { limit: 50, window: '1h' },
    comment: { limit: 200, window: '1h' },
  },
};
```

### **Abuse Detection**
- Track all events in `user_events` table
- Monitor by IP address and session_id
- Detect patterns: rapid account creation, spam posting

---

## **Deployment Configuration**

### **Environment Variables**

```env
# Backend (.env)
DATABASE_URL=postgresql://user:pass@localhost/dbname
JWT_SECRET=your-256-bit-secret-key
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
REDIS_URL=redis://localhost:6379

# Frontend (.env)
VITE_API_URL=http://localhost:8080
VITE_GOOGLE_CLIENT_ID=your-google-client-id
```

### **Commands**

```bash
# Backend
cd api
sqlc generate              # Generate repository code
cd proto && buf generate   # Generate proto code
go mod tidy               # Install dependencies
go run ./cmd/server       # Start server

# Frontend
cd web
bun install               # Install dependencies
bun run dev              # Start dev server
```

---

## **Success Metrics**

### **User Experience**
✅ Zero-friction posting (username only)  
✅ Persistent anonymous identity (no expiry)  
✅ Seamless content migration on upgrade  
✅ No loss of data during auth transitions  

### **Security**
✅ Stateless authentication (JWT-based)  
✅ Username uniqueness enforced  
✅ Rate limiting by auth type  
✅ Abuse detection and tracking  

### **Performance**
✅ Username check < 50ms  
✅ Anonymous session creation < 100ms  
✅ Token validation < 10ms (no DB lookup)  
✅ Content migration < 500ms  

This implementation provides a complete, secure, and user-friendly authentication system optimized for social media applications with location-based content (pins).