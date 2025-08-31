# Feed Service

**Owner:** feature/location-feed  
**Priority:** Phase 2 - Can be developed in parallel

## Dependencies
- Posts service (heavy dependency)
- Post locations data
- Comments service (optional, for counts)
- Reactions service (optional, for counts)

## Database Tables
- `feed_sessions` - Feed session tracking
- `feed_hot_zones` - Cache for hot areas

## Proto Files
- `proto/v1/service/feed.proto`

## Key Responsibilities
- Location-based feed queries
- Multiple query strategies (bbox, radius, geohash)
- Feed refresh/new post detection
- Trending/popular algorithms
- Session management

## Merge Notes
Can be merged in any order after posts-core is merged.
Will integrate with comments/reactions if available.