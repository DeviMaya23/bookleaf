## Context

`InitiateUpload` in the usecase layer already accepts a `*uuid.UUID` for `folderID` and passes it directly to `imageRepo.Create`. The `FolderRepository` interface already exposes `GetByID(ctx, id, userID)` which returns `nil, gorm.ErrRecordNotFound` when not found. No schema or API changes are required — this is purely a usecase-layer behaviour fix.

## Goals / Non-Goals

**Goals:**
- Validate `folder_id` against the authenticated user's folders in `InitiateUpload`
- Fall back to `null` silently when folder is not found, rather than erroring or storing a bad reference

**Non-Goals:**
- Returning an error or warning to the caller when the folder is not found
- Validating `folder_id` on any other endpoint (UpdateImage already handles this via explicit update semantics)

## Decisions

**Validate in the usecase, not the handler.**
The handler already binds and passes the UUID as-is; keeping validation in the usecase keeps HTTP concerns out and makes it testable in isolation without an HTTP layer.

**Silently null-out an unfound folder rather than returning 4xx.**
The user requested this behaviour explicitly. It keeps the upload flow uninterrupted — a missing folder ID is treated as "no folder" rather than a blocking error. An alternative of returning 400 would break callers that send a stale folder ID (e.g. after a folder was deleted).

**Use `folderRepo.GetByID` (already on the interface).**
No new repository method needed. `GetByID` is already scoped to `userID`, so it handles both "doesn't exist" and "belongs to another user" with a single call.

## Risks / Trade-offs

- **Silent data loss risk**: A caller may not realise their `folder_id` was ignored. → Acceptable given the explicit product requirement; no warning is needed in the response.
- **Extra DB read per upload with a folder_id**: One additional SELECT on `folders` when `folder_id` is provided. → Negligible; folder table is small and indexed on `(id, user_id)`.

## Migration Plan

No migrations needed. The change is backward-compatible: existing images are unaffected, and the API contract (request/response shape) is unchanged.
