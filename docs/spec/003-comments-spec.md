# Comments System Feature

## Overview
Hierarchical comment system with soft delete to preserve conversation threads. Supports nested replies, maintains thread integrity when parent comments are deleted, and handles both authenticated and anonymous users.

## Database Schema

### Migration: `005_comments.sql`

```sql
-- Hierarchical comments with soft delete
CREATE TABLE post_comments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  parent_comment_id UUID REFERENCES post_comments(id) ON DELETE RESTRICT,
  
  -- Author (one of these must be set)
  author_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  author_session_id UUID REFERENCES anonymous_sessions(session_id) ON DELETE SET NULL,
  
  content TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  
  -- Soft delete fields
  deleted_at TIMESTAMP WITH TIME ZONE,
  deleted_by VARCHAR(20), -- 'author', 'moderator', 'system'
  
  -- Denormalized for performance
  depth INTEGER NOT NULL DEFAULT 0, -- 0 for root, 1+ for replies
  path LTREE, -- For efficient tree queries (e.g., '0001.0003.0002')
  
  -- Constraints
  CONSTRAINT comments_author_check CHECK (
    (author_user_id IS NOT NULL AND author_session_id IS NULL) OR
    (author_user_id IS NULL AND author_session_id IS NOT NULL)
  ),
  
  CONSTRAINT comments_deleted_by_check CHECK (
    deleted_by IN ('author', 'moderator', 'system')
  )
);

-- Helper view for easier querying
CREATE VIEW comments_display AS
SELECT 
  c.id,
  c.post_id,
  c.parent_comment_id,
  c.author_user_id,
  c.author_session_id,
  CASE 
    WHEN c.deleted_at IS NOT NULL THEN '[deleted]'
    ELSE c.content
  END as display_content,
  c.content as original_content,
  c.created_at,
  c.updated_at,
  c.deleted_at,
  c.deleted_by,
  c.depth,
  c.path,
  COALESCE(u.username, 'Anonymous') as author_name,
  (c.author_user_id IS NOT NULL) as is_verified_author,
  -- Check if this is orphaned (parent deleted but has replies)
  CASE 
    WHEN c.parent_comment_id IS NULL THEN 'root'
    WHEN p.deleted_at IS NOT NULL THEN 'orphaned_reply'
    ELSE 'reply'
  END as comment_type
FROM post_comments c
LEFT JOIN post_comments p ON c.parent_comment_id = p.id
LEFT JOIN users u ON c.author_user_id = u.id;

-- Indexes for performance
CREATE INDEX idx_comments_post ON post_comments(post_id, created_at DESC) 
  WHERE deleted_at IS NULL;
CREATE INDEX idx_comments_parent ON post_comments(parent_comment_id, created_at DESC);
CREATE INDEX idx_comments_author_user ON post_comments(author_user_id, created_at DESC) 
  WHERE deleted_at IS NULL;
CREATE INDEX idx_comments_path ON post_comments USING GIST(path);

-- Trigger to maintain path on insert
CREATE OR REPLACE FUNCTION maintain_comment_path() RETURNS TRIGGER AS $$
DECLARE
  parent_path LTREE;
  new_segment TEXT;
BEGIN
  IF NEW.parent_comment_id IS NULL THEN
    -- Root comment: generate new top-level path
    SELECT COALESCE(MAX(path::text)::ltree, '0')::text::int + 1 
    INTO new_segment
    FROM post_comments 
    WHERE post_id = NEW.post_id AND parent_comment_id IS NULL;
    
    NEW.path = new_segment::text::ltree;
    NEW.depth = 0;
  ELSE
    -- Reply: append to parent's path
    SELECT path, depth + 1 
    INTO parent_path, NEW.depth
    FROM post_comments 
    WHERE id = NEW.parent_comment_id;
    
    SELECT COALESCE(MAX(subpath(path, -1)::text::int), 0) + 1
    INTO new_segment
    FROM post_comments
    WHERE parent_comment_id = NEW.parent_comment_id;
    
    NEW.path = parent_path || new_segment::text::ltree;
  END IF;
  
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER maintain_comment_path_trigger
  BEFORE INSERT ON post_comments
  FOR EACH ROW
  EXECUTE FUNCTION maintain_comment_path();
```

## SQLC Queries

### File: `sql/queries/comments.sql`

```sql
-- name: CreateComment :one
INSERT INTO post_comments (
  post_id, parent_comment_id, author_user_id, author_session_id, content
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetCommentByID :one
SELECT 
  c.*,
  COALESCE(u.username, 'Anonymous') as author_name,
  (c.author_user_id IS NOT NULL) as is_verified_author
FROM post_comments c
LEFT JOIN users u ON c.author_user_id = u.id
WHERE c.id = $1;

-- name: GetCommentsByPost :many
-- Get comments with tree structure preserved
WITH RECURSIVE comment_tree AS (
  -- Root comments (including soft-deleted if they have replies)
  SELECT 
    c.*,
    COALESCE(u.username, 'Anonymous') as author_name,
    (c.author_user_id IS NOT NULL) as is_verified_author,
    ARRAY[c.created_at] as sort_path
  FROM post_comments c
  LEFT JOIN users u ON c.author_user_id = u.id
  WHERE c.post_id = $1 
    AND c.parent_comment_id IS NULL
    AND (
      c.deleted_at IS NULL OR 
      EXISTS (SELECT 1 FROM post_comments WHERE parent_comment_id = c.id)
    )
  
  UNION ALL
  
  -- Recursive replies
  SELECT 
    c.*,
    COALESCE(u.username, 'Anonymous') as author_name,
    (c.author_user_id IS NOT NULL) as is_verified_author,
    ct.sort_path || c.created_at
  FROM post_comments c
  JOIN comment_tree ct ON c.parent_comment_id = ct.id
  LEFT JOIN users u ON c.author_user_id = u.id
  WHERE c.depth <= $2 -- Max depth parameter
)
SELECT * FROM comment_tree
ORDER BY sort_path;

-- name: GetCommentsByPostFlat :many
-- Flat list with pagination
SELECT 
  c.*,
  COALESCE(u.username, 'Anonymous') as author_name,
  (c.author_user_id IS NOT NULL) as is_verified_author,
  CASE 
    WHEN c.deleted_at IS NOT NULL THEN '[deleted]'
    ELSE c.content
  END as display_content
FROM post_comments c
LEFT JOIN users u ON c.author_user_id = u.id
WHERE c.post_id = $1
ORDER BY c.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetCommentReplies :many
-- Get direct replies to a comment
SELECT 
  c.*,
  COALESCE(u.username, 'Anonymous') as author_name,
  (c.author_user_id IS NOT NULL) as is_verified_author
FROM post_comments c
LEFT JOIN users u ON c.author_user_id = u.id
WHERE c.parent_comment_id = $1
  AND (c.deleted_at IS NULL OR 
       EXISTS (SELECT 1 FROM post_comments WHERE parent_comment_id = c.id))
ORDER BY c.created_at ASC;

-- name: UpdateComment :one
UPDATE post_comments
SET 
  content = $2,
  updated_at = NOW()
WHERE id = $1
  AND (author_user_id = $3 OR author_session_id = $4)
  AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteComment :exec
-- Soft delete preserves thread structure
UPDATE post_comments
SET 
  deleted_at = NOW(),
  deleted_by = CASE 
    WHEN author_user_id = $2 OR author_session_id = $3 THEN 'author'
    WHEN $4 = true THEN 'moderator'
    ELSE 'system'
  END,
  content = '[deleted]'
WHERE id = $1
  AND (
    author_user_id = $2 
    OR author_session_id = $3
    OR $4 = true -- is_moderator
  );

-- name: CountCommentsByPost :one
SELECT COUNT(*) as total
FROM post_comments
WHERE post_id = $1 AND deleted_at IS NULL;

-- name: CountRepliesByComment :one
SELECT COUNT(*) as reply_count
FROM post_comments
WHERE parent_comment_id = $1 AND deleted_at IS NULL;

-- name: GetUserComments :many
-- Get all comments by a user
SELECT 
  c.*,
  p.content as post_content,
  p.type as post_type
FROM post_comments c
JOIN posts p ON c.post_id = p.id
WHERE (c.author_user_id = $1 OR c.author_session_id = $2)
  AND c.deleted_at IS NULL
ORDER BY c.created_at DESC
LIMIT $3 OFFSET $4;

-- name: CheckCommentOwnership :one
-- Verify if user owns a comment
SELECT EXISTS(
  SELECT 1 FROM post_comments
  WHERE id = $1
    AND (author_user_id = $2 OR author_session_id = $3)
) as is_owner;
```

## Protocol Buffer Definitions

### File: `proto/v1/entities/comment.proto`

```protobuf
syntax = "proto3";

package entities.v1;

import "google/protobuf/timestamp.proto";

message Comment {
  string id = 1;
  string post_id = 2;
  string parent_comment_id = 3;
  
  // Author info
  string author_user_id = 4;
  string author_session_id = 5;
  string author_username = 6;
  bool is_verified_author = 7;
  
  // Content
  string content = 8;
  string display_content = 9; // Shows "[deleted]" if deleted
  
  // Metadata
  google.protobuf.Timestamp created_at = 10;
  google.protobuf.Timestamp updated_at = 11;
  
  // Deletion info
  bool is_deleted = 12;
  google.protobuf.Timestamp deleted_at = 13;
  string deleted_by = 14;
  
  // Tree structure
  int32 depth = 15;
  string comment_type = 16; // "root", "reply", "orphaned_reply"
  
  // Nested replies (for tree display)
  repeated Comment replies = 17;
  int32 reply_count = 18;
}

message CommentThread {
  repeated Comment root_comments = 1;
  int32 total_count = 2;
  bool has_more = 3;
}
```

### File: `proto/v1/service/comments.proto`

```protobuf
syntax = "proto3";

package service.v1;

import "entities/v1/comment.proto";

service CommentService {
  // CRUD operations
  rpc AddComment(AddCommentRequest) returns (AddCommentResponse);
  rpc UpdateComment(UpdateCommentRequest) returns (UpdateCommentResponse);
  rpc DeleteComment(DeleteCommentRequest) returns (DeleteCommentResponse);
  
  // Query operations
  rpc GetComment(GetCommentRequest) returns (GetCommentResponse);
  rpc GetPostComments(GetPostCommentsRequest) returns (GetPostCommentsResponse);
  rpc GetCommentReplies(GetCommentRepliesRequest) returns (GetCommentRepliesResponse);
  rpc GetUserComments(GetUserCommentsRequest) returns (GetUserCommentsResponse);
}

message AddCommentRequest {
  string post_id = 1;
  string parent_comment_id = 2; // Optional for replies
  string content = 3;
}

message AddCommentResponse {
  entities.v1.Comment comment = 1;
  int32 new_total_count = 2; // Updated count for the post
}

message UpdateCommentRequest {
  string comment_id = 1;
  string content = 2;
}

message UpdateCommentResponse {
  entities.v1.Comment comment = 1;
}

message DeleteCommentRequest {
  string comment_id = 1;
}

message DeleteCommentResponse {
  bool success = 1;
  string message = 2;
}

message GetPostCommentsRequest {
  string post_id = 1;
  
  // Display options
  int32 max_depth = 2; // How deep to fetch replies (0 = root only)
  bool include_deleted = 3; // Show deleted comments as placeholders
  
  // Pagination
  int32 limit = 4;
  int32 offset = 5;
  
  // Sort
  enum SortBy {
    OLDEST = 0;
    NEWEST = 1;
    TOP = 2; // Most replies
  }
  SortBy sort_by = 6;
}

message GetPostCommentsResponse {
  entities.v1.CommentThread thread = 1;
  string next_cursor = 2;
}
```

## Backend Service Implementation

### File: `internal/services/comments/service.go`

```go
package comments

import (
    "context"
    "database/sql"
    "errors"
    
    "connectrpc.com/connect"
    
    "yourproject/internal/repository"
    pbentities "yourproject/proto/gen/v1/entities"
    pbservice "yourproject/proto/gen/v1/service"
)

type Service struct {
    pbservice.UnimplementedCommentServiceServer
    db *repository.Queries
}

func NewService(db *repository.Queries) *Service {
    return &Service{db: db}
}

// AddComment creates a new comment or reply
func (s *Service) AddComment(
    ctx context.Context,
    req *connect.Request[pbservice.AddCommentRequest],
) (*connect.Response[pbservice.AddCommentResponse], error) {
    userID, _ := ctx.Value("user_id").(string)
    sessionID, _ := ctx.Value("session_id").(string)
    
    if userID == "" && sessionID == "" {
        return nil, connect.NewError(connect.CodeUnauthenticated, nil)
    }
    
    // Validate parent comment exists if replying
    if req.Msg.ParentCommentId != "" {
        parent, err := s.db.GetCommentByID(ctx, uuid.MustParse(req.Msg.ParentCommentId))
        if err != nil {
            return nil, connect.NewError(connect.CodeNotFound, 
                errors.New("parent comment not found"))
        }
        
        // Check depth limit (e.g., max 3 levels)
        if parent.Depth >= 2 {
            return nil, connect.NewError(connect.CodeInvalidArgument,
                errors.New("maximum reply depth reached"))
        }
    }
    
    // Create comment
    comment, err := s.db.CreateComment(ctx, repository.CreateCommentParams{
        PostID:          uuid.MustParse(req.Msg.PostId),
        ParentCommentID: sql.NullString{
            String: req.Msg.ParentCommentId,
            Valid:  req.Msg.ParentCommentId != "",
        },
        AuthorUserID:    sql.NullString{String: userID, Valid: userID != ""},
        AuthorSessionID: sql.NullString{String: sessionID, Valid: sessionID != ""},
        Content:         req.Msg.Content,
    })
    
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    
    // Get updated count
    count, _ := s.db.CountCommentsByPost(ctx, uuid.MustParse(req.Msg.PostId))
    
    return connect.NewResponse(&pbservice.AddCommentResponse{
        Comment:       s.convertToProtoComment(comment),
        NewTotalCount: int32(count),
    }), nil
}

// DeleteComment performs soft delete to preserve thread
func (s *Service) DeleteComment(
    ctx context.Context,
    req *connect.Request[pbservice.DeleteCommentRequest],
) (*connect.Response[pbservice.DeleteCommentResponse], error) {
    userID, _ := ctx.Value("user_id").(string)
    sessionID, _ := ctx.Value("session_id").(string)
    isModerator := ctx.Value("is_moderator").(bool)
    
    // Check ownership
    isOwner, err := s.db.CheckCommentOwnership(ctx, repository.CheckCommentOwnershipParams{
        ID:              uuid.MustParse(req.Msg.CommentId),
        AuthorUserID:    sql.NullString{String: userID, Valid: userID != ""},
        AuthorSessionID: sql.NullString{String: sessionID, Valid: sessionID != ""},
    })
    
    if !isOwner && !isModerator {
        return nil, connect.NewError(connect.CodePermissionDenied, nil)
    }
    
    // Soft delete
    err = s.db.SoftDeleteComment(ctx, repository.SoftDeleteCommentParams{
        ID:              uuid.MustParse(req.Msg.CommentId),
        AuthorUserID:    sql.NullString{String: userID, Valid: userID != ""},
        AuthorSessionID: sql.NullString{String: sessionID, Valid: sessionID != ""},
        IsModerator:     isModerator,
    })
    
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    
    return connect.NewResponse(&pbservice.DeleteCommentResponse{
        Success: true,
        Message: "Comment deleted",
    }), nil
}

// GetPostComments returns threaded comments for a post
func (s *Service) GetPostComments(
    ctx context.Context,
    req *connect.Request[pbservice.GetPostCommentsRequest],
) (*connect.Response[pbservice.GetPostCommentsResponse], error) {
    maxDepth := req.Msg.MaxDepth
    if maxDepth == 0 {
        maxDepth = 2 // Default depth
    }
    
    comments, err := s.db.GetCommentsByPost(ctx, repository.GetCommentsByPostParams{
        PostID:   uuid.MustParse(req.Msg.PostId),
        MaxDepth: maxDepth,
    })
    
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    
    // Build tree structure
    thread := s.buildCommentTree(comments)
    
    return connect.NewResponse(&pbservice.GetPostCommentsResponse{
        Thread: thread,
    }), nil
}

// Helper to build nested comment structure
func (s *Service) buildCommentTree(comments []repository.Comment) *pbentities.CommentThread {
    // Map for O(1) lookups
    commentMap := make(map[string]*pbentities.Comment)
    rootComments := []*pbentities.Comment{}
    
    // First pass: create all comments
    for _, c := range comments {
        protoComment := s.convertToProtoComment(c)
        commentMap[c.ID.String()] = protoComment
    }
    
    // Second pass: build tree
    for _, c := range comments {
        if c.ParentCommentID.Valid {
            parent := commentMap[c.ParentCommentID.String]
            if parent != nil {
                parent.Replies = append(parent.Replies, commentMap[c.ID.String()])
            }
        } else {
            rootComments = append(rootComments, commentMap[c.ID.String()])
        }
    }
    
    return &pbentities.CommentThread{
        RootComments: rootComments,
        TotalCount:   int32(len(comments)),
    }
}
```

## Frontend Implementation

### File: `features/comments/hooks/use-post-comments.ts`

```typescript
import { useQuery } from '@tanstack/react-query';
import { useCommentServiceClient } from '@/common/services/connectrpc';

interface UsePostCommentsOptions {
  maxDepth?: number;
  includeDeleted?: boolean;
  sortBy?: 'oldest' | 'newest' | 'top';
}

export function usePostComments(
  postId: string,
  options: UsePostCommentsOptions = {}
) {
  const client = useCommentServiceClient();
  
  return useQuery({
    queryKey: ['comments', postId, options],
    queryFn: async () => {
      const response = await client.getPostComments({
        post_id: postId,
        max_depth: options.maxDepth ?? 2,
        include_deleted: options.includeDeleted ?? true,
        sort_by: options.sortBy ?? 'oldest',
        limit: 50,
      });
      
      return response.thread;
    },
    enabled: !!postId,
    staleTime: 30 * 1000, // 30 seconds
  });
}
```

### File: `features/comments/hooks/use-add-comment.ts`

```typescript
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useCommentServiceClient } from '@/common/services/connectrpc';

export function useAddComment() {
  const client = useCommentServiceClient();
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({
      postId,
      content,
      parentCommentId,
    }: {
      postId: string;
      content: string;
      parentCommentId?: string;
    }) => {
      const response = await client.addComment({
        post_id: postId,
        parent_comment_id: parentCommentId,
        content,
      });
      
      return response.comment;
    },
    onSuccess: (data, variables) => {
      // Invalidate comment queries for this post
      queryClient.invalidateQueries({ 
        queryKey: ['comments', variables.postId] 
      });
      
      // Update post comment count in feed
      queryClient.setQueriesData(
        { queryKey: ['posts'] },
        (oldData: any) => {
          // Update comment count for the specific post
          // Implementation depends on your data structure
          return oldData;
        }
      );
    },
  });
}
```

### File: `features/comments/components/comment-thread.tsx`

```tsx
import { useState } from 'react';
import { Comment } from '@/proto/gen/v1/entities/comment_pb';
import { CommentItem } from './comment-item';
import { CommentForm } from './comment-form';
import { usePostComments } from '../hooks/use-post-comments';

interface CommentThreadProps {
  postId: string;
  onClose?: () => void;
}

export function CommentThread({ postId, onClose }: CommentThreadProps) {
  const { data: thread, isLoading } = usePostComments(postId);
  const [replyingTo, setReplyingTo] = useState<string | null>(null);
  
  if (isLoading) {
    return <CommentThreadSkeleton />;
  }
  
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between p-4 border-b">
        <h3 className="font-semibold">
          Comments ({thread?.total_count || 0})
        </h3>
        {onClose && (
          <button onClick={onClose} className="text-gray-500">
            ✕
          </button>
        )}
      </div>
      
      <div className="p-4">
        <CommentForm postId={postId} />
      </div>
      
      <div className="divide-y">
        {thread?.root_comments.map(comment => (
          <CommentItem
            key={comment.id}
            comment={comment}
            onReply={(id) => setReplyingTo(id)}
            replyingTo={replyingTo}
            onCancelReply={() => setReplyingTo(null)}
            postId={postId}
          />
        ))}
      </div>
      
      {!thread?.root_comments.length && (
        <p className="text-center text-gray-500 py-8">
          No comments yet. Be the first to comment!
        </p>
      )}
    </div>
  );
}
```

### File: `features/comments/components/comment-item.tsx`

```tsx
import { useState } from 'react';
import { Comment } from '@/proto/gen/v1/entities/comment_pb';
import { CommentForm } from './comment-form';
import { useDeleteComment } from '../hooks/use-delete-comment';
import { formatTimeAgo } from '@/common/lib/time';

interface CommentItemProps {
  comment: Comment;
  onReply: (commentId: string) => void;
  replyingTo: string | null;
  onCancelReply: () => void;
  postId: string;
  depth?: number;
}

export function CommentItem({
  comment,
  onReply,
  replyingTo,
  onCancelReply,
  postId,
  depth = 0,
}: CommentItemProps) {
  const [isEditing, setIsEditing] = useState(false);
  const { mutate: deleteComment } = useDeleteComment();
  const isDeleted = comment.is_deleted;
  const maxDepth = 2; // Limit nesting depth
  
  return (
    <div className={`${depth > 0 ? 'ml-8 border-l-2 pl-4' : ''}`}>
      <div className="py-3">
        {/* Comment header */}
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2 text-sm">
            <span className="font-medium">
              {comment.author_username}
              {comment.is_verified_author && (
                <span className="ml-1 text-blue-500">✓</span>
              )}
            </span>
            <span className="text-gray-500">
              {formatTimeAgo(comment.created_at)}
            </span>
            {isDeleted && (
              <span className="text-gray-400 italic">
                deleted by {comment.deleted_by}
              </span>
            )}
          </div>
          
          {!isDeleted && (
            <CommentActions
              commentId={comment.id}
              onEdit={() => setIsEditing(true)}
              onDelete={() => deleteComment({ commentId: comment.id })}
              onReply={() => onReply(comment.id)}
              canReply={depth < maxDepth}
            />
          )}
        </div>
        
        {/* Comment content */}
        {isEditing ? (
          <CommentEditForm
            commentId={comment.id}
            initialContent={comment.content}
            onCancel={() => setIsEditing(false)}
            onSuccess={() => setIsEditing(false)}
          />
        ) : (
          <p className={`text-gray-900 ${isDeleted ? 'italic text-gray-500' : ''}`}>
            {comment.display_content}
          </p>
        )}
        
        {/* Reply form (inline) */}
        {replyingTo === comment.id && (
          <div className="mt-3">
            <CommentForm
              postId={postId}
              parentCommentId={comment.id}
              onSuccess={onCancelReply}
              onCancel={onCancelReply}
              placeholder={`Reply to ${comment.author_username}...`}
            />
          </div>
        )}
        
        {/* Nested replies */}
        {comment.replies && comment.replies.length > 0 && (
          <div className="mt-3 space-y-2">
            {comment.replies.map(reply => (
              <CommentItem
                key={reply.id}
                comment={reply}
                onReply={onReply}
                replyingTo={replyingTo}
                onCancelReply={onCancelReply}
                postId={postId}
                depth={depth + 1}
              />
            ))}
          </div>
        )}
        
        {/* Show collapsed indicator if there are hidden replies */}
        {comment.reply_count > (comment.replies?.length || 0) && (
          <button className="mt-2 text-sm text-blue-500 hover:underline">
            View {comment.reply_count - (comment.replies?.length || 0)} more replies
          </button>
        )}
      </div>
    </div>
  );
}
```

### File: `features/comments/components/comment-form.tsx`

```tsx
import { useState } from 'react';
import { useAddComment } from '../hooks/use-add-comment';

interface CommentFormProps {
  postId: string;
  parentCommentId?: string;
  onSuccess?: () => void;
  onCancel?: () => void;
  placeholder?: string;
}

export function CommentForm({
  postId,
  parentCommentId,
  onSuccess,
  onCancel,
  placeholder = 'Write a comment...',
}: CommentFormProps) {
  const [content, setContent] = useState('');
  const { mutate: addComment, isPending } = useAddComment();
  
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!content.trim()) return;
    
    addComment(
      {
        postId,
        content: content.trim(),
        parentCommentId,
      },
      {
        onSuccess: () => {
          setContent('');
          onSuccess?.();
        },
      }
    );
  };
  
  return (
    <form onSubmit={handleSubmit} className="space-y-2">
      <textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder={placeholder}
        className="w-full p-2 border rounded-lg resize-none"
        rows={parentCommentId ? 2 : 3}
        disabled={isPending}
      />
      
      <div className="flex justify-end gap-2">
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            className="px-3 py-1 text-sm text-gray-600"
          >
            Cancel
          </button>
        )}
        
        <button
          type="submit"
          disabled={!content.trim() || isPending}
          className="px-3 py-1 text-sm bg-blue-500 text-white rounded-lg
                     disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isPending ? 'Posting...' : parentCommentId ? 'Reply' : 'Comment'}
        </button>
      </div>
    </form>
  );
}
```

## Key Implementation Notes

### Thread Preservation Strategy
- Use soft delete (set `deleted_at`) instead of hard delete
- Replace content with "[deleted]" but keep the comment record
- Show deleted comments only if they have replies
- Track who deleted (author, moderator, system)

### Performance Optimizations
- Use LTREE for efficient tree queries in PostgreSQL
- Limit reply depth to prevent deep nesting (max 2-3 levels)
- Denormalize depth and path for faster queries
- Load replies lazily for deep threads

### Tree Structure Management
- Use recursive CTEs for fetching complete threads
- Maintain path using trigger for efficient sorting
- Support both nested and flat display modes
- Cache comment counts at post level

### Reply Depth Limiting
- Frontend: Hide reply button after max depth
- Backend: Reject replies beyond max depth
- Consider flattening deep discussions into root level

### Real-time Considerations (Future)
- Comments are good candidates for real-time updates
- Could use WebSocket/SSE for live comment streams
- For MVP, manual refresh is sufficient

### Moderation Features
- Soft delete preserves evidence for moderation
- Track deletion reason and actor
- Allow moderators to delete any comment
- Consider shadowbanning for problematic users

### User Experience
- Show typing indicators for active repliers (future)
- Auto-save draft comments locally
- Markdown support for formatting (future)
- @mentions for reply notifications (future)
