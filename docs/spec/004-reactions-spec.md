# Reactions/Engagements Feature

## Overview
Flexible reaction system supporting multiple reaction types (like, love, wow, etc.) for posts. Handles both authenticated and anonymous users, prevents duplicate reactions, and provides aggregated counts for display.

## Database Schema

### Migration: `006_reactions.sql`

```sql
-- Post reactions/engagements
CREATE TABLE post_reactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  
  -- Reactor identity (one must be set)
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  session_id UUID REFERENCES anonymous_sessions(session_id) ON DELETE CASCADE,
  
  reaction_type VARCHAR(20) NOT NULL DEFAULT 'like',
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  
  -- Constraints
  CONSTRAINT reactions_actor_check CHECK (
    (user_id IS NOT NULL AND session_id IS NULL) OR
    (user_id IS NULL AND session_id IS NOT NULL)
  ),
  
  CONSTRAINT reactions_type_check CHECK (
    reaction_type IN ('like', 'love', 'wow', 'haha', 'sad', 'angry')
  )
);

-- Prevent duplicate reactions
CREATE UNIQUE INDEX idx_reactions_user_unique 
  ON post_reactions(post_id, user_id, reaction_type) 
  WHERE user_id IS NOT NULL;

CREATE UNIQUE INDEX idx_reactions_session_unique 
  ON post_reactions(post_id, session_id, reaction_type) 
  WHERE session_id IS NOT NULL;

-- Performance indexes
CREATE INDEX idx_reactions_post ON post_reactions(post_id, reaction_type);
CREATE INDEX idx_reactions_user ON post_reactions(user_id, created_at DESC) 
  WHERE user_id IS NOT NULL;
CREATE INDEX idx_reactions_session ON post_reactions(session_id, created_at DESC) 
  WHERE session_id IS NOT NULL;
CREATE INDEX idx_reactions_created ON post_reactions(created_at DESC);

-- Aggregated view for quick counts
CREATE MATERIALIZED VIEW post_reaction_counts AS
SELECT 
  post_id,
  reaction_type,
  COUNT(*) as count,
  COUNT(DISTINCT user_id) as user_count,
  COUNT(DISTINCT session_id) as anon_count
FROM post_reactions
GROUP BY post_id, reaction_type;

CREATE UNIQUE INDEX idx_reaction_counts ON post_reaction_counts(post_id, reaction_type);

-- Refresh trigger for materialized view (optional for MVP)
CREATE OR REPLACE FUNCTION refresh_reaction_counts() RETURNS TRIGGER AS $$
BEGIN
  REFRESH MATERIALIZED VIEW CONCURRENTLY post_reaction_counts;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Optional: Activity tracking for analytics
CREATE TABLE reaction_activity (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  
  -- Track both additions and removals
  action VARCHAR(10) NOT NULL, -- 'added', 'removed', 'changed'
  reaction_type VARCHAR(20) NOT NULL,
  previous_type VARCHAR(20), -- For 'changed' actions
  
  -- Actor
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  session_id UUID REFERENCES anonymous_sessions(session_id) ON DELETE SET NULL,
  
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  
  CONSTRAINT activity_action_check CHECK (
    action IN ('added', 'removed', 'changed')
  )
);

CREATE INDEX idx_reaction_activity_post ON reaction_activity(post_id, created_at DESC);
CREATE INDEX idx_reaction_activity_user ON reaction_activity(user_id, created_at DESC) 
  WHERE user_id IS NOT NULL;
```

## SQLC Queries

### File: `sql/queries/reactions.sql`

```sql
-- name: AddReaction :one
INSERT INTO post_reactions (
  post_id, user_id, session_id, reaction_type
) VALUES (
  $1, $2, $3, $4
)
ON CONFLICT (post_id, user_id, reaction_type) 
  WHERE user_id IS NOT NULL
  DO UPDATE SET created_at = NOW()
ON CONFLICT (post_id, session_id, reaction_type) 
  WHERE session_id IS NOT NULL
  DO UPDATE SET created_at = NOW()
RETURNING *;

-- name: RemoveReaction :exec
DELETE FROM post_reactions
WHERE post_id = $1 
  AND reaction_type = $2
  AND ((user_id = $3 AND $3 IS NOT NULL) OR 
       (session_id = $4 AND $4 IS NOT NULL));

-- name: RemoveAllReactions :exec
-- Remove all reactions from a user for a post
DELETE FROM post_reactions
WHERE post_id = $1
  AND ((user_id = $2 AND $2 IS NOT NULL) OR 
       (session_id = $3 AND $3 IS NOT NULL));

-- name: ChangeReaction :one
-- Atomic reaction change (remove old, add new)
WITH deleted AS (
  DELETE FROM post_reactions
  WHERE post_id = $1
    AND ((user_id = $2 AND $2 IS NOT NULL) OR 
         (session_id = $3 AND $3 IS NOT NULL))
  RETURNING reaction_type as old_type
)
INSERT INTO post_reactions (
  post_id, user_id, session_id, reaction_type
) VALUES (
  $1, $2, $3, $4
)
RETURNING *, (SELECT old_type FROM deleted) as previous_type;

-- name: GetPostReactions :many
-- Get all reactions for a post grouped by type
SELECT 
  reaction_type,
  COUNT(*) as count,
  COUNT(DISTINCT user_id) as verified_count,
  COUNT(DISTINCT session_id) as anonymous_count
FROM post_reactions
WHERE post_id = $1
GROUP BY reaction_type
ORDER BY count DESC;

-- name: GetUserReactionsForPost :many
-- Get current user's reactions for a post
SELECT reaction_type
FROM post_reactions
WHERE post_id = $1
  AND ((user_id = $2 AND $2 IS NOT NULL) OR 
       (session_id = $3 AND $3 IS NOT NULL));

-- name: GetUserReactionsForPosts :many
-- Batch get user's reactions for multiple posts (feed optimization)
SELECT 
  post_id,
  ARRAY_AGG(reaction_type) as reaction_types
FROM post_reactions
WHERE post_id = ANY($1::uuid[])
  AND ((user_id = $2 AND $2 IS NOT NULL) OR 
       (session_id = $3 AND $3 IS NOT NULL))
GROUP BY post_id;

-- name: GetPostsWithReactionCounts :many
-- Optimized query for feed with reaction counts
SELECT 
  p.*,
  COALESCE(r.total_reactions, 0) as reaction_count,
  COALESCE(r.reaction_breakdown, '{}') as reaction_breakdown,
  COALESCE(ur.user_reactions, '{}') as user_reactions
FROM posts p
LEFT JOIN (
  SELECT 
    post_id,
    COUNT(*) as total_reactions,
    jsonb_object_agg(reaction_type, count) as reaction_breakdown
  FROM (
    SELECT post_id, reaction_type, COUNT(*) as count
    FROM post_reactions
    GROUP BY post_id, reaction_type
  ) grouped
  GROUP BY post_id
) r ON p.id = r.post_id
LEFT JOIN (
  SELECT 
    post_id,
    jsonb_agg(reaction_type) as user_reactions
  FROM post_reactions
  WHERE (user_id = $1 OR session_id = $2)
  GROUP BY post_id
) ur ON p.id = ur.post_id
WHERE p.id = ANY($3::uuid[]);

-- name: GetTopReactors :many
-- Get users who react most to a specific user's posts
SELECT 
  COALESCE(u.username, 'Anonymous') as reactor_name,
  pr.user_id,
  COUNT(*) as reaction_count,
  ARRAY_AGG(DISTINCT pr.reaction_type) as reaction_types
FROM post_reactions pr
JOIN posts p ON pr.post_id = p.id
LEFT JOIN users u ON pr.user_id = u.id
WHERE p.author_user_id = $1
  AND pr.created_at > NOW() - INTERVAL '30 days'
GROUP BY pr.user_id, u.username
ORDER BY reaction_count DESC
LIMIT 10;

-- name: GetReactionActivity :many
-- Get recent reaction activity for a post
SELECT 
  ra.action,
  ra.reaction_type,
  ra.previous_type,
  ra.created_at,
  COALESCE(u.username, 'Anonymous') as actor_name,
  (ra.user_id IS NOT NULL) as is_verified
FROM reaction_activity ra
LEFT JOIN users u ON ra.user_id = u.id
WHERE ra.post_id = $1
ORDER BY ra.created_at DESC
LIMIT 20;

-- name: GetUserReactionHistory :many
-- Get a user's reaction history
SELECT 
  p.id as post_id,
  p.content,
  pr.reaction_type,
  pr.created_at
FROM post_reactions pr
JOIN posts p ON pr.post_id = p.id
WHERE (pr.user_id = $1 OR pr.session_id = $2)
ORDER BY pr.created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountReactionsByType :one
-- Analytics: Count total reactions by type
SELECT 
  reaction_type,
  COUNT(*) as total_count,
  COUNT(DISTINCT post_id) as post_count,
  COUNT(DISTINCT user_id) as user_count
FROM post_reactions
WHERE created_at > $1
GROUP BY reaction_type
ORDER BY total_count DESC;
