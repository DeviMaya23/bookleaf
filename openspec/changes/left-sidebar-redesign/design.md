## Context

The current sidebar has a single system entry ("Unsorted") and a flat user folder list. The backend already supports soft-delete (trash), folder `parent_id` nesting, and a `GET /images` all-images endpoint. The frontend has never surfaced these. This change wires them up and updates the sidebar to match the new design.

## Goals / Non-Goals

**Goals:**
- Surface All, Unsorted, and Trash as first-class navigation entries
- Render the user folder list as a recursive tree using `parent_id`
- Enable subfolder creation from the context menu
- Implement the Trash view with oldest-deleted-first ordering and per-image Restore
- Fix the `ListTrashed` cursor so it sorts by `deleted_at` instead of `created_at`

**Non-Goals:**
- Folder counts in the sidebar (no backend support yet)
- Drag-and-drop folder reordering
- Moving images between folders from the sidebar

## Decisions

### 1. View discriminator type instead of `folderId: string | null`

**Decision**: Introduce a discriminated union for the active view:
```ts
type AppView =
  | { type: 'all' }
  | { type: 'unsorted' }
  | { type: 'trash' }
  | { type: 'folder'; id: string }
```

**Rationale**: The current `folderId: string | null` overloads `null` to mean "Unsorted". With three system views (All, Unsorted, Trash) and folder views, a discriminated union is unambiguous and each view branch can select the right query key and API call without `if/else` chains.

**Alternative considered**: String literals (`'all' | 'unsorted' | 'trash' | string`) — rejected because it doesn't enforce exhaustive handling at the type level.

---

### 2. Trash cursor encodes `deleted_at` instead of `created_at`

**Decision**: Add an optional `DeletedAt *time.Time` field to `ImageCursor`. The existing `GET /images` cursor continues to use `CreatedAt`. The trash path populates `DeletedAt` and uses `(deleted_at, id) > (?, ?)` for ASC pagination.

**Rationale**: Trash is sorted `deleted_at ASC` (oldest deleted first). The cursor must carry the same field used in the sort key. Adding an optional field to the shared struct is the smallest change — no new type, no new encode/decode functions, and existing cursors remain valid for the non-trash path.

**Alternative considered**: Separate `TrashCursor` type and encode/decode pair — cleaner isolation but doubles the pagination boilerplate with no benefit since the two paths are never mixed.

---

### 3. Tree built client-side from flat `parent_id` list

**Decision**: `GET /folders` returns a flat list with `parent_id`. The sidebar builds a tree in a utility function before rendering:
```
buildFolderTree(folders: Folder[]): FolderNode[]
```
The recursive `FolderItem` component renders each node at depth `d` with `paddingLeft: 8 + d * 14`.

**Rationale**: The API already returns `parent_id` and the data set is small (user's personal folders). A client-side tree build is a simple O(n) map pass with no backend changes.

**Alternative considered**: Add a `GET /folders/tree` endpoint returning pre-nested JSON — unnecessary overhead; the flat list is sufficient.

---

### 4. `/` becomes the All-images view; `/unsorted` is a new explicit route

**Decision**:
```
/           → All images (GET /images, no filter params)
/unsorted   → Unfiled images (GET /images?unfiled=true)
/trash      → Trashed images (GET /images/trash)
/folders/:id → Folder images (GET /images?folder_id=...)
```

**Rationale**: "/" is the natural default for "see everything". The backend already returns all non-deleted images when `GET /images` is called with no filter — no backend change needed. The existing `/folders/:folderId` routes are unaffected.

---

### 5. `ImageGrid` accepts `AppView` instead of `folderId`

**Decision**: `ImageGrid` receives the `AppView` discriminated union. Internally it selects the query key and fetch function by `view.type`. Trash mode swaps the context menu item from "Delete" to "Restore".

**Rationale**: Centralises view-specific logic in one component rather than splitting it across `AppLayout`, `TrashGrid`, and `ImageGrid`. The trash path is similar enough to the normal path that a separate component would duplicate most of the infinite query + masonry grid code.

## Risks / Trade-offs

- **Cursor encoding change for trash**: Existing in-flight trash cursors (e.g., open tabs with paginated trash) will break on deploy since the cursor structure changes. Risk is low — trash pagination is fresh and rarely held across deploys — but worth noting.
- **Deep nesting UX**: At depth 3+, `paddingLeft: 8 + depth * 14` starts consuming significant horizontal space in a 240px sidebar. No hard cap is enforced; very deep trees will look cramped. Acceptable for now given nesting is a new feature and real usage will inform if a cap is needed.
- **`parent_id` loop guard**: The tree builder must guard against cycles in `parent_id` references (corrupt data). A visited-set check during tree construction prevents infinite recursion.
