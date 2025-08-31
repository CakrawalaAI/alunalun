# Social Media Database Schema Research: Multi-Modal Post Architecture Analysis

## Executive Summary

This research analyzes how major social media platforms (TikTok, Instagram, Facebook, Snapchat) design their database schemas to handle multiple post modalities including text posts, images, videos, live streams, stories, and ephemeral content. We compare these approaches with alunalun's current design to identify potential improvements and architectural patterns.

## Table of Contents
1. [Platform-Specific Architectures](#platform-specific-architectures)
2. [Common Patterns Across Platforms](#common-patterns-across-platforms)
3. [Schema Design Approaches](#schema-design-approaches)
4. [Comparison with Alunalun Design](#comparison-with-alunalun-design)
5. [Recommendations](#recommendations)

## Platform-Specific Architectures

### Facebook - TAO (The Associations and Objects)

Facebook uses a graph-based data model called TAO that handles over **1 billion read requests and millions of write requests per second**.

**Key Architecture Features:**
- **Objects and Associations Model**: Objects are typed nodes (users, posts, comments), associations are typed directed edges
- **Schema Structure**:
  - Objects table: `id (int)`, `otype (string)`, `data (byte)`
  - Associations table: `id1`, `id2` (edge endpoints), `atype` (edge type), `data` (key-value pairs)
  - Maximum one association of a given type between any two objects
- **Post Type Handling**: Different object types with size limits:
  - Users: max 50 characters
  - Posts: max 63,206 characters  
  - Comments: max 8,000 characters
- **Sharding Strategy**: Each object_id contains a shard_id for logical location
- **Consistency Model**: Eventual consistency by default with two-tier caching (followers/leaders)

### Instagram - Multi-Database Architecture

Instagram evolved from Django monolith to microservices, handling **140 billion reels played daily** and **50 million+ video uploads per day**.

**Database Stack:**
- **PostgreSQL**: User data, profiles, followers, interactions (with horizontal sharding)
- **Apache Cassandra**: Large-scale datasets, news feeds, notifications (multiple regional clusters)
- **Memcached/Redis**: Caching layer for frequently accessed data
- **Amazon S3**: Media storage with pre-signed URLs for direct uploads

**Schema Design (Example):**
```sql
-- Core tables
users (user_id, username, email, password_hash, bio, profile_picture)
posts (post_id, user_id, datetime_added, image_url, caption)
comments (comment_id, user_id, post_id, datetime_added, text)
likes (like_id, user_id, post_id, datetime_added)
stories (story_id, user_id, datetime_added, image_url, caption)

-- Relationship tables
followers (follower_id, user_id, following_user_id, datetime_added)
hashtag_mappings (hashtagmap_id, hashtag_id, post_id, datetime_added)
```

### TikTok - Video-First Architecture

TikTok uses a distributed system optimized for short-form video content.

**Database Components:**
- **PostgreSQL**: Structured data (user profiles, relationships, video metadata)
- **NoSQL (Cassandra/Redis)**: Unstructured data, user interactions, scalable video metadata
- **Cloud Object Storage (S3/GCS)**: Original video content storage

**Schema Structure:**
```sql
-- Videos table
videos (
  video_id PRIMARY KEY,
  user_id FOREIGN KEY,
  title, description,
  upload_date, views_count,
  duration, metadata
)

-- Interactions tables
likes (like_id, user_id, video_id, timestamp)
comments (comment_id, user_id, video_id, text, timestamp)
```

**Key Features:**
- Chunked, resumable video uploads
- Multi-CDN strategy for content delivery
- Real-time playlist assembly based on preferences
- Apache Hadoop ecosystem for data processing (Flume, Kafka, Storm)

### Snapchat - Ephemeral Content Architecture

Snapchat specializes in disappearing content with unique storage patterns.

**Database Strategy:**
- **Hybrid Approach**: NoSQL (Cassandra/DynamoDB) for user data and media
- **SQL Databases**: PostgreSQL/MySQL for structured data
- **Redis/Memcached**: Caching for frequently accessed data

**Ephemeral Content Handling:**
- Messages stored temporarily in encrypted database files
- Multi-state deletion process:
  - State 1: Active/viewable
  - State 2: Marked for deletion (68% remain 24-72 hours)
  - State 3: Garbage collection
- Unopened messages remain on servers up to 30 days
- Person-to-person Snaps: <10 seconds retention
- Stories: 24 hours retention

**Storage Locations:**
- Android: `/data/data/com.snapchat.android/` (tcspahn.db, arroyo.db, main.db)
- iOS: Encrypted databases (gallery.encrypteddb)

## Common Patterns Across Platforms

### 1. Database Type Separation
All platforms use a **polyglot persistence** approach:
- **SQL databases** (PostgreSQL/MySQL): User data, relationships, structured metadata
- **NoSQL databases** (Cassandra/MongoDB): High-volume data, flexible schemas
- **Cache layers** (Redis/Memcached): Performance optimization
- **Object storage** (S3/GCS): Media files

### 2. Microservices Architecture
Evolution pattern: **Monolith → Microservices**
- User Service
- Post/Content Service
- Media Service
- Interaction Service (likes, comments)
- Feed Service
- Notification Service

### 3. Sharding Strategies
- **Horizontal partitioning** by user_id or post_id
- **Geographic sharding** for data locality
- **Time-based partitioning** for ephemeral content

### 4. Caching Strategies
- **Multi-tier caching**: Application → Redis → Database
- **CDN integration** for media delivery
- **Pre-computed feeds** in cache

## Schema Design Approaches

### Single Table Inheritance (STI) vs Multi-Table Inheritance (MTI)

**Single Table Inheritance:**
- One `posts` table for all content types
- `type` column distinguishes post types
- JSON/JSONB columns for type-specific data

**Advantages:**
- Simpler queries across all post types
- Better performance for feed generation
- Easier to maintain consistency

**Disadvantages:**
- Many NULL columns for type-specific fields
- Table bloat with diverse content types
- Index inefficiency

**Multi-Table Inheritance:**
- Separate tables per content type
- Base `posts` table with common fields
- Type-specific tables with foreign keys

**Advantages:**
- Better for diverse attribute sets
- Easier to add new content types
- More efficient storage

**Disadvantages:**
- Complex queries across types
- JOIN overhead for feed generation

### Polymorphic Associations

Used for flexible relationships (comments, likes on various content types):
```sql
-- Polymorphic comments table
comments (
  id,
  commentable_id,      -- ID of associated record
  commentable_type,    -- Type of entity (post, video, story)
  content,
  user_id
)
```

## Comparison with Alunalun Design

### Current Alunalun Architecture

From the specification files, alunalun uses:

**Database Schema:**
```sql
-- Core posts table (thin parent)
posts (
  id UUID PRIMARY KEY,
  type VARCHAR(20) DEFAULT 'pin',  -- STI approach
  author_user_id UUID,
  author_session_id UUID,
  content TEXT,
  created_at, updated_at, deleted_at
)

-- Separate location data
post_locations (
  post_id UUID PRIMARY KEY,
  location GEOMETRY(POINT, 4326),
  geohash VARCHAR(12),
  longitude, latitude,
  place_name, city, country_code
)

-- Media attachments (MTI approach)
post_media (
  id UUID PRIMARY KEY,
  post_id UUID,
  media_type VARCHAR(20),
  media_url TEXT,
  thumbnail_url TEXT,
  width, height, file_size_bytes,
  duration_seconds, display_order
)

-- Reactions (polymorphic-like)
post_reactions (
  id UUID,
  post_id UUID,
  user_id UUID,
  session_id UUID,
  reaction_type VARCHAR(20)
)

-- Comments with soft delete
post_comments (
  id UUID,
  post_id UUID,
  parent_comment_id UUID,
  content TEXT,
  deleted_at TIMESTAMP,
  depth INTEGER,
  path LTREE  -- For tree queries
)
```

### Strengths of Alunalun's Design

1. **Hybrid Approach**: Combines STI (post types) with MTI (media attachments)
2. **PostGIS Integration**: Native spatial data support
3. **Geohash Optimization**: Pre-computed for performance
4. **LTREE for Comments**: Efficient hierarchical queries
5. **Soft Delete Pattern**: Preserves conversation threads
6. **Anonymous Support**: Dual author system (user_id/session_id)

### Areas for Potential Improvement

Based on platform research:

1. **Media Storage Strategy**
   - Current: Single `post_media` table
   - Consider: Separate tables for different media types (photos, videos, audio)
   - Benefit: Type-specific optimizations and attributes

2. **Caching Layer**
   - Current: No explicit caching mentioned
   - Consider: Redis for feed caching, hot posts
   - Benefit: Reduced database load

3. **Ephemeral Content Support**
   - Current: Only soft delete
   - Consider: TTL-based partitioning for stories/temporary posts
   - Benefit: Automatic cleanup, storage optimization

4. **Feed Optimization**
   - Current: Complex JOINs for feed generation
   - Consider: Pre-computed feed tables or materialized views
   - Benefit: Faster feed loading

5. **Sharding Preparation**
   - Current: Single database
   - Consider: Design for future sharding (user_id, geohash-based)
   - Benefit: Horizontal scalability

## Recommendations

### 1. Immediate Optimizations

**Add Caching Layer:**
```sql
-- Redis cache structure
feed:{user_id}          -- User's personalized feed
post:{post_id}          -- Individual post data
trending:{geohash}      -- Trending posts by location
user:{user_id}:posts    -- User's own posts
```

**Optimize Media Handling:**
```sql
-- Separate video metadata
CREATE TABLE post_videos (
  id UUID PRIMARY KEY,
  post_id UUID REFERENCES posts(id),
  duration_seconds INTEGER,
  resolution VARCHAR(20),
  codec VARCHAR(20),
  bitrate INTEGER,
  thumbnail_url TEXT,
  hls_manifest_url TEXT,  -- For streaming
  processing_status VARCHAR(20)
);
```

### 2. Medium-term Enhancements

**Implement Feed Pre-computation:**
```sql
-- Materialized view for hot posts
CREATE MATERIALIZED VIEW hot_posts AS
SELECT p.*, 
       COUNT(DISTINCT pr.id) as reaction_count,
       COUNT(DISTINCT pc.id) as comment_count,
       (reaction_weight + comment_weight * 2) * time_decay as score
FROM posts p
LEFT JOIN post_reactions pr ON p.id = pr.post_id
LEFT JOIN post_comments pc ON p.id = pc.post_id
WHERE p.created_at > NOW() - INTERVAL '48 hours'
GROUP BY p.id
ORDER BY score DESC;

-- Refresh periodically
REFRESH MATERIALIZED VIEW CONCURRENTLY hot_posts;
```

**Add Ephemeral Content Support:**
```sql
-- Stories/ephemeral posts
CREATE TABLE ephemeral_posts (
  id UUID PRIMARY KEY,
  post_id UUID REFERENCES posts(id),
  expires_at TIMESTAMP NOT NULL,
  view_once BOOLEAN DEFAULT FALSE,
  viewers UUID[] -- Track who viewed
) PARTITION BY RANGE (expires_at);

-- Auto-create monthly partitions
CREATE TABLE ephemeral_posts_2025_01 
  PARTITION OF ephemeral_posts
  FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
```

### 3. Long-term Architecture Evolution

**Prepare for Microservices:**
- Separate databases for users, posts, media, interactions
- Event-driven architecture with Kafka/RabbitMQ
- Service mesh for inter-service communication

**Scale-out Strategy:**
```sql
-- Sharding by geohash prefix
CREATE TABLE posts_shard_us_west
  (CHECK (LEFT(geohash, 2) IN ('9q', '9p')))
  INHERITS (posts);

CREATE TABLE posts_shard_us_east
  (CHECK (LEFT(geohash, 2) IN ('dr', 'dq')))
  INHERITS (posts);
```

### 4. Multi-Modal Post Type Extensions

Based on platform patterns, consider supporting:

```sql
-- Extended post types
ALTER TYPE post_type ADD VALUE 'story';      -- 24hr ephemeral
ALTER TYPE post_type ADD VALUE 'reel';       -- Short video
ALTER TYPE post_type ADD VALUE 'live';       -- Live stream
ALTER TYPE post_type ADD VALUE 'poll';       -- Interactive polls
ALTER TYPE post_type ADD VALUE 'carousel';   -- Multi-image

-- Type-specific metadata
CREATE TABLE post_type_metadata (
  post_id UUID PRIMARY KEY REFERENCES posts(id),
  metadata JSONB NOT NULL
  -- Examples:
  -- Story: {"expires_at": "...", "music_track": "..."}
  -- Reel: {"duration": 30, "effects": [...]}
  -- Live: {"stream_key": "...", "viewer_count": 0}
  -- Poll: {"options": [...], "votes": {...}}
);
```

## Conclusion

Alunalun's current design shows thoughtful architecture with strong spatial features and hierarchical comment support. The main opportunities for improvement lie in:

1. **Performance optimization** through caching and feed pre-computation
2. **Media handling** sophistication for video-first content
3. **Ephemeral content** support for stories and temporary posts
4. **Scale preparation** through sharding-ready design

The platform research reveals that successful social media architectures evolve from simple to complex, maintaining a balance between feature richness and performance. Alunalun can benefit from adopting proven patterns while maintaining its unique location-based focus.

## References

- Facebook TAO: Engineering at Meta Blog (2013)
- Instagram Architecture: Various engineering blogs (2024-2025)
- TikTok System Design: Multiple technical analyses (2023-2024)
- Snapchat Ephemeral Architecture: Security research papers (2023)
- Database Design Patterns: Martin Fowler's Enterprise Application Architecture