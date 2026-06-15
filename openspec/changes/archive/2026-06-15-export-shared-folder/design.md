## Context

`GET /folders/:id/export` already streams a folder's images as a zip via `folderUsecase.ExportFolder(ctx, folderID, userID, w)`. That method's only authorization-relevant input is `userID`, which it passes straight through to `imageRepo.ListByFolder(ctx, userID, folderID, ...)` — it performs no separate ownership check itself (the handler's prior `GetByID` call is what enforces ownership for the authenticated route).

`GetSharedFolder(ctx, token)` already does the equivalent thing for the public JSON view: it resolves `share.Folder.UserID` via `FolderShareRepository.GetByToken`, then calls `ListByFolder(ctx, share.Folder.UserID, share.FolderID, ...)`. The share token is the authorization gate — anyone holding it can already see every image (titles, thumbnails, full-res presigned URLs) in that folder.

This means `folderUsecase.ExportFolder(ctx, share.FolderID, share.Folder.UserID, w)` produces exactly the same zip a folder owner would get from `/folders/:id/export`, with no additional data exposed beyond what `/share/:token` already returns.

## Goals / Non-Goals

**Goals:**
- Let anyone with a valid share token download all of a shared folder's images as a single zip, via `GET /share/:token/export`.
- Reuse `folderUsecase.ExportFolder` as-is — no duplication of the zip-building loop, entry naming, or sanitization logic.
- Mirror the existing `/folders/:id/export` handler shape (lookup for filename, then stream) for consistency.

**Non-Goals:**
- No changes to `/share/:token` (JSON listing) or `/folders/:id/export` behavior.
- No per-image download endpoint or `Content-Disposition` changes to presigned URLs — out of scope per the proposal (single-image download already works via right-click on the existing full-res URL).
- No new revocation/expiry semantics — the export endpoint is gated by the same token as the rest of the share, so `DELETE /folders/:id/share` already revokes it.

## Decisions

### Reuse `folderUsecase.ExportFolder` directly (Option B), via a narrow handler-local interface

`ShareHandler` gains a `FolderExporter` interface:

```go
type FolderExporter interface {
    ExportFolder(ctx context.Context, folderID uuid.UUID, userID string, w io.Writer) error
}
```

satisfied implicitly by `*folderUsecase`, and injected via `NewShareHandler(shareUsecase, folderExporter, tel)`.

**Alternatives considered:**
- *Extract a shared zip-writing helper inside `package usecase`* (Option A from exploration) — would avoid the handler depending on a second usecase, but duplicates the "list images, scope by owner ID" step that `GetSharedFolder` already demonstrates is safe to do with `share.Folder.UserID`. Option B reuses the entire existing, already-tested `ExportFolder` method verbatim.
- *Have `shareUsecase` itself depend on `folderUsecase.ExportFolder`* — pushes the new cross-usecase dependency down a layer instead of into the handler. Rejected because `ShareHandler` already orchestrates two pieces of data (folder info for the filename, then the export call); keeping the dependency at the handler level matches how `FolderHandler.ExportFolder` itself is structured (handler calls usecase twice — `GetByID` then `ExportFolder`).

This is a new cross-handler/cross-usecase dependency (`ShareHandler` → `folderUsecase`'s export capability) that doesn't exist today, flagged here per the project's decision-boundary process and accepted by the user during exploration.

### New `GetSharedFolderInfo` usecase method for the filename lookup

`FolderHandler.ExportFolder` calls `GetByID` first to get `folder.Name` for `Content-Disposition`, then calls `ExportFolder` separately (two calls, same pattern as the rest of the codebase). `ShareHandler.ExportSharedFolder` follows the same two-call shape:

1. `shareUsecase.GetSharedFolderInfo(ctx, token)` → `*domain.Folder` (has `ID`, `Name`, `UserID`) — wraps `FolderShareRepository.GetByToken`, returns `gorm.ErrRecordNotFound` for unknown/revoked tokens (→ `404`).
2. Set `Content-Type: application/zip` and `Content-Disposition: attachment; filename="<sanitized folder name>.zip"` (reuse existing `sanitizeFilename` helper from `handler/folder.go`), write `200 OK`.
3. `folderExporter.ExportFolder(ctx, folder.ID, folder.UserID, c.Response())`.

**Alternative considered:** reuse `GetSharedFolder` (which already calls `GetByToken`) for the filename — rejected because it also generates presigned URLs for every image, which is wasted work for a request that's about to stream a zip.

## Risks / Trade-offs

- **[Risk]** `GetByToken` is called twice per export request (once via `GetSharedFolderInfo`, once inside... no — `ExportFolder` doesn't call `GetByToken` at all, it takes `folderID`/`userID` directly, so this is just one extra `GetByToken` compared to a hypothetical single-call design) → Acceptable; matches the existing two-call cost of `/folders/:id/export` (`GetByID` + `ExportFolder`'s `ListByFolder`).
- **[Risk]** A revoked share token could theoretically still succeed if `GetSharedFolderInfo` and the export race with a `DELETE /folders/:id/share` — but this is the same TOCTOU window the JSON endpoint already has, and is not worsened here.
- **[Trade-off]** `ShareHandler` now depends on two usecases instead of one, increasing its construction surface — mitigated by the narrow `FolderExporter` interface limiting the visible contract to a single method.

## Migration Plan

No data migration. Purely additive: one new route, one new usecase method, one new handler method, one new constructor parameter. Existing routes and behavior unchanged. Deployable and revertible as a single normal release.
