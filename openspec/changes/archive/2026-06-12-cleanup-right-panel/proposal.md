## Why

`frontend/src/features/right-panel/components/RightPanel.tsx` (288 lines) is
now larger than `ImageGrid.tsx` was after `split-imagegrid-concerns` (221
lines). Unlike `ImageGrid`, its size isn't from bundling orthogonal concerns
— `ImagePanelBody` is fundamentally one concern (edit this image's metadata)
expressed across several similar field sections. A survey of the
`right-panel` feature surfaced four smaller, independently-justified cleanups
that reduce its size and remove real duplication without restructuring that
core orchestration role.

## What Changes

- Extract the "Details" grid (Size/Dimensions/Added, plus the
  `formatFileSize`/`formatDate` helpers) out of `RightPanel.tsx` into a new
  presentational `features/right-panel/components/DetailsGrid.tsx`, taking
  only `image` as a prop.
- Extract the download button (`handleDownload`, `isDownloading` state) out
  of `RightPanel.tsx` into a new self-contained
  `features/right-panel/components/DownloadButton.tsx`, taking only
  `imageId` as a prop.
- Replace `FolderPanelContent`'s hand-rolled title/description autosave
  (local `useState` + `origRef` + blur-diff, duplicating
  `useFieldAutosave`'s logic) with `useFieldAutosave` itself.
- Introduce a shared `features/right-panel/components/TokenInput.tsx`
  generic chip-input-with-autocomplete component, and reimplement
  `FolderInput` and `TagInput` as thin configurations of it. The two
  components are ~90% identical (same state, keyboard navigation, dropdown,
  and blur-timer logic); they differ only in item shape (`Folder` vs `Tag`,
  both `{id, name}`), whether free-text entries can be created
  (`TagInput` only, via `commitRaw`), and placeholder text.
- Relocate `useVisionSuggestion` from `features/right-panel/hooks/` to
  `app-shell/useVisionSuggestion.ts` — it is only consumed by `AppLayout`
  (for the post-upload "suggested folder" toast) and has no relationship to
  `RightPanel` or image-detail editing.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

`frontend-structure` — adds requirements describing the new component file
locations (`DetailsGrid.tsx`, `DownloadButton.tsx`, `TokenInput.tsx`) and the
relocation of `useVisionSuggestion.ts` out of `features/right-panel/hooks/`
into `app-shell/`. This is a zero-functional-change structural cleanup; no
behavior changes.

## Impact

- `frontend/src/features/right-panel/components/RightPanel.tsx` shrinks —
  loses the details grid, the download button, and the format helpers.
- New files: `features/right-panel/components/DetailsGrid.tsx`,
  `features/right-panel/components/DownloadButton.tsx`,
  `features/right-panel/components/TokenInput.tsx`, plus test files for
  each.
- `FolderInput.tsx` and `TagInput.tsx` are reimplemented on top of
  `TokenInput`; their existing test files are adapted accordingly.
- `FolderPanelContent.tsx` drops its inline autosave state/handlers in favor
  of `useFieldAutosave`.
- `features/right-panel/hooks/useVisionSuggestion.ts` (and its test) moves to
  `app-shell/useVisionSuggestion.ts`; `AppLayout.tsx`'s import path updates
  accordingly.
- No change to `RightPanel`'s external props, `AppLayout`'s usage of
  `RightPanel`, or any user-visible behavior.
- No backend, API, database, or browser-extension changes. No new
  dependencies.
