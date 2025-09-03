# Authentication-First Architecture Specification
**Date**: 2025-09-03  
**Status**: Implemented MVP Foundation

## Overview

This document captures the authentication-first architecture for the Alunalun collaborative map platform. The system prioritizes authenticated users while maintaining public read access and preparing for future anonymous user support.

## Core Design Principles

1. **Auth-First**: All write operations require authentication via OAuth
2. **Public Read**: Map viewing and pin browsing available without authentication
3. **Server-Side OAuth**: Full server-side OAuth flow (not PKCE/SPA)
4. **Future-Ready**: Data model supports future anonymous users without migration complexity
5. **Spatial Optimization**: Geohash-based queries for efficient map data loading

## System Architecture

### Authentication Flow

```
Client → Server OAuth → Google → Server Callback → JWT Generation → Client
```

- **No PKCE**: Server handles entire OAuth flow
- **JWT Tokens**: 1-hour expiry for authenticated users
- **Stateless**: Encrypted state in JWT, no server-side session storage required

### Data Access Patterns

| Operation | Authentication Required | Notes |
|-----------|-------------------------|--------|
| List Pins | No | Public read with zoom-based filtering |
| Get Pin | No | Public read for individual pins |
| Create Pin | Yes | Requires valid JWT |
| Delete Pin | Yes | Requires JWT + ownership verification |
| Update Pin | Yes | Requires JWT + ownership verification |

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE,        -- Nullable for future anonymous support
    username TEXT NOT NULL UNIQUE,
    metadata JSONB,           -- Flexible metadata storage
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
```

### Posts Table (Polymorphic)
```sql
CREATE TABLE posts (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    type TEXT NOT NULL,       -- 'pin' for now, 'comment' later
    content TEXT,
    parent_id TEXT REFERENCES posts(id), -- For future comments
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
```

### Posts Location Table (PostGIS)
```sql
CREATE TABLE posts_location (
    post_id TEXT PRIMARY KEY REFERENCES posts(id) ON DELETE CASCADE,
    location GEOMETRY(Point, 4326) NOT NULL,
    geohash TEXT,             -- For efficient spatial queries
    created_at TIMESTAMPTZ NOT NULL
);

-- Spatial indexes
CREATE INDEX idx_posts_location_gist ON posts_location USING GIST (location);
CREATE INDEX idx_posts_location_geohash ON posts_location(geohash text_pattern_ops);
```

### User Events Table
```sql
CREATE TABLE user_events (
    id TEXT PRIMARY KEY,
    user_id TEXT REFERENCES users(id),
    session_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    ip_address INET NOT NULL,
    fingerprint TEXT,         -- Device fingerprint for abuse detection
    created_at TIMESTAMPTZ NOT NULL
);
```

## Service Layer Architecture

### AuthService
- Handles OAuth flow (Google initially)
- JWT token generation and validation
- Session management for future anonymous support
- Already implemented in `internal/services/auth/`

### UserService
- `GetCurrentUser()`: Returns current authenticated user or 401
- `RegisterUser()`: Creates user after OAuth success
- Future: `CreateAnonymousUser()` when OAuth limits hit

### PinService
- `CreatePin()`: Requires auth, creates post + location
- `DeletePin()`: Requires auth + ownership
- `GetPin()`: Public read, no auth required
- `ListPins()`: Public read with smart zoom-based loading

## Zoom-Based Pin Loading Strategy

### Zoom Level Mapping
| Zoom Level | Load Pins? | Geohash Precision | Max Pins | Area Coverage |
|------------|------------|-------------------|----------|---------------|
| 1-11 | No | - | 0 | Too zoomed out |
| 12-14 | Yes | 4 chars | 50 | ~39km (City) |
| 15-16 | Yes | 5 chars | 100 | ~4.9km (District) |
| 17-18 | Yes | 6 chars | 200 | ~1.2km (Street) |
| 19+ | Yes | 7-8 chars | 500 | ~38-152m (Building) |

### Loading Algorithm
1. **Check minimum zoom**: Return empty if zoom < 12
2. **Calculate geohash**: Precision based on zoom level
3. **Set limit**: Dynamic limit based on zoom (50→100→200→500)
4. **Query strategy**:
   - Zoom 12-14: Newest globally within rough area
   - Zoom 15-16: Newest within precise geohash
   - Zoom 17+: Distance + recency weighted

## API Endpoints

### Authentication Endpoints
- `GET /auth/oauth/google` - Initiate OAuth flow
- `GET /auth/oauth/google/callback` - OAuth callback
- `POST /auth/refresh` - Refresh JWT token

### Pin Endpoints
- `GET /pins?lat=X&lng=Y&zoom=Z` - List pins (public)
- `GET /pins/:id` - Get single pin (public)
- `POST /pins` - Create pin (auth required)
- `DELETE /pins/:id` - Delete pin (auth + owner)

### User Endpoints
- `GET /users/current` - Get current user (auth required)

## Proto Service Definitions

### Pin Service
```protobuf
service PinService {
  // Public read operations
  rpc ListPins(ListPinsRequest) returns (ListPinsResponse);
  rpc GetPin(GetPinRequest) returns (GetPinResponse);
  
  // Authenticated write operations
  rpc CreatePin(CreatePinRequest) returns (CreatePinResponse);
  rpc DeletePin(DeletePinRequest) returns (DeletePinResponse);
}
```

### User Service
```protobuf
service UserService {
  rpc GetCurrentUser(GetCurrentUserRequest) returns (GetCurrentUserResponse);
  rpc RegisterUser(RegisterUserRequest) returns (RegisterUserResponse);
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
```

## Error Handling

### Authentication Errors
```json
{
  "error": "auth_required",
  "message": "Sign in to create pins",
  "auth_url": "/auth/oauth/google"
}
```

### Authorization Errors
```json
{
  "error": "permission_denied",
  "message": "You can only delete your own pins"
}
```

## Future Extensions

### Anonymous User Support
When OAuth provider limits are reached (e.g., 2000 daily tokens):

1. **Anonymous Session Creation**
   ```
   POST /auth/anonymous
   → Creates user with email=NULL, id="anon-{uuid}"
   → Returns never-expiring JWT
   ```

2. **Migration to Authenticated**
   ```
   UPDATE users SET email=?, username=? WHERE id='anon-xxx'
   ```
   - No need to update posts table (user_id remains same)
   - Simple, clean migration path

### Comment System
- Leverage existing polymorphic posts table
- Posts with `parent_id` set become comments
- Same permission model applies

## Implementation Status

### Completed
- ✅ Database schema with migrations
- ✅ SQLC repository layer generated
- ✅ Proto definitions for entities and services
- ✅ Auth service with OAuth support
- ✅ JWT token management
- ✅ ProtoConv converters for User and Pin entities

### To Implement
- [ ] PinService with full CRUD operations
- [ ] UserService for authenticated operations  
- [ ] Middleware for auth extraction
- [ ] Main server wiring

### Architecture Notes
- **ProtoConv Pattern**: Using `internal/protoconv/` for converting between SQLC models and protobuf entities
- **No Intermediate Structs**: Direct conversion between `repository.User` ↔ `entitiesv1.User`
- **Auth Integration**: Auth service works directly with protobuf entities, third-party OAuth providers use minimal conversion

## Security Considerations

1. **JWT Security**
   - Short-lived tokens (1 hour)
   - RS256 signing algorithm
   - Public key endpoint for verification

2. **OAuth State Protection**
   - AES-256 encrypted state
   - 10-minute expiry
   - CSRF protection built-in

3. **Rate Limiting**
   - Track via user_events table
   - IP-based for anonymous reads
   - User-based for authenticated writes

## Performance Optimizations

1. **Spatial Indexes**
   - GIST index for PostGIS queries
   - B-tree index for geohash prefix matching

2. **Query Strategies**
   - Geohash for rough filtering
   - PostGIS for precise distance calculations
   - Limit results based on zoom level

3. **Caching Opportunities**
   - Public pin reads cacheable
   - User profile cacheable
   - JWT validation cacheable

## Deployment Considerations

1. **Environment Variables**
   ```
   DATABASE_URL=postgresql://...
   JWT_PRIVATE_KEY_PATH=/path/to/key
   JWT_PUBLIC_KEY_PATH=/path/to/key.pub
   GOOGLE_CLIENT_ID=...
   GOOGLE_CLIENT_SECRET=...
   OAUTH_STATE_KEY=base64-encoded-32-bytes
   ```

2. **Database Requirements**
   - PostgreSQL 14+ with PostGIS extension
   - GIST index support

3. **Scaling Strategy**
   - Stateless services (horizontal scaling)
   - Read replicas for pin queries
   - CDN for static assets