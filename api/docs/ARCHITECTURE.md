# Backend Architecture

## Core Principles

### 1. Contract-First AND Data-Model-First Development
- **Proto definitions** define the API contract and service interfaces
- **Migration schemas** define the data model and database structure
- Both drive the business logic implementation in the middle layer
- Business logic adapts to both constraints, not the other way around

### 2. Clean Architecture with Pragmatism
- **Repository Layer**: SQLC-generated code provides type-safe database access
- **Service Layer**: Business logic organized by features
- **Transport Layer**: ConnectRPC handles HTTP/gRPC communication
- **Pragmatic approach**: Not dogmatic about separation when it adds no value

### 3. Feature-Sliced Services
Services are organized by business domains:
```
internal/services/
├── posts/       # Posts and location data
├── comments/    # Hierarchical comment threads
├── reactions/   # User engagement actions
├── feed/        # Location-based content feeds
├── map/         # Map pin clustering and display
└── media/       # Media upload and processing
```

## Directory Structure

```
api/
├── cmd/                    # Application entry points
│   └── server/
│       └── main.go        # Main server binary
│
├── internal/              # Private application code
│   ├── protocgen/         # Generated protobuf/ConnectRPC code
│   │   └── v1/           # Version 1 API
│   │       ├── entities/ # Domain entities
│   │       └── service/  # Service interfaces
│   │
│   ├── repository/        # SQLC generated database layer
│   │   ├── db.go         # Database connection
│   │   ├── models.go     # Generated models
│   │   └── *.sql.go      # Generated queries
│   │
│   ├── services/          # Business logic layer
│   │   └── {feature}/    # Feature-specific logic
│   │       ├── service.go     # Service implementation
│   │       ├── handlers.go    # Request handlers
│   │       └── validation.go  # Input validation
│   │
│   ├── middleware/        # Cross-cutting concerns
│   │   ├── auth.go       # Authentication
│   │   ├── cors.go       # CORS handling
│   │   └── logging.go    # Request logging
│   │
│   └── server/           # Server setup
│       ├── server.go     # Service registration
│       └── routes.go     # Route configuration
│
├── proto/                # Protocol Buffer definitions
│   └── v1/
│       ├── entities/     # Shared domain models
│       └── service/      # Service definitions
│
├── sql/                  # Database schema and queries
│   ├── migrations/       # Sequential migrations
│   └── queries/         # SQLC query definitions
│
├── docs/                # Backend documentation
├── buf.yaml            # Buf configuration
├── buf.gen.yaml        # Buf generation config
└── sqlc.yaml          # SQLC configuration
```

## Data Flow

1. **Request Reception**: ConnectRPC receives typed request
2. **Middleware Processing**: Auth, logging, rate limiting
3. **Service Handler**: Business logic processing
4. **Repository Layer**: Type-safe database operations via SQLC
5. **Response Formation**: Structured protobuf response

## Service Isolation Rules

### What Services CAN Do:
- Access their own repository queries
- Call shared middleware functions
- Emit events for other services to consume
- Return well-typed errors

### What Services CANNOT Do:
- Import code from other services directly
- Modify database tables they don't own
- Access other services' private methods
- Bypass the repository layer for database access

## Database Layer Architecture

### Migration Ownership
Each feature owns specific migration numbers:
- 004: Posts and locations
- 005: Comments system
- 006: Reactions
- 007: Feed optimizations
- 008: Map optimizations
- 009: Media handling

### Query Organization
Queries are feature-sliced in `sql/queries/`:
- `posts.sql`: Post CRUD and location queries
- `comments.sql`: Comment threading and retrieval
- `reactions.sql`: Reaction aggregation
- `feed.sql`: Feed generation queries
- `map_pins.sql`: Clustering and spatial queries

### Repository Pattern
- All database access through SQLC-generated code
- No raw SQL in service layer
- Type-safe query parameters and results
- Automatic NULL handling via pointers

## Integration Points

### Service Registration (server.go)
```go
// [FEATURE-SERVICE-START]
// Each feature registers here
// [FEATURE-SERVICE-END]
```

### Route Configuration (routes.go)
```go
// [FEATURE-ROUTES-START]
// Each feature adds routes here
// [FEATURE-ROUTES-END]
```

### SQLC Configuration (sqlc.yaml)
```yaml
# [FEATURE-QUERIES]
# Each feature adds query files
```

## Error Handling

### Error Types
1. **Validation Errors**: 400 Bad Request
2. **Authentication Errors**: 401 Unauthorized
3. **Authorization Errors**: 403 Forbidden
4. **Not Found Errors**: 404 Not Found
5. **Business Logic Errors**: 422 Unprocessable Entity
6. **System Errors**: 500 Internal Server Error

### Error Propagation
- Repository → Service: Database errors wrapped with context
- Service → Handler: Business errors with user-friendly messages
- Handler → Client: ConnectRPC error codes with details

## Performance Considerations

### Database Optimization
- Prepared statements via SQLC
- Connection pooling with pgx
- Proper indexing per feature requirements
- Query optimization in SQLC definitions

### Caching Strategy
- No premature caching
- Add caching only after identifying bottlenecks
- Cache at service layer, not repository layer

## Security Principles

### Input Validation
- Proto field validation rules
- Service-layer business validation
- SQL injection prevention via SQLC

### Authentication & Authorization
- JWT tokens for session management
- Middleware-based auth checks
- Service-level authorization logic

## Testing Strategy

### Unit Tests
- Service logic testing with mocked repositories
- Pure function testing
- Validation rule testing

### Integration Tests
- Full request/response cycle
- Database interaction testing
- Multi-service interaction testing

## Deployment Considerations

### Configuration
- Environment variables for secrets
- Config files for non-sensitive settings
- Feature flags for gradual rollouts

### Observability
- Structured logging with correlation IDs
- Metrics collection (Prometheus-ready)
- Distributed tracing support

## Development Workflow

1. Define/update proto contracts
2. Write SQL queries and migrations
3. Generate code: `make proto && make sqlc`
4. Implement service logic
5. Write tests
6. Update integration points with markers

## Best Practices

1. **Keep services small and focused**
2. **Use generated code for type safety**
3. **Don't bypass the repository layer**
4. **Handle errors explicitly**
5. **Log with context and correlation IDs**
6. **Write tests for business logic**
7. **Document complex algorithms**
8. **Use feature flags for risky changes**