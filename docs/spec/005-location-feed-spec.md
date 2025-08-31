# Location Feed Feature

## Overview
The main feed system that queries posts based on geographic location using various strategies (bounding box, radius, geohash). Supports pagination, filtering, sorting, and manual refresh to check for new posts.

## Database Schema

### Migration: `007_feed_optimizations.sql`

```sql
-- Feed-specific indexes and optimizations
-- (Core tables already exist from posts feature)

-- Composite indexes for feed queries
CREATE INDEX idx_feed_location_time 
  ON post_locations(geohash, post_id)
  INCLUDE (longitude, latitude)
  WHERE EXISTS (
    SELECT 1 FROM posts p 
    WHERE p.id = post_locations.post_id 
    AND p.deleted_at IS NULL
  );

-- Partial indexes for different time windows
CREATE INDEX idx_posts_recent_24h 
  ON posts(created_at DESC) 
  WHERE deleted_at IS NULL 
  AND created_at > NOW() - INTERVAL '24 hours';

CREATE INDEX idx_posts_recent_week 
  ON posts(created_at DESC) 
  WHERE deleted_at IS NULL 
  AND created_at > NOW() - INTERVAL '7 days';

-- Feed session tracking (for "new posts" indicator)
CREATE TABLE feed_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  session_id UUID REFERENCES anonymous_sessions(session_id) ON DELETE CASCADE,
  
  -- Location context
  last_location GEOMETRY(POINT, 4326),
  last_geohash VARCHAR(12),
  last_bounds_north DOUBLE PRECISION,
  last_bounds_south DOUBLE PRECISION,
  last_bounds_east DOUBLE PRECISION,
  last_bounds_west DOUBLE PRECISION,
  
  -- Timing
  last_refresh_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  last_post_seen_at TIMESTAMP WITH TIME ZONE,
  
  -- Stats
  posts_viewed_count INTEGER DEFAULT 0,
  
  CONSTRAINT feed_session_owner CHECK (
    (user_id IS NOT NULL AND session_id IS NULL) OR
    (user_id IS NULL AND session_id IS NOT NULL)
  )
);

CREATE INDEX idx_feed_sessions_user ON feed_sessions(user_id, last_refresh_at DESC);
CREATE INDEX idx_feed_sessions_session ON feed_sessions(session_id, last_refresh_at DESC);

-- Hot zones cache (optional, for optimization)
CREATE TABLE feed_hot_zones (
  geohash_prefix VARCHAR(5) PRIMARY KEY,
  post_count INTEGER NOT NULL DEFAULT 0,
  last_post_at TIMESTAMP WITH TIME ZONE,
  avg_engagement DECIMAL(10,2),
  computed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hot_zones_activity ON feed_hot_zones(post_count DESC, last_post_at DESC);

-- Function to check for new posts efficiently
CREATE OR REPLACE FUNCTION count_new_posts(
  p_geohash_prefix VARCHAR,
  p_since TIMESTAMP WITH TIME ZONE,
  p_max_age_hours INTEGER DEFAULT 24
) RETURNS INTEGER AS $$
BEGIN
  RETURN (
    SELECT COUNT(*)
    FROM posts p
    JOIN post_locations pl ON p.id = pl.post_id
    WHERE pl.geohash LIKE p_geohash_prefix || '%'
      AND p.created_at > p_since
      AND p.created_at > NOW() - (p_max_age_hours || ' hours')::INTERVAL
      AND p.deleted_at IS NULL
  );
END;
$$ LANGUAGE plpgsql STABLE;
```

## SQLC Queries

### File: `sql/queries/feed.sql`

```sql
-- name: GetLocationFeed :many
-- Main feed query with all joined data
SELECT 
  p.*,
  pl.longitude,
  pl.latitude,
  pl.geohash,
  pl.place_name,
  pl.city,
  COALESCE(u.username, 'Anonymous') as author_name,
  (p.author_user_id IS NOT NULL) as is_verified_author,
  COALESCE(reaction_counts.total, 0) as reaction_count,
  COALESCE(comment_counts.total, 0) as comment_count,
  COALESCE(user_reactions.types, '{}') as user_reaction_types
FROM posts p
INNER JOIN post_locations pl ON p.id = pl.post_id
LEFT JOIN users u ON p.author_user_id = u.id
LEFT JOIN (
  SELECT post_id, COUNT(*) as total
  FROM post_reactions
  GROUP BY post_id
) reaction_counts ON p.id = reaction_counts.post_id
LEFT JOIN (
  SELECT post_id, COUNT(*) as total
  FROM post_comments
  WHERE deleted_at IS NULL
  GROUP BY post_id
) comment_counts ON p.id = comment_counts.post_id
LEFT JOIN (
  SELECT post_id, ARRAY_AGG(reaction_type) as types
  FROM post_reactions
  WHERE (user_id = $1 OR session_id = $2)
  GROUP BY post_id
) user_reactions ON p.id = user_reactions.post_id
WHERE 
  ST_Within(
    pl.location,
    ST_MakeEnvelope($3, $4, $5, $6, 4326) -- west, south, east, north
  )
  AND p.deleted_at IS NULL
  AND ($7::interval IS NULL OR p.created_at > NOW() - $7::interval)
  AND ($8::boolean = true OR p.author_user_id != $1) -- include_own_posts
ORDER BY 
  CASE WHEN $9 = 'recent' THEN p.created_at END DESC,
  CASE WHEN $9 = 'popular' THEN COALESCE(reaction_counts.total, 0) END DESC
LIMIT $10
OFFSET $11;

-- name: GetRadiusFeed :many
-- Feed by radius from center point
SELECT 
  p.*,
  pl.longitude,
  pl.latitude,
  pl.geohash,
  pl.place_name,
  COALESCE(u.username, 'Anonymous') as author_name,
  ST_Distance(
    pl.location::geography,
    ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography
  ) as distance_meters,
  COALESCE(reaction_counts.total, 0) as reaction_count,
  COALESCE(comment_counts.total, 0) as comment_count
FROM posts p
INNER JOIN post_locations pl ON p.id = pl.post_id
LEFT JOIN users u ON p.author_user_id = u.id
LEFT JOIN (
  SELECT post_id, COUNT(*) as total
  FROM post_reactions
  GROUP BY post_id
) reaction_counts ON p.id = reaction_counts.post_id
LEFT JOIN (
  SELECT post_id, COUNT(*) as total
  FROM post_comments
  WHERE deleted_at IS NULL
  GROUP BY post_id
) comment_counts ON p.id = comment_counts.post_id
WHERE 
  ST_DWithin(
    pl.location::geography,
    ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography,
    $3 -- radius in meters
  )
  AND p.deleted_at IS NULL
  AND ($4::interval IS NULL OR p.created_at > NOW() - $4::interval)
ORDER BY 
  CASE WHEN $5 = 'distance' THEN distance_meters END ASC,
  CASE WHEN $5 = 'recent' THEN p.created_at END DESC,
  CASE WHEN $5 = 'popular' THEN COALESCE(reaction_counts.total, 0) END DESC
LIMIT $6
OFFSET $7;

-- name: GetGeohashFeed :many
-- Feed by geohash prefix (fast)
SELECT 
  p.*,
  pl.longitude,
  pl.latitude,
  pl.geohash,
  pl.place_name,
  COALESCE(u.username, 'Anonymous') as author_name,
  (p.author_user_id IS NOT NULL) as is_verified_author
FROM posts p
INNER JOIN post_locations pl ON p.id = pl.post_id
LEFT JOIN users u ON p.author_user_id = u.id
WHERE 
  pl.geohash LIKE $1 || '%'
  AND p.deleted_at IS NULL
  AND ($2::timestamp IS NULL OR p.created_at < $2) -- cursor for pagination
  AND ($3::interval IS NULL OR p.created_at > NOW() - $3::interval)
ORDER BY p.created_at DESC
LIMIT $4;

-- name: CountNewPosts :one
-- Check for new posts since last refresh
SELECT COUNT(*) as new_count
FROM posts p
INNER JOIN post_locations pl ON p.id = pl.post_id
WHERE 
  pl.geohash LIKE $1 || '%'
  AND p.created_at > $2 -- since timestamp
  AND p.deleted_at IS NULL
  AND ($3::interval IS NULL OR p.created_at > NOW() - $3::interval);

-- name: GetNewPostIDs :many
-- Get just IDs of new posts (lightweight)
SELECT p.id, p.created_at
FROM posts p
INNER JOIN post_locations pl ON p.id = pl.post_id
WHERE 
  ST_Within(
    pl.location,
    ST_MakeEnvelope($1, $2, $3, $4, 4326)
  )
  AND p.created_at > $5 -- since timestamp
  AND p.deleted_at IS NULL
ORDER BY p.created_at DESC
LIMIT 100; -- Cap at 100 to prevent huge responses

-- name: CreateFeedSession :one
INSERT INTO feed_sessions (
  user_id, session_id, last_location, last_geohash,
  last_bounds_north, last_bounds_south, last_bounds_east, last_bounds_west
) VALUES (
  $1, $2, 
  ST_SetSRID(ST_MakePoint($3, $4), 4326),
  $5, $6, $7, $8, $9
) RETURNING *;

-- name: UpdateFeedSession :exec
UPDATE feed_sessions
SET 
  last_refresh_at = NOW(),
  last_location = ST_SetSRID(ST_MakePoint($3, $4), 4326),
  last_geohash = $5,
  posts_viewed_count = posts_viewed_count + $6
WHERE id = $1
  AND (user_id = $2 OR session_id = $2);

-- name: GetHotZones :many
-- Get areas with high activity
SELECT 
  geohash_prefix,
  post_count,
  last_post_at,
  avg_engagement
FROM feed_hot_zones
WHERE post_count > $1
  AND last_post_at > NOW() - INTERVAL '24 hours'
ORDER BY post_count DESC
LIMIT $2;

-- name: GetTrendingInArea :many
-- Get trending posts in an area
WITH post_engagement AS (
  SELECT 
    p.id,
    p.created_at,
    COUNT(DISTINCT pr.id) as reaction_count,
    COUNT(DISTINCT pc.id) as comment_count,
    -- Trending score: reactions + comments * 2, with time decay
    (COUNT(DISTINCT pr.id) + COUNT(DISTINCT pc.id) * 2) * 
    EXP(-EXTRACT(EPOCH FROM (NOW() - p.created_at)) / 86400) as trending_score
  FROM posts p
  LEFT JOIN post_reactions pr ON p.id = pr.post_id
  LEFT JOIN post_comments pc ON p.id = pc.post_id AND pc.deleted_at IS NULL
  WHERE p.id IN (
    SELECT p2.id 
    FROM posts p2
    JOIN post_locations pl ON p2.id = pl.post_id
    WHERE pl.geohash LIKE $1 || '%'
      AND p2.created_at > NOW() - INTERVAL '48 hours'
      AND p2.deleted_at IS NULL
  )
  GROUP BY p.id, p.created_at
)
SELECT 
  p.*,
  pl.longitude,
  pl.latitude,
  pl.geohash,
  pe.reaction_count,
  pe.comment_count,
  pe.trending_score
FROM posts p
JOIN post_engagement pe ON p.id = pe.id
JOIN post_locations pl ON p.id = pl.post_id
LEFT JOIN users u ON p.author_user_id = u.id
ORDER BY pe.trending_score DESC
LIMIT $2;
```

## Protocol Buffer Definitions

### File: `proto/v1/service/feed.proto`

```protobuf
syntax = "proto3";

package service.v1;

import "entities/v1/post.proto";
import "google/protobuf/timestamp.proto";

service FeedService {
  // Main feed operations
  rpc GetLocationFeed(GetLocationFeedRequest) returns (GetLocationFeedResponse);
  rpc RefreshFeed(RefreshFeedRequest) returns (RefreshFeedResponse);
  
  // Specialized feeds
  rpc GetTrendingFeed(GetTrendingFeedRequest) returns (GetTrendingFeedResponse);
  rpc GetNearbyFeed(GetNearbyFeedRequest) returns (GetNearbyFeedResponse);
  
  // Feed session management
  rpc StartFeedSession(StartFeedSessionRequest) returns (StartFeedSessionResponse);
  rpc UpdateFeedSession(UpdateFeedSessionRequest) returns (UpdateFeedSessionResponse);
}

message GetLocationFeedRequest {
  // Location specification
  oneof location {
    BoundingBox bounds = 1;
    RadiusFilter radius = 2;
    string geohash_prefix = 3;
  }
  
  // Filters
  int32 max_age_hours = 4; // Only posts from last N hours
  bool include_own_posts = 5;
  repeated string exclude_post_ids = 6; // Already seen posts
  
  // Sorting
  enum SortBy {
    RECENT = 0;
    DISTANCE = 1;
    POPULAR = 2;
    TRENDING = 3;
  }
  SortBy sort_by = 7;
  
  // Pagination
  int32 limit = 8; // Default 50, max 100
  int32 offset = 9;
  string cursor = 10; // Alternative cursor-based pagination
}

message BoundingBox {
  double north_lat = 1;
  double south_lat = 2;
  double east_lng = 3;
  double west_lng = 4;
}

message RadiusFilter {
  double center_lat = 1;
  double center_lng = 2;
  int32 radius_meters = 3;
}

message GetLocationFeedResponse {
  repeated entities.v1.Post posts = 1;
  
  // Pagination
  string next_cursor = 2;
  bool has_more = 3;
  
  // Metadata
  int32 total_in_area = 4;
  google.protobuf.Timestamp server_time = 5;
  
  // Session info
  string feed_session_id = 6;
}

message RefreshFeedRequest {
  oneof location {
    BoundingBox bounds = 1;
    RadiusFilter radius = 2;
    string geohash_prefix = 3;
  }
  
  google.protobuf.Timestamp since = 4; // Last refresh time
  string feed_session_id = 5;
}

message RefreshFeedResponse {
  int32 new_post_count = 1;
  repeated string new_post_ids = 2;
  google.protobuf.Timestamp server_time = 3;
  
  // Quick preview of new posts
  repeated PostPreview previews = 4;
}

message PostPreview {
  string id = 1;
  string content_snippet = 2; // First 100 chars
  string author_name = 3;
  google.protobuf.Timestamp created_at = 4;
}

message GetTrendingFeedRequest {
  oneof location {
    BoundingBox bounds = 1;
    string geohash_prefix = 2;
  }
  
  int32 limit = 3;
  int32 time_window_hours = 4; // Look back period
}

message GetTrendingFeedResponse {
  repeated TrendingPost posts = 1;
}

message TrendingPost {
  entities.v1.Post post = 1;
  double trending_score = 2;
  int32 velocity = 3; // Engagements per hour
}

message GetNearbyFeedRequest {
  double latitude = 1;
  double longitude = 2;
  
  // Expanding radius search
  repeated int32 radius_meters = 3; // [500, 1000, 2000, 5000]
  int32 min_posts = 4; // Minimum posts to return
  int32 max_posts = 5; // Maximum posts to return
}

message GetNearbyFeedResponse {
  repeated PostWithDistance posts = 1;
  int32 search_radius_used = 2; // Actual radius that returned results
}

message PostWithDistance {
  entities.v1.Post post = 1;
  float distance_meters = 2;
  string relative_direction = 3; // "N", "NE", "E", etc.
}
```

## Backend Service Implementation

### File: `internal/services/feed/service.go`

```go
package feed

import (
    "context"
    "database/sql"
    "time"
    
    "connectrpc.com/connect"
    "github.com/google/uuid"
    
    "yourproject/internal/repository"
    pbentities "yourproject/proto/gen/v1/entities"
    pbservice "yourproject/proto/gen/v1/service"
)

type Service struct {
    pbservice.UnimplementedFeedServiceServer
    db *repository.Queries
}

func NewService(db *repository.Queries) *Service {
    return &Service{db: db}
}

// GetLocationFeed returns posts for a geographic area
func (s *Service) GetLocationFeed(
    ctx context.Context,
    req *connect.Request[pbservice.GetLocationFeedRequest],
) (*connect.Response[pbservice.GetLocationFeedResponse], error) {
    userID, _ := ctx.Value("user_id").(string)
    sessionID, _ := ctx.Value("session_id").(string)
    
    var posts []repository.Post
    var err error
    
    // Route to appropriate query based on location type
    switch loc := req.Msg.Location.(type) {
    case *pbservice.GetLocationFeedRequest_Bounds:
        posts, err = s.getByBounds(ctx, loc.Bounds, req.Msg, userID, sessionID)
        
    case *pbservice.GetLocationFeedRequest_Radius:
        posts, err = s.getByRadius(ctx, loc.Radius, req.Msg, userID, sessionID)
        
    case *pbservice.GetLocationFeedRequest_GeohashPrefix:
        posts, err = s.getByGeohash(ctx, loc.GeohashPrefix, req.Msg, userID, sessionID)
        
    default:
        return nil, connect.NewError(connect.CodeInvalidArgument, 
            errors.New("location required"))
    }
    
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    
    // Create or update feed session
    sessionID := s.createFeedSession(ctx, req.Msg, userID, sessionID)
    
    // Convert to proto
    protoPosts := make([]*pbentities.Post, len(posts))
    for i, p := range posts {
        protoPosts[i] = s.convertToProtoPost(p)
    }
    
    return connect.NewResponse(&pbservice.GetLocationFeedResponse{
        Posts:          protoPosts,
        HasMore:        len(posts) == req.Msg.Limit,
        ServerTime:     timestamppb.Now(),
        FeedSessionId:  sessionID,
    }), nil
}

// RefreshFeed checks for new posts without fetching all
func (s *Service) RefreshFeed(
    ctx context.Context,
    req *connect.Request[pbservice.RefreshFeedRequest],
) (*connect.Response[pbservice.RefreshFeedResponse], error) {
    // Extract location bounds/geohash
    var geohashPrefix string
    var bounds *pbservice.BoundingBox
    
    switch loc := req.Msg.Location.(type) {
    case *pbservice.RefreshFeedRequest_GeohashPrefix:
        geohashPrefix = loc.GeohashPrefix
    case *pbservice.RefreshFeedRequest_Bounds:
        bounds = loc.Bounds
    case *pbservice.RefreshFeedRequest_Radius:
        // Convert radius to geohash prefix for efficient counting
        geohashPrefix = s.radiusToGeohash(loc.Radius)
    }
    
    // Count new posts
    var newCount int64
    var newPostIDs []string
    
    if bounds != nil {
        // Get new post IDs in bounds
        posts, err := s.db.GetNewPostIDs(ctx, repository.GetNewPostIDsParams{
            West:  bounds.WestLng,
            South: bounds.SouthLat,
            East:  bounds.EastLng,
            North: bounds.NorthLat,
            Since: req.Msg.Since.AsTime(),
        })
        
        if err != nil {
            return nil, connect.NewError(connect.CodeInternal, err)
        }
        
        newCount = int64(len(posts))
        for _, p := range posts {
            newPostIDs = append(newPostIDs, p.ID.String())
        }
    } else {
        // Use geohash-based count
        count, err := s.db.CountNewPosts(ctx, repository.CountNewPostsParams{
            GeohashPrefix: geohashPrefix,
            Since:         req.Msg.Since.AsTime(),
            MaxAge:        sql.NullString{String: "24 hours", Valid: true},
        })
        
        if err != nil {
            return nil, connect.NewError(connect.CodeInternal, err)
        }
        
        newCount = count
    }
    
    // Build previews for first few new posts
    var previews []*pbservice.PostPreview
    if len(newPostIDs) > 0 && len(newPostIDs) <= 5 {
        // Fetch and create previews for up to 5 posts
        for _, id := range newPostIDs[:min(5, len(newPostIDs))] {
            post, _ := s.db.GetPostByID(ctx, uuid.MustParse(id))
            if post.ID != uuid.Nil {
                previews = append(previews, &pbservice.PostPreview{
                    Id:             post.ID.String(),
                    ContentSnippet: truncate(post.Content, 100),
                    AuthorName:     post.AuthorName,
                    CreatedAt:      timestamppb.New(post.CreatedAt),
                })
            }
        }
    }
    
    return connect.NewResponse(&pbservice.RefreshFeedResponse{
        NewPostCount: int32(newCount),
        NewPostIds:   newPostIDs,
        ServerTime:   timestamppb.Now(),
        Previews:     previews,
    }), nil
}

// GetTrendingFeed returns trending posts in an area
func (s *Service) GetTrendingFeed(
    ctx context.Context,
    req *connect.Request[pbservice.GetTrendingFeedRequest],
) (*connect.Response[pbservice.GetTrendingFeedResponse], error) {
    var geohashPrefix string
    
    switch loc := req.Msg.Location.(type) {
    case *pbservice.GetTrendingFeedRequest_GeohashPrefix:
        geohashPrefix = loc.GeohashPrefix
    case *pbservice.GetTrendingFeedRequest_Bounds:
        // Convert bounds to geohash prefix
        geohashPrefix = s.boundsToGeohashPrefix(loc.Bounds)
    }
    
    // Get trending posts
    posts, err := s.db.GetTrendingInArea(ctx, repository.GetTrendingInAreaParams{
        GeohashPrefix: geohashPrefix,
        Limit:         req.Msg.Limit,
    })
    
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    
    // Convert to proto with trending metadata
    trendingPosts := make([]*pbservice.TrendingPost, len(posts))
    for i, p := range posts {
        trendingPosts[i] = &pbservice.TrendingPost{
            Post:          s.convertToProtoPost(p),
            TrendingScore: p.TrendingScore,
            Velocity:      s.calculateVelocity(p),
        }
    }
    
    return connect.NewResponse(&pbservice.GetTrendingFeedResponse{
        Posts: trendingPosts,
    }), nil
}

// Helper: Query by bounding box
func (s *Service) getByBounds(
    ctx context.Context,
    bounds *pbservice.BoundingBox,
    req *pbservice.GetLocationFeedRequest,
    userID, sessionID string,
) ([]repository.Post, error) {
    maxAge := "24h"
    if req.MaxAgeHours > 0 {
        maxAge = fmt.Sprintf("%dh", req.MaxAgeHours)
    }
    
    return s.db.GetLocationFeed(ctx, repository.GetLocationFeedParams{
        UserID:         sql.NullString{String: userID, Valid: userID != ""},
        SessionID:      sql.NullString{String: sessionID, Valid: sessionID != ""},
        West:           bounds.WestLng,
        South:          bounds.SouthLat,
        East:           bounds.EastLng,
        North:          bounds.NorthLat,
        MaxAge:         sql.NullString{String: maxAge, Valid: true},
        IncludeOwnPosts: req.IncludeOwnPosts,
        SortBy:         req.SortBy.String(),
        Limit:          req.Limit,
        Offset:         req.Offset,
    })
}

// Helper: Create or update feed session
func (s *Service) createFeedSession(
    ctx context.Context,
    req *pbservice.GetLocationFeedRequest,
    userID, sessionID string,
) string {
    // Implementation to track feed sessions
    // Helps with "new posts" indicators and analytics
    session, _ := s.db.CreateFeedSession(ctx, repository.CreateFeedSessionParams{
        UserID:    sql.NullString{String: userID, Valid: userID != ""},
        SessionID: sql.NullString{String: sessionID, Valid: sessionID != ""},
        // ... location params
    })
    
    return session.ID.String()
}
```

## Frontend Implementation

### File: `features/feed/hooks/use-location-feed.ts`

```typescript
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useFeedServiceClient } from '@/common/services/connectrpc';
import { useMapStore } from '@/features/map/stores/use-map-store';

interface UseFeedOptions {
  maxAgeHours?: number;
  includeOwnPosts?: boolean;
  sortBy?: 'recent' | 'distance' | 'popular' | 'trending';
  limit?: number;
}

export function useLocationFeed(options: UseFeedOptions = {}) {
  const client = useFeedServiceClient();
  const bounds = useMapStore(state => state.bounds);
  const queryClient = useQueryClient();
  
  return useQuery({
    queryKey: ['feed', 'location', bounds, options],
    queryFn: async () => {
      if (!bounds) return { posts: [], hasMore: false };
      
      const response = await client.getLocationFeed({
        location: {
          case: 'bounds',
          value: {
            north_lat: bounds.getNorth(),
            south_lat: bounds.getSouth(),
            east_lng: bounds.getEast(),
            west_lng: bounds.getWest(),
          }
        },
        max_age_hours: options.maxAgeHours ?? 24,
        include_own_posts: options.includeOwnPosts ?? true,
        sort_by: options.sortBy ?? 'recent',
        limit: options.limit ?? 50,
      });
      
      // Store session ID for refresh
      if (response.feed_session_id) {
        queryClient.setQueryData(['feed-session'], response.feed_session_id);
      }
      
      return {
        posts: response.posts,
        hasMore: response.has_more,
        serverTime: response.server_time,
      };
    },
    enabled: !!bounds,
    staleTime: 60 * 1000, // 1 minute
    refetchOnWindowFocus: false,
  });
}
```

### File: `features/feed/hooks/use-feed-refresh.ts`

```typescript
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useFeedServiceClient } from '@/common/services/connectrpc';
import { useMapStore } from '@/features/map/stores/use-map-store';
import { useRef, useCallback } from 'react';

export function useFeedRefresh() {
  const client = useFeedServiceClient();
  const queryClient = useQueryClient();
  const bounds = useMapStore(state => state.bounds);
  const lastRefresh = useRef(new Date());
  
  // Check for new posts periodically
  const { data: refreshData } = useQuery({
    queryKey: ['feed', 'refresh-check', bounds],
    queryFn: async () => {
      if (!bounds) return null;
      
      const sessionId = queryClient.getQueryData<string>(['feed-session']);
      
      const response = await client.refreshFeed({
        location: {
          case: 'bounds',
          value: {
            north_lat: bounds.getNorth(),
            south_lat: bounds.getSouth(),
            east_lng: bounds.getEast(),
            west_lng: bounds.getWest(),
          }
        },
        since: lastRefresh.current,
        feed_session_id: sessionId,
      });
      
      return {
        newPostCount: response.new_post_count,
        newPostIds: response.new_post_ids,
        previews: response.previews,
        serverTime: response.server_time,
      };
    },
    enabled: !!bounds,
    refetchInterval: 60 * 1000, // Check every minute
    refetchIntervalInBackground: false,
  });
  
  const refresh = useCallback(async () => {
    lastRefresh.current = new Date();
    
    // Invalidate and refetch feed
    await queryClient.invalidateQueries({ queryKey: ['feed', 'location'] });
    
    // Clear new post indicator
    queryClient.setQueryData(['feed', 'refresh-check'], null);
  }, [queryClient]);
  
  return {
    newPostCount: refreshData?.newPostCount ?? 0,
    newPostIds: refreshData?.newPostIds ?? [],
    previews: refreshData?.previews ?? [],
    refresh,
    lastRefresh: lastRefresh.current,
  };
}
```

### File: `features/feed/components/feed-container.tsx`

```tsx
import { useState, useEffect } from 'react';
import { useLocationFeed } from '../hooks/use-location-feed';
import { useFeedRefresh } from '../hooks/use-feed-refresh';
import { useInfiniteScroll } from '../hooks/use-infinite-scroll';
import { PostCard } from '@/features/posts/components/post-card';
import { FeedFilters } from './feed-filters';
import { NewPostsIndicator } from './new-posts-indicator';
import { FeedEmptyState } from './feed-empty-state';
import { FeedSkeleton } from './feed-skeleton';

export function FeedContainer() {
  const [filters, setFilters] = useState({
    sortBy: 'recent' as const,
    maxAgeHours: 24,
    includeOwnPosts: true,
  });
  
  const { data, isLoading, error, refetch } = useLocationFeed(filters);
  const { newPostCount, previews, refresh } = useFeedRefresh();
  const { loadMoreRef, isLoadingMore } = useInfiniteScroll({
    hasMore: data?.hasMore ?? false,
    onLoadMore: () => {
      // Load more implementation
    },
  });
  
  if (isLoading) return <FeedSkeleton />;
  
  if (error) {
    return (
      <FeedError 
        error={error}
        onRetry={() => refetch()}
      />
    );
  }
  
  if (!data?.posts.length) {
    return <FeedEmptyState onRefresh={refresh} />;
  }
  
  return (
    <div className="relative space-y-4">
      {/* Filters */}
      <FeedFilters
        filters={filters}
        onChange={setFilters}
      />
      
      {/* New posts indicator */}
      {newPostCount > 0 && (
        <NewPostsIndicator
          count={newPostCount}
          previews={previews}
          onRefresh={refresh}
        />
      )}
      
      {/* Posts list */}
      <div className="space-y-4">
        {data.posts.map(post => (
          <PostCard key={post.id} post={post} />
        ))}
      </div>
      
      {/* Load more trigger */}
      {data.hasMore && (
        <div ref={loadMoreRef} className="py-4 text-center">
          {isLoadingMore ? (
            <span className="text-gray-500">Loading more...</span>
          ) : (
            <button
              onClick={() => loadMoreRef.current?.scrollIntoView()}
              className="text-blue-500 hover:underline"
            >
              Load more posts
            </button>
          )}
        </div>
      )}
      
      {/* Manual refresh button */}
      <RefreshButton
        onClick={refresh}
        className="fixed bottom-20 right-4"
      />
    </div>
  );
}
```

### File: `features/feed/components/new-posts-indicator.tsx`

```tsx
import { motion, AnimatePresence } from 'framer-motion';
import { PostPreview } from '@/proto/gen/v1/service/feed_pb';

interface NewPostsIndicatorProps {
  count: number;
  previews: PostPreview[];
  onRefresh: () => void;
}

export function NewPostsIndicator({
  count,
  previews,
  onRefresh,
}: NewPostsIndicatorProps) {
  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -20 }}
        className="sticky top-0 z-10 bg-blue-500 text-white rounded-lg shadow-lg"
      >
        <button
          onClick={onRefresh}
          className="w-full p-3 text-left"
        >
          <div className="flex items-center justify-between">
            <span className="font-medium">
              {count} new {count === 1 ? 'post' : 'posts'}
            </span>
            <span className="text-sm opacity-90">
              Tap to refresh
            </span>
          </div>
          
          {/* Preview snippets */}
          {previews.length > 0 && (
            <div className="mt-2 space-y-1">
              {previews.slice(0, 2).map(preview => (
                <div key={preview.id} className="text-sm opacity-90">
                  <span className="font-medium">{preview.author_name}:</span>
                  <span className="ml-1">{preview.content_snippet}</span>
                </div>
              ))}
              {count > 2 && (
                <div className="text-sm opacity-75">
                  and {count - 2} more...
                </div>
              )}
            </div>
          )}
        </button>
      </motion.div>
    </AnimatePresence>
  );
}
```

### File: `features/feed/components/feed-filters.tsx`

```tsx
import { useState } from 'react';
import { ChevronDown } from 'lucide-react';

interface FeedFiltersProps {
  filters: {
    sortBy: 'recent' | 'distance' | 'popular' | 'trending';
    maxAgeHours: number;
    includeOwnPosts: boolean;
  };
  onChange: (filters: any) => void;
}

export function FeedFilters({ filters, onChange }: FeedFiltersProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  
  return (
    <div className="bg-white border rounded-lg p-3">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="flex items-center justify-between w-full"
      >
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-600">Sort by:</span>
          <span className="font-medium">{filters.sortBy}</span>
          <span className="text-sm text-gray-600">•</span>
          <span className="text-sm text-gray-600">
            Last {filters.maxAgeHours}h
          </span>
        </div>
        <ChevronDown
          className={`transition-transform ${isExpanded ? 'rotate-180' : ''}`}
          size={20}
        />
      </button>
      
      {isExpanded && (
        <div className="mt-3 pt-3 border-t space-y-3">
          {/* Sort options */}
          <div>
            <label className="text-sm text-gray-600">Sort by</label>
            <div className="grid grid-cols-2 gap-2 mt-1">
              {['recent', 'distance', 'popular', 'trending'].map(sort => (
                <button
                  key={sort}
                  onClick={() => onChange({ ...filters, sortBy: sort })}
                  className={`px-3 py-1 rounded text-sm ${
                    filters.sortBy === sort
                      ? 'bg-blue-500 text-white'
                      : 'bg-gray-100 text-gray-700'
                  }`}
                >
                  {sort.charAt(0).toUpperCase() + sort.slice(1)}
                </button>
              ))}
            </div>
          </div>
          
          {/* Time filter */}
          <div>
            <label className="text-sm text-gray-600">Time range</label>
            <div className="grid grid-cols-3 gap-2 mt-1">
              {[1, 24, 168].map(hours => (
                <button
                  key={hours}
                  onClick={() => onChange({ ...filters, maxAgeHours: hours })}
                  className={`px-3 py-1 rounded text-sm ${
                    filters.maxAgeHours === hours
                      ? 'bg-blue-500 text-white'
                      : 'bg-gray-100 text-gray-700'
                  }`}
                >
                  {hours === 1 ? '1 hour' : hours === 24 ? '24 hours' : '1 week'}
                </button>
              ))}
            </div>
          </div>
          
          {/* Toggle own posts */}
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={filters.includeOwnPosts}
              onChange={(e) => onChange({ 
                ...filters, 
                includeOwnPosts: e.target.checked 
              })}
              className="rounded"
            />
            <span className="text-sm">Show my posts</span>
          </label>
        </div>
      )}
    </div>
  );
}
```

### File: `features/feed/hooks/use-nearby-feed.ts`

```typescript
import { useQuery } from '@tanstack/react-query';
import { useFeedServiceClient } from '@/common/services/connectrpc';
import { useGeolocation } from '@/features/map/hooks/use-geolocation';

export function useNearbyFeed(minPosts = 10, maxPosts = 50) {
  const client = useFeedServiceClient();
  const { location, isLoading: isLoadingLocation } = useGeolocation();
  
  return useQuery({
    queryKey: ['feed', 'nearby', location, minPosts, maxPosts],
    queryFn: async () => {
      if (!location) throw new Error('Location required');
      
      const response = await client.getNearbyFeed({
        latitude: location.latitude,
        longitude: location.longitude,
        radius_meters: [500, 1000, 2000, 5000, 10000], // Expanding search
        min_posts: minPosts,
        max_posts: maxPosts,
      });
      
      return {
        posts: response.posts,
        searchRadius: response.search_radius_used,
      };
    },
    enabled: !!location && !isLoadingLocation,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}
```

### File: `features/feed/hooks/use-trending-feed.ts`

```typescript
import { useQuery } from '@tanstack/react-query';
import { useFeedServiceClient } from '@/common/services/connectrpc';
import { useMapStore } from '@/features/map/stores/use-map-store';

export function useTrendingFeed(timeWindowHours = 48) {
  const client = useFeedServiceClient();
  const bounds = useMapStore(state => state.bounds);
  
  return useQuery({
    queryKey: ['feed', 'trending', bounds, timeWindowHours],
    queryFn: async () => {
      if (!bounds) return { posts: [] };
      
      const response = await client.getTrendingFeed({
        location: {
          case: 'bounds',
          value: {
            north_lat: bounds.getNorth(),
            south_lat: bounds.getSouth(),
            east_lng: bounds.getEast(),
            west_lng: bounds.getWest(),
          }
        },
        limit: 20,
        time_window_hours: timeWindowHours,
      });
      
      return {
        posts: response.posts,
      };
    },
    enabled: !!bounds,
    staleTime: 10 * 60 * 1000, // 10 minutes
  });
}
```

### File: `features/feed/components/feed-empty-state.tsx`

```tsx
import { MapPin, RefreshCw } from 'lucide-react';

interface FeedEmptyStateProps {
  onRefresh: () => void;
}

export function FeedEmptyState({ onRefresh }: FeedEmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-12 px-4">
      <div className="w-20 h-20 bg-gray-100 rounded-full flex items-center justify-center mb-4">
        <MapPin size={32} className="text-gray-400" />
      </div>
      
      <h3 className="text-lg font-semibold text-gray-900 mb-2">
        No posts in this area
      </h3>
      
      <p className="text-gray-600 text-center max-w-sm mb-6">
        Be the first to share what's happening here, or adjust your location to see posts from nearby.
      </p>
      
      <div className="flex gap-3">
        <button
          onClick={onRefresh}
          className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
        >
          <RefreshCw size={16} />
          Refresh
        </button>
        
        <button
          onClick={() => {
            // Open post creation
            document.dispatchEvent(new CustomEvent('open-post-form'));
          }}
          className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
        >
          Create First Post
        </button>
      </div>
    </div>
  );
}
```

### File: `features/feed/store/feed-store.ts`

```typescript
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface FeedState {
  // View preferences
  viewMode: 'list' | 'grid' | 'map';
  filters: {
    sortBy: 'recent' | 'distance' | 'popular' | 'trending';
    maxAgeHours: number;
    includeOwnPosts: boolean;
  };
  
  // Cached data
  lastRefresh: Date | null;
  seenPostIds: Set<string>;
  
  // Actions
  setViewMode: (mode: 'list' | 'grid' | 'map') => void;
  setFilters: (filters: any) => void;
  markPostsSeen: (ids: string[]) => void;
  clearSeenPosts: () => void;
}

export const useFeedStore = create<FeedState>()(
  persist(
    (set) => ({
      viewMode: 'list',
      filters: {
        sortBy: 'recent',
        maxAgeHours: 24,
        includeOwnPosts: true,
      },
      lastRefresh: null,
      seenPostIds: new Set(),
      
      setViewMode: (mode) => set({ viewMode: mode }),
      
      setFilters: (filters) => set((state) => ({
        filters: { ...state.filters, ...filters }
      })),
      
      markPostsSeen: (ids) => set((state) => ({
        seenPostIds: new Set([...state.seenPostIds, ...ids])
      })),
      
      clearSeenPosts: () => set({ seenPostIds: new Set() }),
    }),
    {
      name: 'feed-preferences',
      partialize: (state) => ({
        viewMode: state.viewMode,
        filters: state.filters,
      }),
    }
  )
);
```

## Key Implementation Notes

### Query Strategies
- **Bounding Box**: Best for viewport-based queries (map view)
- **Radius**: Best for "near me" features
- **Geohash**: Best for performance at scale
- Start with geohash prefix for initial query, refine with PostGIS

### Pagination Approaches
- **Offset-based**: Simple but can miss posts during pagination
- **Cursor-based**: Use `created_at` timestamp as cursor
- **Hybrid**: Use offset for initial load, cursor for "load more"

### Performance Optimizations
- Create composite indexes for common query patterns
- Use partial indexes for time-based queries
- Consider materialized views for hot zones
- Implement query result caching on backend

### Feed Freshness
- Check for new posts every 60 seconds
- Show preview of new posts without auto-refresh
- Track last refresh time per session
- Use lightweight count queries for refresh check

### Sorting Strategies
- **Recent**: Simple `ORDER BY created_at DESC`
- **Distance**: Requires PostGIS `ST_Distance` calculation
- **Popular**: Sort by engagement (reactions + comments)
- **Trending**: Time-decay algorithm with engagement velocity

### Location Privacy
- Don't expose exact coordinates in feed
- Round coordinates to reduce precision
- Allow "fuzzy" location mode
- Respect user's location sharing preferences

### Scale Considerations
- Partition posts table by geohash prefix at scale
- Use read replicas for feed queries
- Implement Redis caching for hot posts
- Consider ElasticSearch for complex queries

### Error Handling
- Graceful degradation when location unavailable
- Fallback to IP-based location
- Handle stale location data
- Retry logic for failed refreshes

### Future Enhancements
- Real-time updates via WebSocket/SSE
- Personalized feed algorithm
- Save feed positions across sessions
- Offline support with local caching
- Feed analytics and insights
