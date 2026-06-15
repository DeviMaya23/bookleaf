## Context

`folder-sharing` introduces the first publicly-accessible (unauthenticated) data endpoint in the backend. Everything under `protected` today requires a Kinde-authenticated user; `/share/:token` must sit outside that group while still going through the same global middleware (CORS, recovery, logging).

The owner-facing share management (create/get/delete) follows the existing per-domain handler/usecase/repository pattern (`trash`, `tag`), and the public read path reuses the same image-listing and presigned-URL machinery `image_usecase.go` already has for `GetImage`/`ListFolderImages`.

## Goals / Non-Goals

**Goals:**
- One share per folder, addressable by an opaque token, independent of the folder's own UUID
- Owner can create, inspect, and revoke a folder's share
- Public endpoint returns folder name + notes + ordered images (title, thumbnail URL, full-res URL) for direct folder members only
- New code follows the consumer-defines-interfaces convention — no widening of existing `FolderRepository`/`FolderImageRepository`/`ImageRepository` interfaces

**Non-Goals:**
- Sharing child/nested folders or aggregating their images
- Multiple/rotatable share links per folder, expiry, or password protection
- Any frontend work (separate proposal)
- Surfacing share status on existing folder list/detail endpoints

## Decisions

### 1. New `folder_shares` table, one row per shared folder
```sql
CREATE TABLE folder_shares (
    id         UUID PRIMARY KEY,
    folder_id  UUID NOT NULL UNIQUE REFERENCES folders(id) ON DELETE CASCADE,
    token      TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```
- `folder_id UNIQUE` enforces "one share per folder" at the DB level and gives `POST .../share` idempotent semantics.
- `ON DELETE CASCADE` means deleting a folder automatically removes its share — no orphaned tokens to clean up.
- Migration `000015_create_folder_shares` (up/down), following the numbering of `000014_add_pending_kinde_deletion_to_users`.

**Alternative considered:** a `shared boolean` / `share_token` column directly on `folders`. Rejected — folder is already a fairly wide entity, and a separate table keeps the share lifecycle (and its cascade-delete) independent of folder updates, matching the proposal's original instinct.

### 2. Token: 16 random bytes, base64 URL-encoded (no padding)
`crypto/rand` → `base64.RawURLEncoding` → 22-character opaque string (128 bits of entropy), generated in `ShareUsecase`, persisted as-is by the repository.

**Alternative considered:** reuse `uuid.New()` (as folders/images do). Rejected to keep share tokens visually and structurally distinct from internal resource IDs — a token leaking shouldn't read as "this looks like a folder ID."

### 3. New `ShareHandler` / `ShareUsecase` / `FolderShareRepository`, following the `trash` pattern
- `backend/internal/domain/folder_share.go` — `FolderShare{ID, FolderID, Token, CreatedAt, Folder Folder}` (`Folder` association for the public read path)
- `backend/internal/usecase/share_usecase.go` — defines:
  ```go
  type FolderShareRepository interface {
      Create(ctx context.Context, folderID uuid.UUID, token string) (*domain.FolderShare, error)
      GetByFolderID(ctx context.Context, folderID uuid.UUID) (*domain.FolderShare, error)
      GetByToken(ctx context.Context, token string) (*domain.FolderShare, error) // preloads Folder
      DeleteByFolderID(ctx context.Context, folderID uuid.UUID) error
  }

  // Narrow interfaces this usecase needs from existing data — NOT the
  // broader FolderRepository/FolderImageRepository interfaces.
  type ShareFolderRepository interface {
      GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Folder, error)
  }

  type ShareImageRepository interface {
      ListByFolder(ctx context.Context, userID string, folderID uuid.UUID, sortField *string, direction *string) ([]*domain.Image, error)
  }
  ```
  `ShareFolderRepository` and `ShareImageRepository` are satisfied implicitly by the existing `folderRepository` and `imageRepository` structs (`repository/folder_repository.go`, `repository/image_repository.go`) — no changes to those files beyond passing the same instances into `NewShareUsecase`.
- `backend/internal/repository/folder_share_repository.go` — GORM implementation of `FolderShareRepository`
- `backend/internal/handler/share.go` — `ShareHandler` with `CreateShare`, `GetShare`, `DeleteShare` (protected) and `GetSharedFolder` (public)

### 4. Owner-facing endpoint semantics
- `POST /folders/:id/share` — ownership check via `ShareFolderRepository.GetByID(id, userID)` (404 if not found/not owned), then `FolderShareRepository.GetByFolderID`; if present return it, else generate a token and `Create`. Returns `{ "token": "..." }`, 200 (existing) or 201 (newly created) — handler returns 201 only when it actually created a row, 200 if returning an existing one.
- `GET /folders/:id/share` — same ownership check, then `GetByFolderID`; 404 if no share row exists.
- `DELETE /folders/:id/share` — same ownership check, then `DeleteByFolderID`; 204 regardless of whether a row existed (delete is idempotent).

### 5. Public endpoint (`GET /share/:token`)
- Registered on `e` directly (outside `protected`), still behind the global CORS/recovery middleware in `initEcho`.
- `ShareUsecase.GetSharedFolder(ctx, token)`:
  1. `FolderShareRepository.GetByToken` → 404 if not found (unknown/revoked token)
  2. `ShareImageRepository.ListByFolder(ctx, share.Folder.UserID, share.FolderID, nil, nil)` — `nil, nil` sort params reuse the existing default ordering by `image_folders.position`, same as the owner's folder view
  3. For each image, generate `thumbnail_url` (from `ThumbnailPath`, may be `nil`) and `full_res_url` (from `R2Path`) via `StorageService.GeneratePresignedGetURL` with the existing `presignedGetTTL` (24h) — same helper pattern as `imageUsecase.thumbnailURL`
- Response:
  ```json
  {
    "folder": { "name": "...", "notes": "..." },
    "images": [
      { "title": "...", "thumbnail_url": "...|null", "full_res_url": "..." }
    ]
  }
  ```
  `notes` maps to `Folder.Description` (nullable → `null` if unset).

## Risks / Trade-offs

- **Presigned URL TTL vs. "stable" link** → The share *link* (`/share/:token`) never expires, but `full_res_url`/`thumbnail_url` values expire after 24h. Mitigated by always generating fresh presigned URLs on each `GET /share/:token` request — acceptable since the public view is expected to be loaded fresh, not cached long-term by the client.
- **Concurrent first-share requests** → Two simultaneous `POST /folders/:id/share` for the same folder could both pass the `GetByFolderID` check and attempt `Create`, tripping the `folder_id UNIQUE` constraint on the second. Mitigated by catching the unique-violation in the repository and falling back to `GetByFolderID` (return the row the other request just created).
- **Exposing folder notes/name publicly once shared** → This is the intended behavior per the proposal, but worth the owner explicitly opting in via the share action (no auto-share). No redaction needed.
- **Token enumeration** → 128 bits of entropy from `crypto/rand` makes brute-forcing a valid token infeasible; no rate-limiting is added in this change since it doesn't exist for any other endpoint either.

## Migration Plan

Purely additive: new table, new routes, no changes to existing schemas or response shapes. Standard migration up/down (`000015_create_folder_shares`). Rollback is dropping the table — no data migration/backfill involved.
