# Posts Service

**Owner:** feature/posts-core  
**Priority:** Phase 1 - Must be completed first

## Dependencies
- None (base feature)

## Database Tables
- `posts` - Core posts table
- `post_locations` - Location data with PostGIS
- `post_media` - Media attachments

## Proto Files
- `proto/v1/entities/post.proto`
- `proto/v1/service/posts.proto`

## Key Responsibilities
- Post CRUD operations
- Location tagging
- Media attachment handling
- Author management (user/anonymous)

## Merge Notes
This feature must be merged first as it provides the foundation for all other features.