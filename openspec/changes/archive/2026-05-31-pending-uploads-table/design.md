## Context

The two-step upload flow (InitiateUpload → client uploads to R2 → CompleteUpload) currently stores in-progress upload records in the `images` table with `is_uploaded = false`. This creates an invisible invariant: all user-facing queries must filter `is_uploaded = true` or they silently expose uncommitted records. The invariant is not enforced by the ORM and is already inconsistently applied (`GetByID`, `CountByFolderID`, and `image_folders` rows all see limbo records). A pending upload is not an image domain object; the two share a table as a shortcut.

## Goals / Non-Goals

**Goals:**
- Make `images` a clean, always-valid table — no filter needed to exclude uncommitted records
- Eliminate phantom rows in `image_folders` pointing to uncommitted images (folder assignment moves to CompleteUpload)
- Give the stale cleaner a dedicated table to operate on, decoupled from the image read path
- Preserve the public HTTP API contract entirely (no handler changes)

**Non-Goals:**
- Resumable or chunked uploads
- Per-upload status tracking beyond pending / committed / stale
- Changes to the R2 storage layout or presigned URL TTL

## Decisions

### `pending_uploads` as a separate table, not a status column

The alternative — replacing the boolean with a status enum (`pending`, `uploaded`) — carries the same flaw: every query must opt into filtering. Only a separate table makes the correctness invariant structural rather than conventional. The operational cost (a transaction in `CompleteUpload` instead of a plain UPDATE) is negligible at current scale.

### `pending_uploads.id` becomes `images.id` on commit

The UUID generated at `InitiateUpload` time is reused as the committed image's ID. This avoids exposing two different IDs to the client across the two calls and keeps the presigned URL path stable (`users/{userID}/images/{id}.ext` is computed at initiation and stored in `pending_uploads.r2_path`).

### Folder assignment moves from `InitiateUpload` to `CompleteUpload`

Currently `SetImageFolder` is called at `InitiateUpload`, creating `image_folders` rows that reference uncommitted images. Moving the call inside the `CompleteUpload` transaction means `image_folders` rows only ever reference committed images. The `folder_id` is stored in `pending_uploads` at initiation and used during the commit transaction.

### `CompleteUpload` transaction: INSERT images → SetImageFolder → DELETE pending_uploads

All three steps run in a single DB transaction. If any step fails, the pending row survives and the stale cleaner eventually removes it along with its R2 object. This matches the existing failure semantics without requiring a saga or compensating transaction.

### `PendingUploadRepository` injected into `imageUsecase`

Rather than creating a separate `pendingUploadUsecase`, the new repository is wired as a second repository dependency of the existing `imageUsecase`. The usecase already orchestrates `imageRepo` and `folderRepo`; adding `pendingUploadRepo` follows the same pattern and keeps `CleanupStaleUploads` co-located with the image lifecycle logic.

### Migration drops limbo rows, not migrates them

At deploy time, any `images` rows with `is_uploaded = false` are hard-deleted by the migration before the column is dropped. In-flight uploads at deploy time will fail at `CompleteUpload` (image not found) and the client will see an error. This is acceptable given the 30-minute stale window and the expectation of a short maintenance window during deploy. Migrating limbo rows to `pending_uploads` would add migration complexity for negligible benefit.

## Risks / Trade-offs

- **In-flight uploads fail at deploy** → Acceptable; stale uploads already fail silently after 30 min. Document as a known deploy behaviour.
- **Transaction in `CompleteUpload` is heavier than a plain UPDATE** → Negligible at current scale; thumbnail generation dominates CompleteUpload latency by orders of magnitude.
- **`pending_uploads` rows accumulate if server crashes mid-transaction** → Same as today; stale cleaner handles them within 30 minutes.
- **`pending_uploads.folder_id` references `folders(id)`** → If a folder is deleted between `InitiateUpload` and `CompleteUpload`, the FK would reject the insert. Use `ON DELETE SET NULL` on the FK so the commit proceeds and the image lands as unfiled (mirrors current folder-not-found fallback in `InitiateUpload`).

## Migration Plan

Migration `000011` runs in order:

1. `DELETE FROM images WHERE is_uploaded = false` — purge limbo rows before structural changes
2. `ALTER TABLE images DROP COLUMN is_uploaded` — remove the flag
3. `CREATE TABLE pending_uploads (...)` — new table with all initiation-time fields

Down migration reverses:
1. `DROP TABLE pending_uploads`
2. `ALTER TABLE images ADD COLUMN is_uploaded BOOLEAN NOT NULL DEFAULT true` — default true so existing committed rows are treated as uploaded
