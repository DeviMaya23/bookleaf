## Context

The app currently has no drag-and-drop capability. Image cards in `ImageGrid` and folder items in `FolderSidebar` are static. Reorganizing requires context menus. The upload modal has a basic drop zone with no preview and no metadata fields beyond title.

The backend already supports all necessary mutations: `PATCH /images/:id` accepts `folder_id`, `PUT /folders/:id` accepts `parent_id`. No backend changes are needed.

## Goals / Non-Goals

**Goals:**
- Move images between folders (including clearing folder assignment) via drag-and-drop
- Nest and promote folders via drag-and-drop within the sidebar
- Auto-upload files dragged from the OS onto the main content area
- Redesign the upload modal with thumbnail preview and collapsible metadata fields

**Non-Goals:**
- Reordering images within a folder (no sort order field exists on the backend)
- Multi-select drag (one item at a time)
- Drag-and-drop on mobile / touch devices (no touch sensor configured)
- Drag images to Trash (destructive action stays in context menu only)

## Decisions

### D1: @dnd-kit/core for UI drag-and-drop

**Chosen**: `@dnd-kit/core`  
**Alternatives considered**: `react-dnd` (uses legacy context API, in maintenance mode), native HTML5 drag-and-drop (feasible but brittle cross-browser, no accessible keyboard support)  
**Rationale**: dnd-kit is the current standard, tree-shakable, no external dependencies, has a `DragOverlay` portal for smooth floating previews, and supports keyboard accessibility out of the box.

### D2: DndContext lives in AppLayout

`DndContext` wraps all of `AppLayout`. This is required because drag sources (image cards in `ImageGrid`) and drop targets (folder items in `FolderSidebar`) are siblings — they cannot share context if it is scoped to either component alone. `onDragEnd` in `AppLayout` dispatches to the correct mutation depending on whether the dragged item is an image or a folder.

```
AppLayout
└── DndContext (onDragEnd)
    ├── FolderSidebar  ← drop targets + folder drag sources
    ├── ImageGrid      ← image drag sources
    └── DragOverlay    ← floating preview portal
```

### D3: OS file drops use native browser events, not dnd-kit

File drags originating from the OS have no dnd-kit draggable source, so dnd-kit never activates. A `dragover`/`drop` listener on the `<main>` element handles these. Gate on `event.dataTransfer.types.includes('Files')` to avoid interfering with dnd-kit internal drags.

A full-page overlay (rendered inside `<main>`) appears during a valid OS file drag, providing a clear drop target.

### D4: Drag item data schema

Each draggable carries a typed `data` object so `onDragEnd` can route correctly:

```ts
// image card
{ type: 'image', imageId: string, currentFolderId: string | null }

// folder item
{ type: 'folder', folderId: string, name: string, parentId: string | null }
```

Each droppable carries:
```ts
// sidebar folder item
{ type: 'folder', folderId: string }

// "Unsorted" system entry
{ type: 'unsorted' }

// root drop zone (empty space below folder list)
{ type: 'root' }
```

### D5: Circular folder drop prevention

Before calling `moveFolder`, walk the flat folder list to collect the full subtree of the dragged folder. If the drop target's `folderId` appears in that subtree (or equals the dragged folder itself), abort silently. This check is O(n) on the folder list, acceptable for typical folder counts.

### D6: Auto-upload opens right panel via getImage

`completeUpload` returns only `image_id`. To open the right panel, call `getImage(getToken, image_id)` after completing the upload to obtain the full `Image` object, then pass it to `setSelectedImage`. The title field auto-focus is handled by a `useEffect` in `RightPanel` that runs when `image.id` changes — it fires on the ref of the title input.

### D7: Upload modal thumbnail preview uses object URLs

When a file is selected in the modal, generate a preview with `URL.createObjectURL(file)`. Store the URL in a ref. Revoke it in the modal's `handleClose` function and whenever a new file replaces the previous one. This avoids memory leaks without needing a `useEffect` cleanup cycle tied to the file state.

### D8: moveFolder in folders.ts

`PUT /folders/:id` requires `name` (the field is non-nullable on the backend). `moveFolder` therefore accepts the current folder name alongside the new `parentId`:

```ts
moveFolder(getToken, id: string, name: string, parentId: string | null): Promise<Folder>
```

The caller (FolderSidebar's `onDragEnd` handler) already has the name from the drag item's data payload.

## Risks / Trade-offs

- **Stale folder tree during drag**: If the folder list re-fetches mid-drag, the drag operation can reference a stale ID. Mitigation: `staleTime` on the folders query is already 60s; folder mutations invalidate immediately after drop, not during.
- **dnd-kit and ScrollArea**: The gallery is inside a shadcn `ScrollArea` (custom scrollbar). dnd-kit's pointer sensor may need `activationConstraint` (e.g. 8px distance) to avoid accidental drags on scroll. Set `{ distance: 8 }` on `PointerSensor`.
- **Object URL leak on unmount**: If the component unmounts while a file is staged (e.g. navigating away), the object URL is not revoked. Low-severity memory leak; acceptable for a modal that is conditionally rendered.
