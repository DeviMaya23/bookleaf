## Context

The backend follows clean architecture (`handler → usecase → repository → domain`). The existing single-image download (`image-download` capability) returns a presigned R2 URL — the backend never touches image bytes. Folder export is different: the backend itself must read every image's bytes from R2 and assemble them into a zip, then deliver that zip as the HTTP response body.

Relevant existing pieces:
- `usecase.ImageRepository.ListByFolder(ctx, userID, folderID, sortField, direction)` — already returns all non-deleted images in a folder, ordered by position. No cursor/limit, exactly what's needed for a single-level export.
- `usecase.StorageService.GetObject(ctx, key) (io.ReadCloser, error)` — already used elsewhere (e.g. trash) for reading R2 objects directly.
- `usecase.downloadFileExtension(mimeType string) string` — existing MIME→extension mapping used by single-image download.
- `folderUsecase` currently depends on `ImageCounter` (an interface with only `CountByFolderID`), not the full `ImageRepository`, and has no `StorageService` dependency.
- `FolderHandler.GetFolder` already performs the ownership check (folder exists and belongs to `userID`) and returns `404` otherwise — this check needs to happen *before* any zip bytes are written, since headers can't be changed once streaming starts.

## Goals / Non-Goals

**Goals:**
- Stream a zip of a folder's direct images (no subfolder recursion) to the HTTP response with constant memory, reusing existing R2 and repository methods.
- Keep the zip-building logic in the usecase layer (business logic: which images, what they're named, dedup), while the handler owns HTTP framing (status, headers, `Content-Disposition`).
- Reuse `downloadFileExtension` and `ListByFolder` rather than introducing parallel logic.

**Non-Goals:**
- Recursing into child folders.
- Background/async export, temporary R2 storage for zips, or job-status tracking (future "export everything" work).
- Retrying or recovering from a mid-stream R2 failure.
- Progress reporting to the frontend beyond a single "preparing" state.

## Decisions

### 1. Usecase streams into an `io.Writer`; handler owns HTTP framing

`FolderUsecase` gains:
```go
ExportFolder(ctx context.Context, folderID uuid.UUID, userID string, w io.Writer) error
```
It calls `imageRepo.ListByFolder(...)`, then for each image: `store.GetObject`, create a `zip.Writer` entry with a deduped name, `io.Copy`, close the reader. `zip.NewWriter(w)` / `.Close()` brackets the loop.

The handler:
1. Calls `folderUsecase.GetByID(ctx, id, userID)` first (existing method) — returns `404` if the folder doesn't exist or isn't owned, *before* anything is written.
2. On success, sets `Content-Type: application/zip` and `Content-Disposition: attachment; filename="<sanitized folder name>.zip"`, writes `200 OK`.
3. Calls `folderUsecase.ExportFolder(ctx, id, userID, c.Response())`.
4. If `ExportFolder` returns an error after bytes have been written, logs it server-side — the response is already underway and can't be converted to an error status (this is the "truncated response = failed download" behavior agreed in the proposal).

**Alternative considered**: handler does the R2/zip work directly, usecase just returns the image list. Rejected — it would put storage-fetching and filename/dedup logic (business rules) in the handler layer, breaking the existing dependency direction where handlers stay thin.

### 2. Extend `folderUsecase`'s dependencies

Two additions to `folderUsecase`'s constructor dependencies, both following patterns already used by `imageUsecase`:

- Add `ListByFolder` to the image-repo-subset interface `folderUsecase` depends on. That interface is currently named `ImageCounter` (just `CountByFolderID`); rename it to `FolderImageRepository` and add `ListByFolder` to it. The concrete `imageRepository` already implements both methods, so no repository changes — only the interface definition and `NewFolderUsecase` call sites change.
- Add `store usecase.StorageService` as a new constructor parameter (same interface `imageUsecase` already uses for `GetObject`). Wire the existing `r2Storage` instance into `NewFolderUsecase(...)` in `main.go`, same as it's wired into the image usecase.

**Alternative considered**: introduce a separate `FolderExportUsecase` type to avoid touching `folderUsecase`'s constructor. Rejected as an unnecessary new abstraction for one method — `folderUsecase` already owns folder-scoped operations, and the new dependencies are interfaces already used elsewhere in the codebase, not new external dependencies.

### 3. Filename handling (two distinct sanitization needs)

- **Archive filename** (`Content-Disposition`): sanitize the folder's `Name` for use as a filename — strip/replace characters invalid in filenames (e.g. `/`, `\`, control characters), trim whitespace. If the result is empty, fall back to a default like `export`. Append `.zip`. This lives in the handler (it's purely an HTTP header concern), as a small helper alongside the existing handler code.
- **Per-entry filenames** (inside the zip): `<image.Title>.<ext>` via `downloadFileExtension(image.MIMEType)`, same as single-image download. Image titles are free-text and may contain `/`, which `archive/zip` would otherwise interpret as a directory separator inside the archive. The usecase sanitizes titles the same way (replace path separators) before use. Duplicate resulting names (including post-sanitization collisions) are disambiguated by appending ` (1)`, ` (2)`, etc., tracked via a `map[string]int` for the duration of the export.

### 4. No flushing/progress signaling

The frontend uses `response.blob()`, which doesn't resolve until the response completes — so incremental flushing wouldn't improve perceived UX for v1. Given "not that big" folder sizes (per product decision), the default buffering behavior of Echo's response writer is sufficient. Not adding explicit `http.Flusher` calls keeps the implementation minimal; can be revisited if folder sizes grow.

## Risks / Trade-offs

- **Mid-stream R2 failure produces a truncated zip after `200 OK` has been sent** → Accepted per proposal ("easiest" = no special handling); usecase logs the error server-side via existing telemetry so failures are diagnosable even though the client just sees a failed download.
- **Image titles containing `/` could otherwise create unintended nested paths inside the zip** → Mitigated by sanitizing titles before using them as zip entry names (Decision 3).
- **Folder name sanitization could reduce to an empty string** (e.g. a name made entirely of characters invalid in filenames) → Mitigated by a fallback default (`export.zip`).
- **Browser buffers the full zip in memory via `blob()`** → Acceptable for current "not that big" folder sizes; if usage grows, this is the first thing that would need to change (and would likely coincide with moving to the async "export everything" design, which is explicitly out of scope here).
- **Renaming `ImageCounter` → `FolderImageRepository`** touches its one existing call site (`NewFolderUsecase` construction in `main.go` and any test doubles) → small, mechanical change; no behavioral risk.

## Migration Plan

No database migration required. This is an additive change:
- New route `GET /folders/:id/export` registered alongside existing `/folders` routes.
- New frontend button is additive to `FolderPanelContent`.
- Rollback is a standard revert of the route registration and UI change — no data to back out.
