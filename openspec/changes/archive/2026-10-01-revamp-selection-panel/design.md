## Context

The selection panel currently uses a flat search-list `SelectionFolderPicker` component, has no download capability, and does not expose the already-wired `onExitSelectMode` callback. The chip-style `TokenInput`/`FolderInput` components exist in the right panel and just need to be reused. A folder export endpoint already streams a ZIP from R2 using `archive/zip`; the bulk export follows the same pattern.

## Goals / Non-Goals

**Goals:**
- Revamp `SelectionPanelBody`: chip folder picker (multi-folder, deferred Apply), download ZIP button, exit (×) button in the header
- Add `POST /images/bulk/export` that streams a ZIP of images by ID list, reusing the existing zip-writing logic from `ExportFolder`
- No changes to selection mechanics, toolbar, or `AppLayout` state

**Non-Goals:**
- Bulk tag add (separate proposal)
- Client-side ZIP fallback
- Removing images from folders in bulk
- Any change to `GET /folders/:id/export` or `bulkAddImagesToFolder` behavior

## Decisions

### 1. Shared zip-writing helper, not duplicated logic

The core zip loop (`GetObject` → `zip.Create` → `io.Copy`) currently lives inline in `folderUsecase.ExportFolder`. Rather than duplicating it into `imageUsecase`, extract it as a package-level function in the `usecase` package:

```go
// writeImagesToZip writes a zip archive of the given images to w.
func writeImagesToZip(ctx context.Context, images []*domain.Image, store StorageService, w io.Writer) error
```

Both `folderUsecase.ExportFolder` and `imageUsecase.BulkExportImages` call this helper. This keeps the naming/deduplication logic in one place.

**Alternative considered:** Put the helper on `folderUsecase` and call it from `imageUsecase`. Rejected — cross-usecase calls violate the single-responsibility boundary; a package-level function is the idiomatic Go pattern for shared behavior within a package.

### 2. New repository method: GetManyByIDs

`BulkExportImages` needs to fetch full `*domain.Image` records (with R2Path, Title, MIMEType) for a list of IDs. The existing `FilterOwnedImageIDs` only returns IDs. A new method is added to `ImageRepository`:

```go
// GetManyByIDs returns non-deleted images matching the given IDs that belong to userID.
// Results include Tags and ImageFolders preloaded. Order is not guaranteed.
GetManyByIDs(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) ([]*domain.Image, error)
```

Implemented as a single `WHERE id IN (...) AND user_id = ? AND deleted_at IS NULL` query with preloads — no N+1. IDs not owned by the user or not found are silently dropped (same tolerance as `bulkAddImagesToFolder`).

**Alternative considered:** Loop `GetByID` per image. Rejected — N+1 queries; unacceptable for even modest selections.

### 3. Deferred Apply for folder chips

Folder chips in the selection panel are local UI state. The Apply button fires one `bulkAddImagesToFolder` call per selected folder. Removing a chip before Apply fires no API call.

**Alternative considered:** Immediate fire on chip add (matching image detail panel autosave). Rejected — no `bulkRemoveFromFolder` endpoint exists, so undo would require a new endpoint. The Apply pattern is also clearer for bulk operations and naturally enables multi-folder add.

### 4. POST /images/bulk/export streams directly, no presigned URL

The handler sets `Content-Type: application/zip` and `Content-Disposition: attachment; filename="bookleaf-export.zip"`, writes `200 OK`, then streams. Same pattern as `GET /folders/:id/export`. The FE receives the response as a blob, creates an `<a>` element, and triggers download — identical to `FolderPanelContent.handleExport`.

**Alternative considered:** Return a presigned URL to a pre-generated ZIP in R2. Rejected — requires temporary storage management, adds latency before download starts, and introduces cleanup concerns. Direct streaming is simpler and already proven by folder export.

### 5. Exit button placement: inline with selected count

The `×` button sits inline with "N selected" text in the header row (flex row, `×` to the right of the count). The chevron is `absolute top-2 right-2` and at panel width 320px, the text + button occupies the left ~120px at most — no collision. `SelectionPanelBody` receives `onExitSelectMode` via a new prop (already in `PanelContent.selection`; `RightPanel.deriveContent` just wasn't forwarding it).

## Risks / Trade-offs

- **Large selections slow ZIP streaming** → The ZIP is streamed synchronously from R2 to the client, one image at a time. For very large selections (100+ images) this could be slow and tie up a server goroutine. Acceptable for current user scale; async job approach is the mitigation path if needed later.
- **Partial ZIP on mid-stream error** → If `GetObject` fails mid-archive, bytes already written remain on the wire. The client will receive a truncated ZIP. This is identical behavior to folder export — accepted trade-off of direct streaming.
- **Apply exits select mode** → After the Apply button fires `bulkAddImagesToFolder` successfully, select mode exits (matching current folder-pick behavior). Users who want to add to multiple folders must re-enter select mode. This is a UX simplification; the alternative (stay in select mode after Apply) adds complexity with little gain.
