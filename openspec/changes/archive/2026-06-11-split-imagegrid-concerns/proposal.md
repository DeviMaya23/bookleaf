## Why

`frontend/src/features/gallery/components/ImageGrid.tsx` (~399 lines) is now
the largest non-test file in `frontend/src/`, having absorbed four largely
independent responsibilities: view-driven data fetching (a 4-way switch over
`AppView['type']` for query keys/fetchers), manual-reorder drag-and-drop
persistence (optimistic local ordering, position mutation, trigger-driven
reorder effect), image lifecycle actions (delete/restore/hard-delete
mutations and a confirm dialog), and the masonry rendering itself. This is
the same shape of problem the prior `refactor-fe-folder-structure` change
addressed in `AppLayout.tsx` — and since the gallery is the app's main view,
it's the file most likely to be touched (and most likely to keep growing) as
agent-driven changes continue.

This change splits `ImageGrid.tsx` into three focused hooks plus a thin
rendering shell, following the same hook-extraction pattern already
established by `useFolderMutations`, `useFieldAutosave`, and
`useImageDetailsData`. It is a structural refactor: no user-visible behavior
changes (with one explicit decision point — see design.md — about whether
trash-restore should become optimistic for consistency, which if accepted
would be a small, separately-flagged behavior tweak).

## What Changes

- Extract `useImageLifecycle` (delete/restore/hard-delete mutations,
  `handleAction` routing, confirm-delete dialog state) into
  `features/gallery/hooks/useImageLifecycle.ts`, plus a presentational
  `DeleteImageDialog` component into
  `features/gallery/components/DeleteImageDialog.tsx`.
- Introduce a `removeImage(id)` callback — initially defined inline in
  `ImageGrid` and passed into `useImageLifecycle` — as the seed of a
  semantic mutation API for the grid's locally-held image list.
- Extract `useManualReorder` (the `orderedImages` state, `dragOverId`,
  the `sortEndTrigger`-consuming effect, `positionMutation`,
  `useDndMonitor`, `sortableItems`) into
  `features/gallery/hooks/useManualReorder.ts`. `removeImage` moves to live
  inside this hook and is returned from it; `useImageLifecycle`'s call site
  is unchanged.
- As part of `useManualReorder`, wrap `fetchedImages`/`allImages` in
  `useMemo` for referential stability (required for the hook's resync effect
  to depend on its `images` input directly).
- Decide (see design.md) whether `restoreMutation` should also call
  `removeImage` for consistency with delete/hard-delete's optimistic
  removal, vs. keeping today's refetch-based restore.
- Extract `useGalleryImages` (`queryKeyFor`/`fetcherFor`/`useInfiniteQuery`
  + the memoized images list) into
  `features/gallery/hooks/useGalleryImages.ts`.
- Remove `ImageGrid`'s redeclared `SortBy`/`SortDir` types; import the
  existing exports from `features/gallery/hooks/useGalleryControls.ts`
  instead.
- Add focused test files alongside each new hook
  (`useImageLifecycle.test.ts`, `useManualReorder.test.ts`,
  `useGalleryImages.test.ts`); prune `ImageGrid.test.tsx` of the scenarios
  that move to each, as each step lands.
- Document two follow-ups in design.md, explicitly deferred to future
  proposals (not implemented here):
  - Aligning `RightPanel`'s `tags` (inline) and `selectedFolders`
    (raw-setter) state with the "local optimistic state + semantic
    mutator" convention this change establishes via `useManualReorder`.
  - Whether `useGalleryImages`'s resulting parameter list should be
    consolidated into a shared `GalleryQuery` type alongside
    `useGalleryControls` and `GalleryToolbar`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

`frontend-structure` — adds requirements describing the new hook/component
file locations this change introduces under `features/gallery/` (see
`specs/frontend-structure/spec.md`). This is a zero-functional-change
structural refactor; the capability change is about file organization, not
behavior. (If the restore-optimism decision in design.md is accepted, that is
a minor, separately-flagged UX tweak to trash behavior, not a change to this
proposal's capability scope.)

## Impact

- `frontend/src/features/gallery/components/ImageGrid.tsx` shrinks
  substantially to a rendering shell composing the three new hooks.
- New files: `features/gallery/hooks/useImageLifecycle.ts`,
  `features/gallery/components/DeleteImageDialog.tsx`,
  `features/gallery/hooks/useManualReorder.ts`,
  `features/gallery/hooks/useGalleryImages.ts`, plus their test files.
- `ImageGrid.test.tsx` is pruned as scenarios move to the new hooks' test
  files.
- No changes to `AppLayout`, `GalleryToolbar`, `useGalleryControls`, or any
  other feature's public exports — `ImageGrid`'s own props are unchanged.
- No backend, API, database, or browser-extension changes. No new
  dependencies.
