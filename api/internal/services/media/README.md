# Media Service

**Owner:** feature/media-upload  
**Priority:** Phase 2 - Can be developed in parallel

## Dependencies
- Posts service (uses post_media table)
- S3/Storage backend

## Database Tables
- `post_media` (from posts feature)
- `media_uploads` - Upload tracking

## Proto Files
- `proto/v1/entities/media.proto`
- `proto/v1/service/media.proto`

## Key Responsibilities
- File upload to S3/CDN
- Image optimization
- Video thumbnail generation
- Media validation
- URL generation

## Merge Notes
Can be merged in any order after posts-core is merged.
Enhances posts with media handling.