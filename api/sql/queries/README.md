# SQL Query Files

Each feature owns their query file. When adding queries:

1. **Stay in your file** - Don't modify other feature's query files
2. **Follow naming conventions** - ActionEntityCondition (e.g., GetPostsByUser)
3. **Add comments** - Document complex queries
4. **Use transactions** - For multi-table operations

## File Ownership

| File | Owner | Feature Branch |
|------|-------|----------------|
| posts.sql | @posts-team | feature/posts-core |
| comments.sql | @comments-team | feature/comments-system |
| reactions.sql | @reactions-team | feature/reactions-system |
| feed.sql | @feed-team | feature/location-feed |
| map_pins.sql | @map-team | feature/map-pins |

## SQLC Configuration

All query files are registered in `sqlc.yaml`. When adding a new query file:
1. Add the path to the queries section
2. Keep the existing order
3. Don't remove other query files

## Common Patterns

### Pagination
```sql
-- name: GetItemsPaginated :many
SELECT * FROM items
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
```

### Soft Delete
```sql
-- name: SoftDeleteItem :exec
UPDATE items
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
```

### Ownership Check
```sql
-- name: CheckItemOwnership :one
SELECT EXISTS(
  SELECT 1 FROM items
  WHERE id = $1 
    AND (author_user_id = $2 OR author_session_id = $3)
) as is_owner;
```