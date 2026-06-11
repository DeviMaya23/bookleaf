## Context

`ImageGrid.tsx` (~399 lines) currently bundles four largely independent
concerns:

- **Data fetching** — `sortParamsFor`/`queryKeyFor`/`fetcherFor` (a 4-way
  switch over `AppView['type']`), `useInfiniteQuery`, and the
  `fetchedImages`/`allImages` derivation (with client-side search/tag/mime
  filtering for folder views only).
- **Manual reorder** — `orderedImages` state, `dragOverId` +
  `useDndMonitor`, the `sortEndTrigger`-consuming effect (optimistic
  reorder, `computeNewPosition`, `positionMutation`, rollback on error), and
  two refs (`orderedImagesRef`, `lastProcessedTriggerTs`) needed to make that
  effect work correctly.
- **Image lifecycle** — `deleteMutation`/`restoreMutation`/
  `hardDeleteMutation`, `handleAction`, and the "delete permanently" confirm
  dialog.
- **Rendering** — container-width measurement via `ResizeObserver`,
  `MasonryLayout` + `ImageCard`, loading/empty/load-more states.

`SortBy`/`SortDir` are also redeclared locally in `ImageGrid.tsx`, duplicating
the types already exported from `features/gallery/hooks/useGalleryControls.ts`
(which `GalleryToolbar` imports).

Separately, `app-shell/useAppDragAndDrop.tsx` imports `SortEndTrigger` — a
type defined in `ImageGrid.tsx` — and is the producer of values of that type
(it detects "image dropped on image" and constructs the trigger). This
cross-file contract is not changed by this proposal; it's noted here so it
isn't mistaken for an oversight.

This change follows the hook-extraction pattern already established by
`useFolderMutations` (folder-sidebar), `useFieldAutosave` and
`useImageDetailsData` (right-panel), and `useGalleryControls`/
`useAppDragAndDrop` (gallery/app-shell) — all of which pulled a cohesive
piece of state+effects out of a larger component into a named hook with its
own test file.

## Goals / Non-Goals

**Goals:**
- Reduce `ImageGrid.tsx` to a thin shell that composes
  `useGalleryImages`, `useManualReorder`, and `useImageLifecycle`, plus
  container measurement and the masonry render.
- Each of the three extraction steps is independently shippable: build +
  tests green before moving to the next step (mirrors D4 of the prior
  `refactor-fe-folder-structure` design).
- Resolve the `SortBy`/`SortDir` duplication.
- Produce a second concrete instance of the "local optimistic state +
  semantic mutator" shape (via `useManualReorder`'s `removeImage`), as a
  precedent for the deferred RightPanel follow-up.

**Non-Goals:**
- No change to `ImageGrid`'s external props or any other component's public
  exports (`AppLayout`, `GalleryToolbar`, `useGalleryControls`,
  `useAppDragAndDrop`).
- No change to the cross-file `SortEndTrigger` contract between
  `features/gallery` and `app-shell` — it stays as-is.
- No consolidation of `useGalleryImages`'s parameters into a `GalleryQuery`
  type shared with `useGalleryControls`/`GalleryToolbar` — see Deferred
  Follow-ups.
- No changes to `RightPanel`'s `tags`/`selectedFolders` handling — see
  Deferred Follow-ups.
- Aside from the restore-optimism question (Decision D4), no behavior
  changes of any kind.

## Decisions

### D1: Sequencing — `useImageLifecycle` → `useManualReorder` → `useGalleryImages`

```
Step 1: useImageLifecycle      (no dependencies — can start immediately)
            │
            │  seeds `removeImage` (D2)
            ▼
Step 2: useManualReorder        (also needs the useMemo prerequisite, D3)
            │
            │  needs referentially-stable `images` to relocate (D3)
            ▼
Step 3: useGalleryImages         (also resolves SortBy/SortDir dup, D5)
```

Lifecycle (C) has the fewest dependencies — it only needs `isTrash`,
`onImageDeleted`, and a `removeImage` callback it doesn't need to define
itself. Reorder (B) is next because it's where `removeImage` actually lives
and where the `useMemo` prerequisite is introduced. Data fetching (A) is
last because its extraction is what makes the `useMemo`'d `images` available
to relocate, and because it's the step that reopens the `GalleryQuery`
question (deferred, not decided here) — doing it last means that question
doesn't block the other two steps.

**Alternative considered**: extract A first (it's the "most central"
concern). Rejected because A's extraction surfaces the `GalleryQuery`
consolidation question, which is explicitly deferred — starting there risks
that question blocking progress on the other two, lower-risk extractions.

### D2: `removeImage` handoff — defined in Step 1, relocated in Step 2

In Step 1, `ImageGrid` keeps `orderedImages`/`setOrderedImages` itself (not
yet extracted) and defines:

```ts
const removeImage = useCallback(
  (id: string) => setOrderedImages((prev) => prev.filter((img) => img.id !== id)),
  []
)
```

...passing it to `useImageLifecycle(isTrash, removeImage, onImageDeleted)`.
In Step 2, this exact function body moves inside `useManualReorder` (which
now owns `orderedImages`) and is returned from it; `useImageLifecycle`'s call
site is unchanged — only where `removeImage` is *defined* moves.

**Rationale**: Step 1 becomes a pure relocation with no new shared-state
design decisions of its own. The "who owns `orderedImages` and how do
outsiders mutate it" question is answered once, in Step 2, where the owner
actually changes — rather than being half-answered in Step 1 and revisited
in Step 2.

### D3: `useMemo` for `fetchedImages`/`allImages` (Step 2 prerequisite)

`useManualReorder`'s resync effect needs to depend on `images` directly
(`useEffect(() => setOrderedImages(images), [images])`). Today,
`fetchedImages = data?.pages.flatMap(...)` and
`allImages = fetchedImages.filter(...)` are recomputed every render, so a
naive `[images]` dependency would fire on every render and clobber any
in-flight optimistic reorder.

Step 2 wraps both in `useMemo`:
- `fetchedImages`: `useMemo(() => data?.pages.flatMap((p) => p.images) ?? [], [data])`
- `allImages`: `useMemo(() => ..., [fetchedImages, trimmedSearch, filterTagIds, filterMimeTypes])`

This is a no-op behaviorally (same values, just memoized) and is done in
Step 2 at `ImageGrid`'s current location — Step 3 then relocates the
already-memoized computation into `useGalleryImages` unchanged.

### D4: Restore-optimism — keep as-is (Option A, confirmed)

`deleteMutation` and `hardDeleteMutation` both call `removeImage(id)` for
instant removal from the displayed list. `restoreMutation` does not — it
only invalidates `['images', 'trash']` and relies on the resulting refetch
(prefix-matched against the trash view's query key) to make the restored
image disappear.

**Decision**: keep this asymmetry. `restoreMutation` is migrated into
`useImageLifecycle` as-is, without calling `removeImage`. This avoids
introducing a rollback-on-error case that doesn't exist today (if the
restore API call failed after an optimistic `removeImage`, the image would
need to reappear — delete/hard-delete don't handle this either, but they
don't need to since failure there just means "stays in trash," whereas
failed restore would mean "stays removed from trash but wasn't actually
restored").

**Alternative considered**: have `restoreMutation` also call `removeImage`
for consistency with delete/hard-delete. Rejected — it's a small UX
inconsistency, but fixing it introduces a new failure-handling case that's
out of scope for a structural refactor.

### D5: Resolve `SortBy`/`SortDir` duplication (Step 3)

`ImageGrid.tsx` redeclares:
```ts
type SortBy = 'manual' | 'created_at' | 'title'
type SortDir = 'asc' | 'desc'
```
identically to the types exported from `useGalleryControls.ts`. Step 3
removes the local declarations and imports `SortBy`/`SortDir` from
`features/gallery/hooks/useGalleryControls`, alongside the
`useGalleryImages` extraction (which uses these types in its signature).

### D6: Test strategy — new test per hook, `ImageGrid.test.tsx` pruned to shell composition

Mirrors D5 of the prior `refactor-fe-folder-structure` design: each
extraction step adds a focused test file in the same step, and
`ImageGrid.test.tsx` is pruned of only the scenarios that moved, in that same
step (not after).

- **`useImageLifecycle.test.ts`**: `handleAction` routing (trash → restore,
  else → delete), confirm-dialog open/confirm/cancel, mutation
  success/error toasts, `onImageDeleted` invocation, `removeImage`
  invocation on delete/hard-delete (and restore, per D4's resolution).
- **`useManualReorder.test.ts`**: via `renderHook`, drive `sortEndTrigger`
  sequences — successful reorder + position persistence, rollback on
  `positionMutation` error, the duplicate-trigger guard
  (`lastProcessedTriggerTs`), and `dragOverId` updates. `useDndMonitor`
  requires a `DndContext` ancestor — the test wraps `renderHook` in a
  minimal `<DndContext>` provider (consistent with how `dragHandlers.test.ts`
  / `AppLayout.test.tsx` already exercise dnd-kit).
- **`useGalleryImages.test.ts`**: `queryKeyFor`/`fetcherFor` branch coverage
  per `view.type`, and that `images` is referentially stable across
  re-renders when inputs are unchanged.
- **`ImageGrid.test.tsx`** (pruned): loading/empty/grid rendering states,
  `SortableContext` wrapping condition, container-width-driven layout, and
  that the three hooks' outputs are wired into the render correctly —
  i.e., shell composition, not each hook's internals.
- **`DeleteImageDialog`**: simple presentational component; covered via
  `useImageLifecycle`'s consumer-level tests rather than its own file,
  unless it grows additional states later.

## Risks / Trade-offs

- **Splitting `ImageGrid.test.tsx` could silently drop coverage** if an
  extracted hook's new test doesn't fully replicate what the original
  covered — mitigated by D6 (write new tests in the same step, prune
  originals only after, per step).
- **`useDndMonitor` requires a `DndContext` ancestor** — `useManualReorder`
  calling it is fine at runtime (ImageGrid already renders inside
  `AppLayout`'s `<DndContext>`), but `useManualReorder.test.ts` needs an
  explicit `<DndContext>` wrapper that the current `ImageGrid.test.tsx` may
  already provide via existing test utilities — confirm during Step 2.
- **D4 (restore-optimism) is a real, if small, behavior question** — flagged
  explicitly rather than decided silently, per CLAUDE.md's
  Mid-Implementation Adaptations guidance.
- **Step 3 is the step most likely to invite scope creep into the
  `GalleryQuery` consolidation** (Deferred Follow-ups below) — explicitly
  out of scope; if it starts to feel necessary mid-implementation, stop and
  raise it rather than proceeding.

## Deferred Follow-ups

These were identified during exploration but are explicitly **out of scope**
for this change:

1. **RightPanel convention alignment.** `useFieldAutosave` already follows
   the "local state synced from a prop + semantic mutator" shape cleanly.
   `useManualReorder`'s `removeImage` (this change) will be a second
   instance. RightPanel has two *unconverted* instances of the same shape:
   - `tags` state in `ImagePanelBody` — fully inline (local `useState`,
     resync `useEffect` on `image.id`, `handleTagsChange` as the mutator),
     never extracted into a named hook.
   - `selectedFolders` in `useImageDetailsData` — extracted into a hook, but
     the hook returns the raw `setSelectedFolders` setter rather than a
     semantic mutator; `ImagePanelBody` calls the setter directly and
     separately triggers `folderSaveMutation`.

   Once this change establishes the convention via `useManualReorder`,
   revisit RightPanel's `tags` and `selectedFolders` to match it. This is a
   candidate for a follow-up change once this one has landed and the
   convention is proven out.

2. **`GalleryQuery` consolidation.** `useGalleryImages` (Step 3) ends up with
   an 8-parameter signature (`view`, `searchTerm`, `debouncedSearchTerm`,
   `sortBy`, `sortDir`, `filterTagIds`, `filterMimeTypes`,
   `filterFolderIds`) — exactly the "query-relevant subset" of
   `useGalleryControls`'s 24-field return value. `GalleryToolbar` currently
   takes the entire `controls: GalleryControls` object; `ImageGrid` currently
   destructures 7 of these fields individually from `gallery.*` in
   `AppLayout`. Whether to consolidate these into a shared `GalleryQuery`
   type used by `useGalleryControls`, `GalleryToolbar`, and
   `useGalleryImages` is a separate decision — it introduces a new shared
   type crossing feature boundaries, which per CLAUDE.md's Decision
   Boundaries needs its own proposal and confirmation. Noted as a candidate
   follow-up only.

## Open Questions

None — D4 resolved (keep restore refetch-based, Option A).
