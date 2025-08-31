# Posts Core Feature

## Overview
Core posting functionality with location tagging and media attachments. This is the foundation for all content in the system, supporting both authenticated users and anonymous sessions.

## Database Schema

### Migration: `004_posts_and_locations.sql`

```sql
-- Core posts table (thin parent for all post types)
CREATE TABLE posts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  type VARCHAR(20) NOT NULL DEFAULT 'pin',
  author_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  author_session_id UUID REFERENCES anonymous_sessions(session_id) ON DELETE SET NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMP WITH TIME ZONE,
  
  -- Ensure exactly one author type
  CONSTRAINT posts_author_check CHECK (
    (author_user_id IS NOT NULL AND author_session_id IS NULL) OR
    (author_user_id IS NULL AND author_session_id IS NOT NULL)
  ),
  
  CONSTRAINT posts_type_check CHECK (type IN ('pin', 'story', 'video'))
);

-- Spatial data with PostGIS
CREATE TABLE post_locations (
  post_id UUID PRIMARY KEY REFERENCES posts(id) ON DELETE CASCADE,
  location GEOMETRY(POINT, 4326) NOT NULL,
  geohash VARCHAR(12) NOT NULL, -- Precomputed in application
  
  -- Generated for client convenience
  longitude DOUBLE PRECISION GENERATED ALWAYS AS (ST_X(location)) STORED,
  latitude DOUBLE PRECISION GENERATED ALWAYS AS (ST_Y(location)) STORED,
  
  -- Optional metadata
  accuracy_meters REAL,
  altitude_meters REAL,
  place_name TEXT,
  city VARCHAR(100),
  country_code CHAR(2)
);

-- Media attachments
CREATE TABLE post_media (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  media_type VARCHAR(20) NOT NULL,
  media_url TEXT NOT NULL,
  thumbnail_url TEXT,
  width INTEGER,
  height INTEGER,
  file_size_bytes BIGINT,
  duration_seconds INTEGER, -- for videos
  display_order INTEGER DEFAULT 0,
  
  CONSTRAINT post_media_type_check CHECK (media_type IN ('image', 'video', 'audio'))
);

-- Indexes for performance
CREATE INDEX idx_posts_created ON posts(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_posts_author_user ON posts(author_user_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_posts_type ON posts(type, created_at DESC) WHERE deleted_at IS NULL;

CREATE INDEX idx_post_locations_gist ON post_locations USING GIST(location);
CREATE INDEX idx_post_locations_geohash ON post_locations(geohash text_pattern_ops);
-- Optimize for different zoom levels
CREATE INDEX idx_post_locations_geohash_3 ON post_locations(LEFT(geohash, 3));
CREATE INDEX idx_post_locations_geohash_5 ON post_locations(LEFT(geohash, 5));
CREATE INDEX idx_post_locations_geohash_7 ON post_locations(LEFT(geohash, 7));

CREATE INDEX idx_post_media_post ON post_media(post_id, display_order);
```

## SQLC Queries

### File: `sql/queries/posts.sql`

```sql
-- name: CreatePost :one
INSERT INTO posts (
  type, author_user_id, author_session_id, content
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: CreatePostLocation :one
INSERT INTO post_locations (
  post_id, location, geohash, accuracy_meters, place_name, city, country_code
) VALUES (
  $1,
  ST_SetSRID(ST_MakePoint($2, $3), 4326), -- longitude, latitude
  $4, $5, $6, $7, $8
) RETURNING *;

-- name: CreatePostMedia :one
INSERT INTO post_media (
  post_id, media_type, media_url, thumbnail_url, 
  width, height, file_size_bytes, duration_seconds, display_order
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetPostByID :one
SELECT 
  p.*,
  pl.longitude,
  pl.latitude,
  pl.geohash,
  pl.place_name,
  pl.city,
  pl.country_code,
  COALESCE(u.username, 'Anonymous') as author_name,
  (p.author_user_id IS NOT NULL) as is_verified_author
FROM posts p
LEFT JOIN post_locations pl ON p.id = pl.post_id
LEFT JOIN users u ON p.author_user_id = u.id
WHERE p.id = $1 AND p.deleted_at IS NULL;

-- name: GetPostMedia :many
SELECT * FROM post_media
WHERE post_id = $1
ORDER BY display_order ASC;

-- name: GetMyPosts :many
SELECT 
  p.*,
  pl.longitude,
  pl.latitude,
  pl.geohash,
  pl.place_name,
  COALESCE(u.username, 'Anonymous') as author_name
FROM posts p
LEFT JOIN post_locations pl ON p.id = pl.post_id
LEFT JOIN users u ON p.author_user_id = u.id
WHERE 
  (p.author_user_id = $1 OR p.author_session_id = $2)
  AND p.deleted_at IS NULL
ORDER BY p.created_at DESC
LIMIT $3
OFFSET $4;

-- name: UpdatePost :one
UPDATE posts
SET content = $2, updated_at = NOW()
WHERE id = $1 
  AND (author_user_id = $3 OR author_session_id = $4)
  AND deleted_at IS NULL
RETURNING *;

-- name: DeletePost :exec
UPDATE posts
SET deleted_at = NOW()
WHERE id = $1
  AND (author_user_id = $2 OR author_session_id = $3);

-- name: GetPostsByBoundingBox :many
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
  ST_Within(
    pl.location,
    ST_MakeEnvelope($1, $2, $3, $4, 4326) -- west, south, east, north
  )
  AND p.deleted_at IS NULL
  AND ($5::interval IS NULL OR p.created_at > NOW() - $5::interval)
ORDER BY p.created_at DESC
LIMIT $6;

-- name: GetPostsByRadius :many
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
  ) as distance_meters
FROM posts p
INNER JOIN post_locations pl ON p.id = pl.post_id
LEFT JOIN users u ON p.author_user_id = u.id
WHERE 
  ST_DWithin(
    pl.location::geography,
    ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography,
    $3 -- radius in meters
  )
  AND p.deleted_at IS NULL
  AND ($4::interval IS NULL OR p.created_at > NOW() - $4::interval)
ORDER BY distance_meters ASC
LIMIT $5;

-- name: GetPostsByGeohashPrefix :many
SELECT 
  p.*,
  pl.longitude,
  pl.latitude,
  pl.geohash,
  pl.place_name,
  COALESCE(u.username, 'Anonymous') as author_name
FROM posts p
INNER JOIN post_locations pl ON p.id = pl.post_id
LEFT JOIN users u ON p.author_user_id = u.id
WHERE 
  pl.geohash LIKE $1 || '%'
  AND p.deleted_at IS NULL
  AND ($2::timestamp IS NULL OR p.created_at < $2) -- cursor
ORDER BY p.created_at DESC
LIMIT $3;
```

## Protocol Buffer Definitions

### File: `proto/v1/entities/post.proto`

```protobuf
syntax = "proto3";

package entities.v1;

import "google/protobuf/timestamp.proto";

message Post {
  string id = 1;
  string type = 2;
  string content = 3;
  
  // Author (one of these will be set)
  string author_user_id = 4;
  string author_session_id = 5;
  string author_username = 6;
  bool is_verified_author = 7;
  
  google.protobuf.Timestamp created_at = 8;
  google.protobuf.Timestamp updated_at = 9;
  
  // Optional location
  Location location = 10;
  
  // Media attachments
  repeated Media media = 11;
  
  // Engagement counts (populated by join queries)
  int32 reaction_count = 12;
  int32 comment_count = 13;
  
  // Current user's engagement
  repeated string user_reactions = 14;
}

message Location {
  double latitude = 1;
  double longitude = 2;
  string geohash = 3;
  float accuracy_meters = 4;
  string place_name = 5;
  string city = 6;
  string country_code = 7;
}

message Media {
  string id = 1;
  string type = 2; // "image", "video", "audio"
  string url = 3;
  string thumbnail_url = 4;
  int32 width = 5;
  int32 height = 6;
  int32 duration_seconds = 7;
  int32 display_order = 8;
}
```

### File: `proto/v1/service/posts.proto`

```protobuf
syntax = "proto3";

package service.v1;

import "entities/v1/post.proto";

service PostService {
  // Core CRUD operations
  rpc CreatePost(CreatePostRequest) returns (CreatePostResponse);
  rpc GetPost(GetPostRequest) returns (GetPostResponse);
  rpc UpdatePost(UpdatePostRequest) returns (UpdatePostResponse);
  rpc DeletePost(DeletePostRequest) returns (DeletePostResponse);
  
  // List operations
  rpc GetMyPosts(GetMyPostsRequest) returns (GetMyPostsResponse);
  rpc GetUserPosts(GetUserPostsRequest) returns (GetUserPostsResponse);
  
  // Location queries
  rpc GetPostsByLocation(GetPostsByLocationRequest) returns (GetPostsByLocationResponse);
  rpc GetPostsByRadius(GetPostsByRadiusRequest) returns (GetPostsByRadiusResponse);
}

message CreatePostRequest {
  string content = 1;
  entities.v1.Location location = 2;
  repeated MediaInput media = 3;
  string type = 4; // defaults to "pin"
}

message MediaInput {
  string type = 1;
  string url = 2;
  string thumbnail_url = 3;
  int32 width = 4;
  int32 height = 5;
  int32 duration_seconds = 6;
}

message CreatePostResponse {
  entities.v1.Post post = 1;
}

// ... additional message definitions
```

## Backend Service Implementation

### File: `internal/services/posts/service.go`

```go
package posts

import (
    "context"
    "database/sql"
    
    "connectrpc.com/connect"
    "github.com/mmcloughlin/geohash"
    
    "yourproject/internal/repository"
    pbentities "yourproject/proto/gen/v1/entities"
    pbservice "yourproject/proto/gen/v1/service"
)

type Service struct {
    pbservice.UnimplementedPostServiceServer
    db *repository.Queries
}

func NewService(db *repository.Queries) *Service {
    return &Service{db: db}
}

// CreatePost creates a new post with optional location and media
func (s *Service) CreatePost(
    ctx context.Context,
    req *connect.Request[pbservice.CreatePostRequest],
) (*connect.Response[pbservice.CreatePostResponse], error) {
    // Extract auth from context (set by middleware)
    userID, _ := ctx.Value("user_id").(string)
    sessionID, _ := ctx.Value("session_id").(string)
    
    if userID == "" && sessionID == "" {
        return nil, connect.NewError(connect.CodeUnauthenticated, nil)
    }
    
    // Begin transaction for atomic creation
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    defer tx.Rollback()
    
    // 1. Create post
    post, err := s.db.WithTx(tx).CreatePost(ctx, repository.CreatePostParams{
        Type:            req.Msg.Type,
        AuthorUserID:    sql.NullString{String: userID, Valid: userID != ""},
        AuthorSessionID: sql.NullString{String: sessionID, Valid: sessionID != ""},
        Content:         req.Msg.Content,
    })
    
    // 2. Add location if provided
    if req.Msg.Location != nil {
        gh := geohash.EncodeWithPrecision(
            req.Msg.Location.Latitude,
            req.Msg.Location.Longitude,
            12, // max precision ~3.7cm
        )
        
        _, err = s.db.WithTx(tx).CreatePostLocation(ctx, repository.CreatePostLocationParams{
            PostID:    post.ID,
            Longitude: req.Msg.Location.Longitude,
            Latitude:  req.Msg.Location.Latitude,
            Geohash:   gh,
            // ... other fields
        })
    }
    
    // 3. Add media if provided
    for i, media := range req.Msg.Media {
        _, err = s.db.WithTx(tx).CreatePostMedia(ctx, repository.CreatePostMediaParams{
            PostID:       post.ID,
            MediaType:    media.Type,
            MediaURL:     media.Url,
            ThumbnailURL: sql.NullString{String: media.ThumbnailUrl, Valid: media.ThumbnailUrl != ""},
            DisplayOrder: int32(i),
            // ... other fields
        })
    }
    
    if err := tx.Commit(); err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    
    // Convert to proto and return
    protoPost := s.convertToProtoPost(post)
    
    return connect.NewResponse(&pbservice.CreatePostResponse{
        Post: protoPost,
    }), nil
}

// Helper to convert DB model to proto
func (s *Service) convertToProtoPost(dbPost repository.Post) *pbentities.Post {
    // Implementation to map database model to proto message
    return &pbentities.Post{
        Id:              dbPost.ID.String(),
        Type:            dbPost.Type,
        Content:         dbPost.Content,
        AuthorUserId:    dbPost.AuthorUserID.String,
        AuthorSessionId: dbPost.AuthorSessionID.String,
        // ... map other fields
    }
}
```

## Frontend Implementation

### File: `features/posts/hooks/use-create-post.ts`

```typescript
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { usePostServiceClient } from '@/common/services/connectrpc';
import { encodeGeohash } from '@/common/lib/geohash';

interface CreatePostInput {
  content: string;
  location?: {
    lat: number;
    lng: number;
    accuracy?: number;
    placeName?: string;
  };
  media?: File[];
}

export function useCreatePost() {
  const client = usePostServiceClient();
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (input: CreatePostInput) => {
      // Upload media first if present
      const mediaInputs = input.media ? 
        await uploadMedia(input.media) : [];
      
      // Compute geohash client-side
      const location = input.location ? {
        latitude: input.location.lat,
        longitude: input.location.lng,
        geohash: encodeGeohash(input.location.lat, input.location.lng, 12),
        accuracy_meters: input.location.accuracy,
        place_name: input.location.placeName,
      } : undefined;
      
      const response = await client.createPost({
        content: input.content,
        location,
        media: mediaInputs,
        type: 'pin',
      });
      
      return response.post;
    },
    onSuccess: () => {
      // Invalidate relevant queries
      queryClient.invalidateQueries({ queryKey: ['posts'] });
      queryClient.invalidateQueries({ queryKey: ['my-posts'] });
      queryClient.invalidateQueries({ queryKey: ['feed'] });
    },
  });
}

// Helper function for media upload
async function uploadMedia(files: File[]): Promise<MediaInput[]> {
  // Implementation for uploading to S3/CDN
  // Returns array of media URLs and metadata
}
```

### File: `features/posts/components/post-form.tsx`

```tsx
import { useState } from 'react';
import { useCreatePost } from '../hooks/use-create-post';
import { MediaUploader } from './media-uploader';
import { LocationPicker } from './location-picker';

export function PostForm({ 
  onSuccess,
  initialLocation 
}: {
  onSuccess?: () => void;
  initialLocation?: { lat: number; lng: number };
}) {
  const [content, setContent] = useState('');
  const [media, setMedia] = useState<File[]>([]);
  const [location, setLocation] = useState(initialLocation);
  const { mutate: createPost, isPending } = useCreatePost();
  
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    createPost(
      { content, location, media },
      {
        onSuccess: () => {
          setContent('');
          setMedia([]);
          onSuccess?.();
        },
      }
    );
  };
  
  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder="What's happening here?"
        className="w-full p-3 border rounded-lg"
        maxLength={280}
      />
      
      <div className="flex gap-2">
        <LocationPicker
          value={location}
          onChange={setLocation}
        />
        
        <MediaUploader
          value={media}
          onChange={setMedia}
          maxFiles={4}
        />
      </div>
      
      <button
        type="submit"
        disabled={!content.trim() || isPending}
        className="px-4 py-2 bg-blue-500 text-white rounded-lg"
      >
        {isPending ? 'Posting...' : 'Post'}
      </button>
    </form>
  );
}
```

### File: `features/posts/components/post-card.tsx`

```tsx
import { Post } from '@/proto/gen/v1/entities/post_pb';
import { MediaGallery } from './media-gallery';
import { LocationBadge } from './location-badge';
import { formatTimeAgo } from '@/common/lib/time';

export function PostCard({ post }: { post: Post }) {
  return (
    <article className="p-4 border rounded-lg space-y-3">
      <header className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="w-10 h-10 rounded-full bg-gray-200" />
          <div>
            <p className="font-medium">
              {post.author_username}
              {post.is_verified_author && (
                <span className="ml-1 text-blue-500">✓</span>
              )}
            </p>
            <p className="text-sm text-gray-500">
              {formatTimeAgo(post.created_at)}
            </p>
          </div>
        </div>
      </header>
      
      <p className="text-gray-900">{post.content}</p>
      
      {post.media.length > 0 && (
        <MediaGallery media={post.media} />
      )}
      
      {post.location && (
        <LocationBadge location={post.location} />
      )}
      
      <footer className="flex items-center gap-4 pt-2 border-t">
        {/* Reaction and comment buttons will be added by other features */}
        <div className="text-sm text-gray-500">
          {post.reaction_count} reactions · {post.comment_count} comments
        </div>
      </footer>
    </article>
  );
}
```

## Key Implementation Notes

### Geohash Strategy
- Compute geohash in application code before storing
- Use 12-character precision (approximately 3.7cm)
- Create indexes for different prefix lengths (3, 5, 7 chars) for zoom levels

### Media Handling
- Upload media to S3/CDN before creating post
- Store URLs in database, not binary data
- Generate thumbnails server-side for videos
- Limit to 4 media items per post for MVP

### Location Privacy
- Allow users to reduce location precision
- Option to show city/neighborhood instead of exact location
- Never expose raw GPS coordinates in public APIs

### Performance Optimizations
- Use PostGIS spatial indexes for location queries
- Batch media uploads in parallel
- Lazy load media in feed views
- Cache user's own posts client-side

### Error Handling
- Validate content length (280 chars max)
- Check location permissions before allowing location posts
- Handle media upload failures gracefully
- Rollback transaction if any step fails

### Security Considerations
- Sanitize HTML content to prevent XSS
- Validate media types and sizes
- Rate limit post creation (e.g., 10 posts per hour)
- Check user/session ownership for updates/deletes
