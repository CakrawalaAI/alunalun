# Architecture Decisions and Principles

## Executive Summary

This document captures the architectural decisions and principles for the Alunalun project, establishing a foundation for parallel development by 5+ teams and AI coding agents.

## Core Architectural Principles

### 1. Monorepo with Clear Boundaries

**Decision**: Two top-level directories only: `api/` and `web/`

**Rationale**:
- Clear separation of backend and frontend concerns
- Each directory is self-contained with its own tooling
- Simplifies deployment and CI/CD pipelines
- Prevents confusion about where code belongs

**Structure**:
```
alunalun/
├── api/          # All backend concerns
├── web/          # All frontend concerns
├── docs/         # Development reference (for architects)
├── data/         # Local development data
└── worktrees/    # Git worktrees for parallel development
```

### 2. Contract-First AND Data-Model-First Development

**Decision**: Both API contracts (protobuf) and data models (SQL migrations) drive development

**Rationale**:
- Protobuf defines the API surface and service boundaries
- SQL migrations define the data structure and relationships
- Business logic adapts to both constraints
- Prevents impedance mismatch between API and database

**Implementation**:
1. Define proto contracts for service interfaces
2. Define SQL migrations for data structure
3. Generate code from both (SQLC + protobuf)
4. Implement business logic that satisfies both contracts

### 3. Feature-Sliced Architecture

**Decision**: Both frontend and backend organized by business features

**Frontend** (Feature-Sliced Design):
```
web/src/features/
├── posts/
├── comments/
├── reactions/
├── feed/
├── map-pins/
└── media/
```

**Backend** (Feature-Sliced Services):
```
api/internal/services/
├── posts/
├── comments/
├── reactions/
├── feed/
├── map/
└── media/
```

**Rationale**:
- Enables parallel development with minimal conflicts
- Clear ownership boundaries
- Features can evolve independently
- Easy to understand and maintain

### 4. No Cross-Feature Dependencies

**Decision**: Features cannot directly depend on each other

**Frontend Rule**:
- Features cannot import from sibling features
- Composition only at pages/blocks level
- Shared code in common/

**Backend Rule**:
- Services cannot import from other services
- Communication through events or orchestration layer
- Shared code in middleware/

**Rationale**:
- Prevents circular dependencies
- Enables independent deployment
- Simplifies testing
- Reduces cognitive load

### 5. Generated Code as Foundation

**Decision**: Rely heavily on code generation for type safety

**Generators**:
- **SQLC**: Database queries → Type-safe Go code
- **Protobuf**: API definitions → Go and TypeScript code
- **ConnectRPC**: Service definitions → Client/server code

**Rationale**:
- Single source of truth
- Type safety across boundaries
- Reduces boilerplate
- Prevents drift between layers

## Directory Organization Principles

### Backend (api/)

**Principle**: All backend concerns under `api/`

```
api/
├── cmd/           # Entry points
├── internal/      # Private code
├── proto/         # API contracts
├── sql/           # Database schema
├── docs/          # Backend docs
├── buf.yaml       # Proto config
├── buf.gen.yaml   # Generation config
└── sqlc.yaml      # SQL generation config
```

**Key Decisions**:
- Config files (`*.yaml`) in api/ root for tooling simplicity
- `internal/` prevents accidental exports
- `proto/` and `sql/` co-located with backend code

### Frontend (web/)

**Principle**: Strict 5-directory structure in src/

```
web/src/
├── app/          # Application setup
├── pages/        # Route pages
├── blocks/       # Complex UI blocks
├── features/     # Business features
└── common/       # Shared utilities
```

**Key Decisions**:
- No other directories at src/ root
- Features are isolated modules
- Blocks compose features
- Pages compose blocks and features

## Integration Points and Merge Strategies

### Designated Merge Points

Files with known conflicts have designated sections:

1. **api/internal/server/server.go**
   ```go
   // [FEATURE-SERVICE-START]
   // Feature-specific service registration
   // [FEATURE-SERVICE-END]
   ```

2. **api/sqlc.yaml**
   ```yaml
   # [FEATURE-QUERIES]
   # Feature-specific query files
   ```

3. **web/package.json**
   ```json
   // [FEATURE-DEPS-START]
   // Feature-specific dependencies
   // [FEATURE-DEPS-END]
   ```

### Merge Conflict Resolution

**Strategy**: Additive merging
- Always keep both sides of conflicts in designated sections
- Never remove another feature's code
- Order doesn't matter in most cases
- Use markers to identify ownership

## Parallel Development Enablement

### Team Boundaries

Each team owns:
1. **Backend service directory**: `api/internal/services/{feature}/`
2. **Frontend feature directory**: `web/src/features/{feature}/`
3. **Proto definitions**: `api/proto/v1/service/{feature}.proto`
4. **SQL files**: `api/sql/queries/{feature}.sql`
5. **Migration number**: Reserved in `api/sql/migrations/README.md`

### Worktree Strategy

```bash
# Each team works in a worktree
worktrees/
├── feature-posts-core/
├── feature-comments-system/
├── feature-reactions-system/
├── feature-location-feed/
└── feature-map-pins/
```

### Development Phases

**Phase 1**: Posts Core (Foundation)
- Must complete first
- Other features depend on this
- Establishes base data model

**Phase 2**: All Other Features (Parallel)
- Can develop simultaneously
- No inter-dependencies
- Merge in any order

## Technology Stack Decisions

### Backend Stack
- **Language**: Go
- **RPC Framework**: ConnectRPC
- **Database**: PostgreSQL with PostGIS
- **ORM**: None (SQLC for queries)
- **Serialization**: Protocol Buffers

### Frontend Stack
- **Framework**: React 19
- **Build Tool**: Vite
- **Package Manager**: Bun
- **State Management**: Zustand (client), React Query (server)
- **Styling**: Tailwind CSS
- **Type System**: TypeScript

### Infrastructure
- **Container**: Docker
- **Orchestration**: Docker Compose (dev)
- **CI/CD**: GitHub Actions
- **Monitoring**: To be determined

## Performance Principles

### Backend Performance
- Prepared statements via SQLC
- Connection pooling
- Proper database indexing
- No N+1 queries

### Frontend Performance
- Code splitting by feature
- Lazy loading of routes
- Optimistic updates
- Image optimization

## Security Principles

### Input Validation
- Proto field validation
- Service-layer validation
- SQLC prevents SQL injection
- Frontend validation for UX

### Authentication
- JWT tokens
- HTTP-only cookies
- Refresh token rotation
- Anonymous user support

## Testing Strategy

### Backend Testing
- Unit tests for business logic
- Integration tests for API endpoints
- Repository tests with test database
- No mocking of database

### Frontend Testing
- Component testing with Testing Library
- Hook testing in isolation
- Integration tests at feature level
- E2E tests for critical paths

## Documentation Standards

### Code Documentation
- Document WHY, not WHAT
- Complex algorithms need explanation
- API endpoints need examples
- Keep docs close to code

### Architecture Documentation
- `api/docs/`: Backend architecture
- `web/docs/`: Frontend architecture
- `docs/spec/`: System specifications
- `README.md`: Getting started

## Build and Deployment

### Build Configuration
- Build configs at appropriate level
- `api/`: Backend build configs
- `web/`: Frontend build configs
- Root: Orchestration only (Makefile)

### Deployment Strategy
- Single binary for backend
- Static files for frontend
- Environment-based configuration
- Feature flags for gradual rollout

## Future Considerations

### Potential Evolution
1. **Microservices**: Services can be extracted
2. **GraphQL**: Can be added as a layer
3. **Real-time**: WebSocket support ready
4. **Mobile**: API-first enables mobile clients

### Technical Debt Management
- Regular refactoring windows
- Deprecation notices
- Migration guides
- Breaking change process

## Decision Record

| Decision | Date | Rationale | Revisit |
|----------|------|-----------|---------|
| Monorepo structure | 2024-01 | Simplify development | 2024-07 |
| Contract-first | 2024-01 | Type safety | Stable |
| Feature slices | 2024-01 | Parallel development | Stable |
| No cross-deps | 2024-01 | Maintainability | Stable |
| SQLC + Proto | 2024-01 | Code generation | 2024-06 |

## Principles for AI Agents

### AI-Friendly Architecture
1. **Clear file ownership**: Each file has one owner
2. **Obvious patterns**: Consistent structure across features
3. **Merge markers**: Explicit sections for additions
4. **Type safety**: Generated types prevent errors
5. **Test coverage**: Confidence in changes

### AI Agent Guidelines
- Work within assigned feature boundaries
- Use merge markers for shared files
- Run code generation after changes
- Follow existing patterns
- Test before committing

## Summary

This architecture is optimized for:
- **Parallel development** by multiple teams
- **AI agent compatibility** with clear boundaries
- **Type safety** through code generation
- **Maintainability** through isolation
- **Scalability** through feature slicing

The key insight is that both **contract-first** (API) and **data-model-first** (database) approaches are combined, with business logic adapting to both constraints rather than driving them.