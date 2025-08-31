# Reactions Service

**Owner:** feature/reactions-system  
**Priority:** Phase 2 - Can be developed in parallel

## Dependencies
- Posts service (read-only dependency)
- Requires `posts` table to exist

## Database Tables
- `post_reactions` - Reaction records
- `reaction_activity` - Activity tracking (optional)
- `post_reaction_counts` - Materialized view for performance

## Proto Files
- `proto/v1/entities/reaction.proto`
- `proto/v1/service/reactions.proto`

## Key Responsibilities
- Multiple reaction types (like, love, wow, etc.)
- Prevent duplicate reactions
- Batch operations for feed
- Activity tracking

## Merge Notes
Can be merged in any order after posts-core is merged.