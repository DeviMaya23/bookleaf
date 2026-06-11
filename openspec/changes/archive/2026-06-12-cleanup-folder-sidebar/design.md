## Context

`FolderSidebar.tsx` (270 lines) already follows good extraction precedent
within its own feature (`FolderItem`, `FolderNameDialog`, `RootDropZone`,
`UnsortedEntry`, `useFolderMutations`). Two pieces left inline don't match
that precedent:

- The "Delete folder" confirmation dialog, which is structurally identical
  to `features/gallery/components/DeleteImageDialog.tsx`.
- The "Trash" nav row (active-state styling, `onClick`, a context menu, an
  "Empty trash" confirmation dialog, and `emptyTrashMutation`) — the same
  *kind* of thing as `UnsortedEntry.tsx`, just left inline with its own
  state/mutation instead of extracted to its own file.

Both changes are zero-functional-change extractions — no new state shape,
no new network calls, no prop changes on `FolderSidebar` itself.

## Goals / Non-Goals

**Goals:**
- Extract the inline delete-folder dialog to match the `DeleteImageDialog`
  precedent already established in the gallery feature.
- Extract the inline "Trash" nav row to match the `UnsortedEntry` precedent
  already established in this feature.

**Non-Goals:**
- No change to `FolderSidebar`'s props, `AppLayout`'s usage, or any
  rendered markup/styling.
- No change to the create/rename dialog flow (`nameDialog` state machine) —
  this was identified as a possible follow-up during exploration but is a
  "nice to have if already in there," not bundled into this change.
- No new shared/generic confirmation-dialog component in `components/ui/` —
  considered during exploration and rejected: the existing `Dialog`/
  `DialogFooter`/`Button` primitives are already the shared layer, and a
  generic wrapper would save only ~5-8 lines per call site while
  introducing a new cross-feature abstraction (a Decision Boundary item).
  `DeleteFolderDialog` and `EmptyTrashDialog` follow the existing
  feature-owned, presentational pattern instead.

## Decisions

### D1: `DeleteFolderDialog` mirrors `DeleteImageDialog` exactly

```ts
interface DeleteFolderDialogProps {
  folder: Folder | null
  onCancel: () => void
  onConfirm: () => void
}
```

Presentational only — `FolderSidebar` keeps owning `deleteTarget` state and
`handleDelete` (which calls `deleteMutation.mutate(deleteTarget.id)`), the
same way `ImageGrid` keeps owning `confirmDeleteImage`/`confirmDelete` for
`DeleteImageDialog`. Only the dialog markup (open-state binding, title,
body text, Cancel/Delete buttons) moves into the new file.

### D2: `TrashEntry` mirrors `UnsortedEntry`'s file shape, but isn't a drop target

```ts
interface TrashEntryProps {
  active: boolean
  onClick: () => void
}
```

Unlike `UnsortedEntry`, the Trash row is not a `useDroppable` target today,
so `TrashEntry` doesn't need `activeDragType`. What it does own, fully
self-contained (same way `UnsortedEntry` owns its `useDroppable`):

- The `ContextMenu`/`ContextMenuTrigger`/`ContextMenuContent` with the
  "Empty trash" item.
- `confirmEmptyTrash` state (renamed to a local `open` state).
- `emptyTrashMutation` (`useKindeAuth`, `useQueryClient`,
  `invalidateQueries(['images', 'trash'])`, `toast`) — moved verbatim.
- Renders `<EmptyTrashDialog open={open} onCancel={...} onConfirm={...} />`.

### D3: `EmptyTrashDialog` is presentational, boolean-driven

```ts
interface EmptyTrashDialogProps {
  open: boolean
  onCancel: () => void
  onConfirm: () => void
}
```

Unlike `DeleteFolderDialog`/`DeleteImageDialog` (driven by a nullable
target item), there's no "item" for emptying trash — just a yes/no
confirmation — so `open` is a plain boolean, owned and toggled by
`TrashEntry`.

### D4: Sequencing

The two extractions (`DeleteFolderDialog`, and `TrashEntry` +
`EmptyTrashDialog`) touch disjoint pieces of `FolderSidebar.tsx` and have no
dependency on each other — they can be implemented and tested in either
order.

## Risks / Trade-offs

- **[Risk]** The existing `FolderSidebar.test.tsx` "FolderSidebar trash
  context menu" describe block exercises the empty-trash flow through
  `FolderSidebar`. → **Mitigation**: those scenarios move to
  `TrashEntry.test.tsx` (testing `TrashEntry` directly), and are removed
  from `FolderSidebar.test.tsx` rather than duplicated.
- **[Trade-off]** `TrashEntry` ends up owning a `useMutation` directly
  (rather than going through `useFolderMutations`), matching how
  `emptyTrashMutation` already worked in `FolderSidebar` — this is a pure
  relocation, not a new pattern, so it's left as-is.
