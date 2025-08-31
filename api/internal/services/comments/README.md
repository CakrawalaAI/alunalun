# Comments Service

**Owner:** feature/comments-system  
**Priority:** Phase 2 - Can be developed in parallel

## Dependencies
- Posts service (read-only dependency)
- Requires `posts` table to exist

## Database Tables
- `post_comments` - Hierarchical comments with soft delete

## Proto Files
- `proto/v1/entities/comment.proto`
- `proto/v1/service/comments.proto`

## Key Responsibilities
- Threaded comment system
- Soft delete to preserve threads
- Reply management
- Comment moderation

## Merge Notes
Can be merged in any order after posts-core is merged.