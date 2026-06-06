## Context

The backend exposes two permanent deletion endpoints (`DELETE /images/trash/:id` and `DELETE /images/trash`) that are fully implemented but have no frontend wiring. The frontend trash view currently only supports restoring images. This change adds permanent deletion to two surfaces: the image card context menu (single delete) and the Trash sidebar entry (bulk delete).

## Goals / Non-Goals

**Goals:**
- Wire single permanent delete via image card context menu in trash view
- Wire empty trash via Trash sidebar context menu
- Show appropriate confirmation dialogs for each action

**Non-Goals:**
- Multi-select permanent delete (not in scope)
- Trash item count badge in sidebar
- Any backend changes

## Decisions

### `ImageCard` gets a second action prop for permanent delete

Currently `onAction` is dual-purpose (restore or soft-delete depending on context). Rather than extend it further, a second `onDeletePermanent?: (image: Image) => void` prop is added to `ImageCard`. This keeps each prop's intent unambiguous and lets the context menu render both items independently. The separator comes from `ContextMenuSeparator`, which is already exported from `ui/context-menu`.

### Confirmation dialogs live in the component that owns the action

- Single delete dialog lives in `ImageGrid` (as local state `confirmDeleteImage: Image | null`)
- Bulk delete dialog lives in `FolderSidebar` (as local state `confirmEmptyTrash: boolean`)

Both use the existing `Dialog` / `DialogContent` / `DialogFooter` pattern from `@/components/ui/dialog`, already imported in `FolderSidebar` and importable in `ImageGrid`.

### Trash sidebar entry is wrapped in `ContextMenu`

The Trash `div` in `FolderSidebar` is wrapped with `ContextMenu` + `ContextMenuTrigger` + `ContextMenuContent`, consistent with how `FolderItem` handles its context menu. No structural change to the sidebar layout is needed.

### Query invalidation strategy

- **Single delete**: Optimistically remove the image from `orderedImages` state (same pattern as soft-delete), then invalidate `['images', 'trash']`.
- **Empty trash**: Invalidate `['images', 'trash']` after success. The trash view re-fetches and renders the existing empty state. No navigation away.

## Risks / Trade-offs

- **Empty trash no-op**: If the user opens "Empty trash" when trash is already empty, the confirmation dialog still appears. On confirm, the API returns 204 with a no-op and the query invalidates — the empty state is already showing, so there is no visible change. This is acceptable; the confirmation dialog is the real guard.
