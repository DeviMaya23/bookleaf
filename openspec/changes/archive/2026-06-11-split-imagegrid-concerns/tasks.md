## 1. Step 1: Extract `useImageLifecycle` + `DeleteImageDialog`

- [x] 1.1 Create `features/gallery/hooks/useImageLifecycle.ts`: move
      `deleteMutation`, `restoreMutation`, `hardDeleteMutation`,
      `handleAction`, and the confirm-delete dialog open/close state out of
      `ImageGrid.tsx`.
- [x] 1.2 In `ImageGrid.tsx`, define `removeImage(id)` inline (filters
      `orderedImages`) and pass it to `useImageLifecycle` per D2; wire
      `delete`/`hardDelete` to call it, leave `restore` refetch-based per D4
      (Option A).
- [x] 1.3 Create `features/gallery/components/DeleteImageDialog.tsx`
      (presentational), moving the confirm-delete dialog JSX out of
      `ImageGrid.tsx`; render it from `ImageGrid` driven by
      `useImageLifecycle`'s state.
- [x] 1.4 Add `features/gallery/hooks/useImageLifecycle.test.ts`: cover
      `handleAction` routing (trash → restore, else → delete), confirm-dialog
      open/confirm/cancel, mutation success/error toasts, `onImageDeleted`
      invocation, and `removeImage` invocation on delete/hard-delete (not on
      restore, per D4).
- [x] 1.5 Prune `ImageGrid.test.tsx` of the lifecycle scenarios now covered by
      `useImageLifecycle.test.ts`.
- [x] 1.6 Run FE build and tests; confirm green before moving to Step 2.

## 2. Step 2: Extract `useManualReorder`

- [x] 2.1 Wrap `fetchedImages` and `allImages` in `useMemo` in `ImageGrid.tsx`
      (D3 prerequisite) — no behavior change, just memoization.
- [x] 2.2 Create `features/gallery/hooks/useManualReorder.ts`: move
      `orderedImages` state, `dragOverId`, `useDndMonitor`, the
      `sortEndTrigger`-consuming effect, `orderedImagesRef`,
      `lastProcessedTriggerTs`, `positionMutation`, and `sortableItems`. Move
      `removeImage` (from Step 1) to live inside this hook and return it.
- [x] 2.3 Update `useImageLifecycle`'s call site in `ImageGrid.tsx` to use the
      `removeImage` returned from `useManualReorder` — no change to
      `useImageLifecycle`'s own signature.
- [x] 2.4 Add `features/gallery/hooks/useManualReorder.test.ts`: via
      `renderHook` wrapped in a minimal `<DndContext>`, drive
      `sortEndTrigger` sequences — successful reorder + position persistence,
      rollback on `positionMutation` error, the duplicate-trigger guard
      (`lastProcessedTriggerTs`), and `dragOverId` updates.
- [x] 2.5 Prune `ImageGrid.test.tsx` of the reorder scenarios now covered by
      `useManualReorder.test.ts`.
- [x] 2.6 Run FE build and tests; confirm green before moving to Step 3.

## 3. Step 3: Extract `useGalleryImages` + resolve `SortBy`/`SortDir` duplication

- [x] 3.1 Remove `ImageGrid.tsx`'s local `SortBy`/`SortDir` type declarations
      and import them from `features/gallery/hooks/useGalleryControls`
      instead (D5).
- [x] 3.2 Create `features/gallery/hooks/useGalleryImages.ts`: move
      `EMPTY_FILTER`, `sortParamsFor`, `queryKeyFor`, `fetcherFor`, the
      `useInfiniteQuery` call, and the memoized `fetchedImages`/`allImages`
      computation (already memoized from Step 2) out of `ImageGrid.tsx`.
- [x] 3.3 Wire `useGalleryImages`'s `images` output into `useManualReorder` as
      its `images` input.
- [x] 3.4 Add `features/gallery/hooks/useGalleryImages.test.ts`: cover
      `queryKeyFor`/`fetcherFor` branch behavior per `view.type`, and that
      `images` is referentially stable across re-renders when inputs are
      unchanged.
- [x] 3.5 Prune `ImageGrid.test.tsx` down to shell-composition scenarios:
      loading/empty/grid rendering states, `SortableContext` wrapping
      condition, container-width-driven layout, and that the three hooks'
      outputs are wired into the render correctly.
- [x] 3.6 Run FE build and tests; confirm green.

## 4. Final verification

- [x] 4.1 Run `npm run build` from `frontend/` and fix any type or lint
      issues that arise.
- [x] 4.2 Run the full FE test suite and confirm all tests pass.
