# Migration Number Reservations

Each feature branch MUST use their reserved migration number.
DO NOT change numbers that are already reserved.

## Reserved Migration Numbers

| Number | Feature | Branch | Owner | Status |
|--------|---------|--------|-------|--------|
| 001 | Initial setup | main | - | ✅ Merged |
| 002 | Users table | main | - | ✅ Merged |
| 003 | Anonymous sessions | main | - | ✅ Merged |
| 004 | Posts and locations | feature/posts-core | @posts-team | 🔒 Reserved |
| 005 | Comments | feature/comments-system | @comments-team | 🔒 Reserved |
| 006 | Reactions | feature/reactions-system | @reactions-team | 🔒 Reserved |
| 007 | Feed optimizations | feature/location-feed | @feed-team | 🔒 Reserved |
| 008 | Map optimizations | feature/map-pins | @map-team | 🔒 Reserved |
| 009 | Media handling | feature/media-upload | @media-team | 🔒 Reserved |
| 010-020 | (Available) | - | - | ✅ Available |

## Merge Order Requirements

### Phase 1 - Foundation (MUST merge first)
1. **feature/posts-core** (004)
   - Creates base tables: posts, post_locations, post_media
   - All other features depend on this

### Phase 2 - Parallel Features (merge in ANY order)
- **feature/comments-system** (005)
  - Adds: post_comments table
  - Depends on: posts table
  
- **feature/reactions-system** (006)
  - Adds: post_reactions, reaction_activity tables
  - Depends on: posts table
  
- **feature/location-feed** (007)
  - Adds: feed_sessions, feed_hot_zones tables
  - Depends on: posts, post_locations tables
  
- **feature/map-pins** (008)
  - Adds: map_pin_clusters, user_map_preferences tables
  - Depends on: posts, post_locations tables
  
- **feature/media-upload** (009)
  - Adds: media_uploads table
  - Depends on: post_media table

## Conflict Resolution

If you encounter migration number conflicts:
1. DO NOT renumber existing migrations
2. Check this file for the correct reservation
3. Use the next available number if needed (010+)
4. Update this file with your new reservation

## Creating Your Migration

```bash
# Use your reserved number
touch sql/migrations/00X_your_feature.sql

# Add header to your migration
echo "-- Migration: 00X_your_feature.sql
-- Owner: feature/your-branch
-- Dependencies: List any required tables
-- Description: What this migration does
" > sql/migrations/00X_your_feature.sql
```