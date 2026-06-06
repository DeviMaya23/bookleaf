## Context

Trashed images are currently only permanently deleted by the `PurgeExpiredTrash` background worker after a 30-day retention window. The deletion sequence it performs — delete R2 object, delete thumbnail, hard-delete DB record — is the same sequence needed for user-initiated deletion. This change exposes that behaviour through two new authenticated endpoints, reusing existing infrastructure.

Existing relevant pieces:
- `ImageRepository.HardDelete(ctx, id, userID)` — scoped hard delete, already exists
- `ImageRepository.ListTrashed(ctx, userID, cursor, limit)` — lists soft-deleted images for a user
- `store.DeleteObject(ctx, path)` — R2 object deletion, used in `PurgeExpiredTrash`
- `PurgeExpiredTrash` in the usecase layer — reference implementation for the deletion sequence

## Goals / Non-Goals

**Goals:**
- Allow a user to permanently delete a single trashed image on demand
- Allow a user to permanently delete all of their trashed images at once
- Reuse the existing deletion sequence (R2 + thumbnail + DB) without duplicating logic

**Non-Goals:**
- Bulk delete by multiple explicit IDs (deferred; no FE multi-select yet)
- Frontend implementation
- Any change to the 30-day automatic purge behaviour

## Decisions

### Reuse `HardDelete` per item rather than a new bulk repo method

The `EmptyTrash` usecase fetches all trashed images via `ListTrashed` (unpaginated, no cursor) then loops — calling `store.DeleteObject` and `HardDelete` per item, matching the pattern in `PurgeExpiredTrash`. A single `DELETE WHERE user_id = ?` SQL shortcut was considered but rejected: it would skip R2 cleanup, leaving orphaned objects in storage.

### `DeleteFromTrash` validates the image is actually trashed before deleting

Before running the deletion sequence, the usecase calls the existing `GetDeletedByID(ctx, id, userID)` on `ImageRepository`, which queries only soft-deleted records. If the image is not found in trash — either it doesn't exist, belongs to another user, or isn't soft-deleted — a not-found error is returned. This prevents hard-deleting an image that is still active.

### Separate usecase methods, not extending `PurgeExpiredTrash`

`EmptyTrash` and `DeleteFromTrash` are new methods on `imageUsecase`, not a refactor of `PurgeExpiredTrash`. The worker flow (age-based, no user scope) and the user flow (on-demand, user-scoped) have different inputs and logging semantics. Sharing an implementation would couple them unnecessarily.

### New repo method: `ListAllTrashed(ctx, userID)`

`EmptyTrash` needs all trashed images for a user without pagination. The existing `ListTrashed` requires a cursor and limit. A new unpaginated variant avoids workarounds like setting an arbitrarily large limit.

### Routes

```
DELETE /images/trash        → handler.EmptyTrash
DELETE /images/trash/:id    → handler.DeleteFromTrash
```

Both sit under the existing `protected` group (JWT-authenticated). `/images/trash` already exists as a `GET` — adding a `DELETE` on the same path is additive and unambiguous.

## Risks / Trade-offs

- **Irreversibility**: Unlike soft delete, these operations permanently remove R2 objects and DB records. There is no undo. → The API contract must make this clear; no mitigation needed at the backend layer.
- **Empty-trash on large collections**: `ListAllTrashed` loads all trashed images into memory before looping. For users with hundreds of trashed images this is acceptable; for extreme cases it could be slow. → Acceptable for now; a streaming/batch approach can be introduced later if needed.
- **Partial failure on empty trash**: If R2 deletion fails for some items mid-loop, the loop continues (same best-effort policy as `PurgeExpiredTrash`). DB records are still hard-deleted. → Consistent with existing policy; logged at warn level.
