## Why

Soft-deleted images accumulate in the database indefinitely with no automatic cleanup — their DB records and R2 objects (full image + thumbnail) are never removed. This leaks storage and leaves user data sitting past any reasonable retention window. A 30-day trash retention policy (consistent with common consumer cloud storage) gives users time to restore accidental deletions while bounding data and storage growth.

## What Changes

- New `PurgeExpiredTrash` method on `ImageUsecase` that hard-deletes images trashed more than 30 days ago — deletes both the R2 object and thumbnail (best-effort), then hard-deletes the DB record
- New `ListExpiredTrash` and `HardDelete` methods on `ImageRepository` interface and implementation
- Background goroutine in `cmd/server/main.go` running on a 24-hour ticker, calling `PurgeExpiredTrash` with a 30-day threshold
- No DB schema changes — `deleted_at` already exists on the `images` table

## Capabilities

### New Capabilities

- `trash-purge`: Periodic hard-deletion of images that have been in the trash for more than 30 days, including cleanup of their R2 objects

### Modified Capabilities

(none — no existing spec-level requirements are changing)

## Impact

- **Database**: Hard deletes via `Unscoped()` on `images`; records are permanently removed
- **Storage (R2)**: `r2_path` and `thumbnail_path` (when present) are deleted for each purged image; reuses `DeleteObject` introduced in `stale-upload-cleanup`
- **Startup**: Second goroutine added to `cmd/server/main.go` alongside the existing stale-upload cleanup goroutine
- **No API changes**: Purge is entirely background; no new endpoints
