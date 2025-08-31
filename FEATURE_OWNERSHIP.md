# Feature Ownership Map

Quick reference for who owns what in the parallel development structure.

## Feature Teams

### 🎯 Posts Core Team
**Branch**: `feature/posts-core`  
**Priority**: PHASE 1 - MUST COMPLETE FIRST  
**Owner**: @posts-team  

**Owned Files**:
- `api/internal/services/posts/`
- `api/sql/migrations/004_posts_and_locations.sql`
- `api/sql/queries/posts.sql`
- `api/proto/v1/entities/post.proto`
- `api/proto/v1/service/posts.proto`
- `web/src/features/posts/`
- `web/src/common/clients/posts.ts`

---

### 💬 Comments Team
**Branch**: `feature/comments-system`  
**Priority**: Phase 2  
**Owner**: @comments-team  

**Owned Files**:
- `api/internal/services/comments/`
- `api/sql/migrations/005_comments.sql`
- `api/sql/queries/comments.sql`
- `api/proto/v1/entities/comment.proto`
- `api/proto/v1/service/comments.proto`
- `web/src/features/comments/`
- `web/src/common/clients/comments.ts`

---

### ❤️ Reactions Team
**Branch**: `feature/reactions-system`  
**Priority**: Phase 2  
**Owner**: @reactions-team  

**Owned Files**:
- `api/internal/services/reactions/`
- `api/sql/migrations/006_reactions.sql`
- `api/sql/queries/reactions.sql`
- `api/proto/v1/entities/reaction.proto`
- `api/proto/v1/service/reactions.proto`
- `web/src/features/reactions/`
- `web/src/common/clients/reactions.ts`

---

### 📍 Feed Team
**Branch**: `feature/location-feed`  
**Priority**: Phase 2  
**Owner**: @feed-team  

**Owned Files**:
- `api/internal/services/feed/`
- `api/sql/migrations/007_feed_optimizations.sql`
- `api/sql/queries/feed.sql`
- `api/proto/v1/service/feed.proto`
- `web/src/features/feed/`
- `web/src/common/clients/feed.ts`

---

### 🗺️ Map Team
**Branch**: `feature/map-pins`  
**Priority**: Phase 2  
**Owner**: @map-team  

**Owned Files**:
- `api/internal/services/map/`
- `api/sql/migrations/008_map_optimizations.sql`
- `api/sql/queries/map_pins.sql`
- `api/proto/v1/service/map_pins.proto`
- `web/src/features/map-pins/`
- `web/src/common/clients/map.ts`

---

### 📸 Media Team
**Branch**: `feature/media-upload`  
**Priority**: Phase 2  
**Owner**: @media-team  

**Owned Files**:
- `api/internal/services/media/`
- `api/sql/migrations/009_media_handling.sql`
- `api/sql/queries/media.sql`
- `api/proto/v1/service/media.proto`
- `web/src/features/media/`
- `web/src/common/clients/media.ts`

---

## Shared Files (Merge Points)

These files will be modified by multiple teams. Look for your `[FEATURE-START]` markers:

| File | Purpose | Merge Strategy |
|------|---------|----------------|
| `api/internal/server/server.go` | Service registration | Keep all sections |
| `api/internal/server/routes.go` | Route registration | Keep all sections |
| `api/cmd/server/main.go` | Service initialization | Keep all sections |
| `api/sqlc.yaml` | Query/migration registration | Keep all paths |
| `web/src/app/layout.tsx` | Provider wrappers | Keep all providers |
| `web/src/app/page.tsx` | Component mounting | Keep all components |
| `web/package.json` | Dependencies | Keep all, use latest versions |
| `Makefile` | Build commands | Keep all commands |

## Do's and Don'ts

### ✅ DO's
- Work only in your owned files
- Use your reserved migration number
- Add code between your designated markers
- Keep all sections when resolving conflicts
- Test with posts-core merged before PR

### ❌ DON'TS
- Don't modify other teams' files
- Don't change migration numbers
- Don't remove merge markers
- Don't delete other teams' sections
- Don't merge before posts-core

## Merge Conflict Resolution

When you encounter a conflict in a shared file:

```bash
# 1. See what's conflicting
git status

# 2. For shared files, usually keep both:
git checkout --theirs [file]  # Take their version
# Then manually add your section back

# 3. For your owned files:
git checkout --ours [file]  # Keep your version
```

## Quick Checks

Before creating PR:
- [ ] Only modified files I own?
- [ ] Used correct migration number?
- [ ] Added code between my markers?
- [ ] Tests pass with posts-core?
- [ ] No changes to other features?

## Communication Channels

- **Daily Standup**: 10am in #dev-parallel
- **Blocked?**: Post in #dev-parallel-help
- **Merging soon**: Announce in #dev-parallel 30min before
- **Merge complete**: Update #dev-parallel with status

## Emergency Contacts

- **Posts Core Lead**: @posts-lead
- **DevOps/Merge Conflicts**: @devops-team
- **Architecture Questions**: @tech-lead