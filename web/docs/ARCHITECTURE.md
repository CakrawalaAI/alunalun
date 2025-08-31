# Frontend Architecture

## Core Principles

### 1. Feature-Sliced Design (FSD)
- **Standardization**: Consistent structure across all features
- **Isolation**: Features are independent and self-contained
- **Explicit Dependencies**: Clear data flow between layers
- **Scalability**: Easy to add features without touching existing code

### 2. No Cross-Feature Imports
- Features CANNOT import from sibling features
- Composition happens only at higher layers (blocks/pages)
- Shared code lives in common/
- Low coupling, high cohesion

### 3. Layer Hierarchy
```
pages → blocks → features → common
```
- Higher layers can import from lower layers
- Lower layers cannot import from higher layers
- Features are siblings and cannot import from each other

## Directory Structure

```
web/
├── src/
│   ├── app/                    # Application initialization
│   │   ├── routes/            # Route definitions
│   │   ├── layout.tsx         # Root layout with providers
│   │   └── page.tsx          # Root page component
│   │
│   ├── pages/                 # Route pages (composition layer)
│   │   ├── home/             # Home page
│   │   ├── explore/          # Explore page
│   │   └── profile/          # Profile page
│   │
│   ├── blocks/                # Complex UI blocks
│   │   └── interactive-map/  # Map with features composed
│   │       ├── index.tsx
│   │       └── map-with-pins.tsx
│   │
│   ├── features/              # Business features (isolated)
│   │   ├── posts/            # Posts functionality
│   │   ├── comments/         # Comments system
│   │   ├── reactions/        # Reactions/engagement
│   │   ├── feed/            # Location-based feed
│   │   ├── map-pins/        # Map pin management
│   │   └── media/           # Media upload/display
│   │
│   └── common/               # Shared utilities
│       ├── clients/         # API clients (RPC)
│       ├── connectrpc/      # Generated proto types
│       ├── components/      # Shared UI components
│       ├── hooks/          # Shared React hooks
│       └── utils/          # Utility functions
│
├── docs/                     # Frontend documentation
├── package.json             # Dependencies
├── tsconfig.json           # TypeScript config
└── vite.config.ts         # Build configuration
```

## Feature Anatomy

Each feature follows this structure:
```
features/posts/
├── components/           # UI components
│   ├── post-card.tsx    # Display component
│   └── post-form.tsx    # Input component
├── hooks/               # Feature-specific hooks
│   └── use-posts.ts    # Data fetching/mutations
├── store/              # State management (Zustand)
│   └── posts-store.ts  # Local state
├── types/              # TypeScript definitions
│   └── index.ts       # Feature types
├── lib/               # Feature utilities
│   └── validators.ts  # Input validation
└── index.ts          # Public API exports
```

## Import Rules

### ✅ ALLOWED Imports

**Pages can import from:**
- blocks/
- features/
- common/

**Blocks can import from:**
- features/
- common/

**Features can import from:**
- common/
- Their own subdirectories

**Common can import from:**
- Other common modules
- External packages

### ❌ FORBIDDEN Imports

**Features CANNOT import from:**
- Other features (siblings)
- blocks/
- pages/

**Common CANNOT import from:**
- features/
- blocks/
- pages/

**Blocks CANNOT import from:**
- pages/
- Other blocks (without explicit reason)

## State Management

### Client State (Zustand)
- Feature-local state in feature stores
- Shared UI state in common stores (sparingly)
- No cross-feature state sharing

### Server State (React Query)
- Managed through hooks in features
- Cache configuration per feature
- Optimistic updates at feature level

### State Boundaries
```typescript
// ✅ GOOD: Feature owns its state
// features/posts/store/posts-store.ts
interface PostsState {
  selectedPostId: string | null;
  isCreating: boolean;
}

// ❌ BAD: Cross-feature state reference
interface PostsState {
  selectedPostId: string | null;
  commentsVisible: boolean; // Should be in comments feature
}
```

## Composition Patterns

### Page-Level Composition
```typescript
// pages/home/index.tsx
import { PostList } from '@/features/posts';
import { CommentThread } from '@/features/comments';
import { ReactionBar } from '@/features/reactions';

// Compose features at page level
export function HomePage() {
  return (
    <div>
      <PostList />
      <CommentThread postId={selectedId} />
      <ReactionBar targetId={selectedId} />
    </div>
  );
}
```

### Block-Level Composition
```typescript
// blocks/interactive-map/index.tsx
import { MapRenderer } from '@/features/map';
import { PinManager } from '@/features/map-pins';
import { LocationFeed } from '@/features/feed';

// Compose related features in a block
export function InteractiveMap() {
  return (
    <MapContainer>
      <MapRenderer />
      <PinManager />
      <LocationFeed />
    </MapContainer>
  );
}
```

## API Client Organization

### Client Structure
```
common/clients/
├── posts.ts        # Posts service client
├── comments.ts     # Comments service client
├── reactions.ts    # Reactions service client
├── feed.ts        # Feed service client
└── index.ts       # Client exports
```

### Client Usage
```typescript
// features/posts/hooks/use-posts.ts
import { postsClient } from '@/common/clients';

export function usePosts() {
  return useQuery({
    queryKey: ['posts'],
    queryFn: () => postsClient.listPosts({}),
  });
}
```

## Type System

### Generated Types
- Proto-generated types in `common/connectrpc/`
- Imported by features as needed
- Single source of truth for API types

### Feature Types
- Local types defined in feature
- Extend or wrap API types as needed
- No cross-feature type dependencies

## Performance Patterns

### Code Splitting
- Features are lazy-loaded by default
- Routes use dynamic imports
- Heavy components loaded on demand

### Optimization
```typescript
// Lazy load heavy features
const MapView = lazy(() => import('@/features/map'));

// Memoize expensive computations
const processedData = useMemo(() => 
  processData(rawData), [rawData]
);
```

## Testing Strategy

### Unit Tests
- Component testing with Testing Library
- Hook testing with renderHook
- Pure function testing

### Integration Tests
- Feature-level integration tests
- Mock API responses at client level
- Test feature in isolation

## Error Handling

### API Errors
- Handled at feature level
- User-friendly messages
- Fallback UI components

### Runtime Errors
- Error boundaries at feature level
- Graceful degradation
- Logging to monitoring service

## Development Workflow

1. **Feature Development**
   - Work in isolated feature directory
   - Export only public API through index.ts
   - No cross-feature imports

2. **Integration**
   - Compose features in pages or blocks
   - Handle inter-feature communication via props/events
   - Test integration at page level

3. **Code Review Checklist**
   - ✓ No cross-feature imports
   - ✓ Public API clearly defined
   - ✓ State properly scoped
   - ✓ Types from generated sources
   - ✓ Error handling in place

## Best Practices

1. **Keep features small and focused**
2. **Export minimal public API**
3. **Use TypeScript strictly**
4. **Prefer composition over inheritance**
5. **Handle loading and error states**
6. **Memoize expensive operations**
7. **Use semantic HTML**
8. **Follow accessibility guidelines**
9. **Write tests for critical paths**
10. **Document complex logic**

## Anti-Patterns to Avoid

### ❌ Cross-Feature Imports
```typescript
// BAD: Feature importing from another feature
// features/comments/components/comment.tsx
import { PostCard } from '@/features/posts';
```

### ❌ Feature State in Common
```typescript
// BAD: Feature-specific state in common
// common/stores/app-store.ts
interface AppState {
  posts: Post[];  // Should be in posts feature
}
```

### ❌ Business Logic in Pages
```typescript
// BAD: Business logic at page level
// pages/home/index.tsx
function HomePage() {
  // This should be in a feature
  const processedPosts = posts.filter(...).map(...);
}
```

### ❌ Direct API Calls
```typescript
// BAD: Direct API call without client
// features/posts/hooks/use-posts.ts
const response = await fetch('/api/posts');
```

## Migration Guidelines

When adding new features:
1. Create feature directory structure
2. Define public API in index.ts
3. Implement components, hooks, and store
4. Add client in common/clients/
5. Compose in pages or blocks
6. Add tests
7. Update documentation