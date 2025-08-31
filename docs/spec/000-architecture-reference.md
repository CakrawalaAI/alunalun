# Architecture Reference

## Overview

This document describes the architectural patterns and design decisions for the Alunalun application, which follows **Feature Sliced Design (FSD)** on the frontend and **Clean Architecture** on the backend, connected through **ConnectRPC** with **Protocol Buffers** for API contracts and **sqlc** for type-safe database access.

## Frontend Architecture: Feature Sliced Design

### Core Principles

Feature Sliced Design is a methodology for organizing frontend applications into layers and slices, promoting:
- **Standardization**: Consistent project structure across teams
- **Scalability**: Easy to add new features without touching existing code
- **Isolation**: Features are independent and can be developed in parallel
- **Explicit dependencies**: Clear data flow and dependencies between modules

### Layer Structure

```
web/src/
├── app/                 # Application initialization and providers
│   ├── routes/         # Route definitions
│   ├── styles/         # Global styles
│   └── main.tsx        # Application entry point
├── pages/              # Route pages (compositional layer)
├── features/           # Business features (isolated domains)
│   ├── auth/
│   ├── comments/
│   ├── feed/
│   ├── map/
│   ├── map-pins/
│   ├── posts/
│   ├── reactions/
│   └── user/
├── blocks/             # Complex UI blocks (reusable across pages)
│   └── interactive-map/
├── common/             # Shared utilities and infrastructure
│   ├── connectrpc/     # Generated RPC types
│   ├── lib/           # Utility functions
│   ├── logger/        # Logging infrastructure
│   ├── services/      # API client setup
│   └── utils/         # Common utilities
└── services/          # External service clients
    └── clients/       # Feature-specific API clients
```

### Feature Slice Anatomy

Each feature follows a consistent internal structure:

```
features/auth/
├── components/         # UI components specific to the feature
│   ├── auth-prompt.tsx
│   ├── google-auth-button.tsx
│   └── username-modal.tsx
├── hooks/             # Custom React hooks
│   ├── use-auth.ts
│   └── use-authenticate.ts
├── store/             # State management (Zustand)
│   └── auth-store.ts
├── lib/               # Feature-specific utilities
│   └── token-utils.ts
├── types/             # TypeScript type definitions
│   └── index.ts
└── index.ts           # Public API exports
```

### Key Rules

1. **Import Direction**: Higher layers can import from lower layers, never vice versa
   - `pages` → `features` → `common`
   - `features` cannot import from `pages`
   - Features should not directly import from each other

2. **Public API**: Each feature exposes only necessary exports through `index.ts`

3. **Isolation**: Features are self-contained with their own components, hooks, and state

## Backend Architecture: Clean Architecture

### Core Principles

The backend follows Clean Architecture (Hexagonal Architecture) principles:
- **Separation of Concerns**: Business logic is independent of frameworks and databases
- **Dependency Inversion**: High-level modules don't depend on low-level modules
- **Testability**: Business logic can be tested without external dependencies
- **Maintainability**: Clear boundaries between layers

### Layer Structure

```
api/
├── cmd/
│   └── api/
│       └── main.go          # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   │   └── config.go
│   ├── middleware/          # HTTP/RPC middleware
│   │   └── auth.go
│   ├── protocgen/           # Generated protobuf code
│   │   └── v1/
│   │       ├── common/
│   │       ├── entities/
│   │       └── service/
│   ├── repository/          # Data access layer (sqlc generated)
│   │   ├── db.go           # Database connection
│   │   ├── models.go       # sqlc models
│   │   └── *.sql.go        # sqlc generated queries
│   ├── services/            # Business logic layer
│   │   ├── auth/
│   │   ├── comments/
│   │   ├── feed/
│   │   ├── map/
│   │   ├── media/
│   │   ├── posts/
│   │   └── reactions/
│   └── server/              # Server setup and routing
│       └── server.go
├── proto/                   # Internal proto definitions
├── sql/
│   ├── migrations/          # Database migrations
│   └── queries/            # SQL queries for sqlc
├── buf.gen.yaml            # Buf code generation config
├── buf.yaml                # Buf configuration
└── sqlc.yaml               # sqlc configuration
```

### Service Layer Pattern

Each service implements the ConnectRPC service interface:

```go
// services/posts/service.go
type Service struct {
    db *repository.Queries
    // other dependencies
}

func NewService(db *repository.Queries) *Service {
    return &Service{db: db}
}

func (s *Service) CreatePost(
    ctx context.Context,
    req *connect.Request[postsv1.CreatePostRequest],
) (*connect.Response[postsv1.CreatePostResponse], error) {
    // Business logic here
}
```

## Protocol Buffers & ConnectRPC

### Proto Organization

Protocol Buffer definitions serve as the single source of truth for API contracts:

```
proto/v1/
├── entities/           # Shared data models
│   ├── user.proto
│   ├── post.proto
│   ├── comment.proto
│   └── reaction.proto
├── service/           # Service definitions
│   ├── auth.proto
│   ├── posts.proto
│   ├── comments.proto
│   ├── reactions.proto
│   ├── feed.proto
│   └── map_pins.proto
└── common/           # Shared messages
    └── pagination.proto
```

### Service Definition Example

```protobuf
// proto/v1/service/posts.proto
syntax = "proto3";

package alunalun.v1.service;

import "v1/entities/post.proto";
import "v1/common/pagination.proto";

service PostsService {
  rpc CreatePost(CreatePostRequest) returns (CreatePostResponse);
  rpc GetPost(GetPostRequest) returns (GetPostResponse);
  rpc ListPosts(ListPostsRequest) returns (ListPostsResponse);
}
```

### Code Generation Pipeline

1. **Backend Generation** (via Buf):
   ```yaml
   # buf.gen.yaml
   version: v2
   plugins:
     - remote: buf.build/protocolbuffers/go
       out: internal/protocgen
     - remote: buf.build/connectrpc/go
       out: internal/protocgen
   ```

2. **Frontend Generation**:
   - Types are generated to `web/src/common/connectrpc/v1/`
   - Provides type-safe client interfaces

## Database Layer: sqlc

### Query-First Development

SQL queries are written first, then Go code is generated:

```sql
-- sql/queries/posts.sql
-- name: CreatePost :one
INSERT INTO posts (
    user_id, content, location_point, location_name
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetPostByID :one
SELECT * FROM posts WHERE id = $1;
```

### Generated Type-Safe Code

sqlc generates type-safe Go code from SQL queries:

```go
// repository/posts.sql.go (generated)
type CreatePostParams struct {
    UserID       uuid.UUID
    Content      string
    LocationPoint pgtype.Point
    LocationName string
}

func (q *Queries) CreatePost(ctx context.Context, arg CreatePostParams) (Post, error) {
    // Generated implementation
}
```

### Configuration

```yaml
# sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "sql/queries"
    schema: "sql/migrations"
    gen:
      go:
        package: "repository"
        out: "api/internal/repository"
        emit_json_tags: true
        emit_interface: true
```

## Data Flow

### Request Lifecycle

1. **Client Request** (Frontend)
   ```typescript
   // services/clients/posts.ts
   const client = new PostsServiceClient(transport);
   const response = await client.createPost(request);
   ```

2. **RPC Transport** (ConnectRPC)
   - HTTP/2 with binary protobuf or JSON
   - Built-in streaming support
   - Type-safe contract enforcement

3. **Server Handler** (Backend)
   ```go
   // Server receives typed request
   func (s *Service) CreatePost(ctx, req) (res, error) {
       // Validate request
       // Execute business logic
       // Call repository layer
   }
   ```

4. **Database Operation** (sqlc)
   ```go
   post, err := s.db.CreatePost(ctx, repository.CreatePostParams{
       UserID:  userID,
       Content: req.Msg.Content,
   })
   ```

5. **Response** (Protobuf)
   - Structured response sent back through ConnectRPC
   - Automatically serialized/deserialized

## Authentication Flow

### Token-Based Auth

1. **Anonymous Users**: Automatically receive a JWT for basic access
2. **Authenticated Users**: OAuth providers (Google) with username selection
3. **Token Storage**: Secure HTTP-only cookies + localStorage backup
4. **Middleware**: Validates tokens and attaches user context

### Auth State Management

```typescript
// Frontend auth store (Zustand)
interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isAnonymous: boolean;
  login: (provider: string) => Promise<void>;
  logout: () => void;
}
```

## Best Practices

### Frontend

1. **Component Composition**: Build complex UIs from simple, reusable components
2. **State Management**: Use Zustand stores sparingly, prefer local state
3. **API Calls**: Centralize in service clients, handle errors consistently
4. **Type Safety**: Leverage generated types from protobuf definitions

### Backend

1. **Service Isolation**: Each service owns its domain logic
2. **Repository Pattern**: All database access through sqlc-generated code
3. **Error Handling**: Use Connect error codes for consistent API errors
4. **Context Propagation**: Pass context through all layers for cancellation/tracing

### Cross-Cutting Concerns

1. **Logging**: Structured logging with correlation IDs
2. **Monitoring**: Metrics and traces for observability
3. **Security**: Input validation, rate limiting, CORS configuration
4. **Testing**: Unit tests for business logic, integration tests for APIs

## Migration Strategy

### Database Migrations

```sql
-- sql/migrations/001_initial.sql
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### API Versioning

- Proto packages include version (e.g., `alunalun.v1.service`)
- Breaking changes require new version
- Backward compatibility maintained through proto evolution rules

## Development Workflow

1. **Define Proto Contract**: Start with API design in `.proto` files
2. **Generate Code**: Run `buf generate` for both frontend and backend
3. **Write SQL Queries**: Define database operations in `.sql` files
4. **Generate Repository**: Run `sqlc generate` for type-safe DB code
5. **Implement Service**: Write business logic in service layer
6. **Build UI**: Create feature slices with components and state
7. **Test End-to-End**: Verify full data flow from UI to database

## Tooling

- **Buf**: Protocol buffer toolchain for linting and generation
- **sqlc**: Generate type-safe Go code from SQL
- **ConnectRPC**: Modern RPC framework built on HTTP/2
- **Bun**: JavaScript runtime and package manager
- **Vite**: Frontend build tool
- **Zustand**: Lightweight state management
- **React Query/Tanstack Query**: Server state management (planned)

## Future Considerations

1. **GraphQL Gateway**: Potential GraphQL layer over ConnectRPC services
2. **Event Sourcing**: For audit trails and complex state transitions
3. **Microservices**: Service layer can be split into separate deployments
4. **Real-time Updates**: WebSocket or Server-Sent Events for live features
5. **Caching Layer**: Redis for session management and query caching