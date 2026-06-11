## 1. Extract DetailsGrid

- [x] 1.1 Create `frontend/src/features/right-panel/components/DetailsGrid.tsx`, taking `image: Image` as its only prop, containing the `formatFileSize`/`formatDate` helpers and the Size/Dimensions/Added grid markup moved from `RightPanel.tsx`.
- [x] 1.2 Replace the inline "Details" section in `ImagePanelBody` with `<DetailsGrid image={image} />`; remove the now-unused helpers and markup from `RightPanel.tsx`.
- [x] 1.3 Add `DetailsGrid.test.tsx` covering size/dimensions/date formatting (including the `null`/missing-dimension fallback cases).

## 2. Extract DownloadButton

- [x] 2.1 Create `frontend/src/features/right-panel/components/DownloadButton.tsx`, taking `imageId: string` as its only prop, containing the `handleDownload`/`isDownloading` state and the sticky-footer button markup moved from `RightPanel.tsx`.
- [x] 2.2 Replace the sticky footer in `ImagePanelBody` with `<DownloadButton imageId={image.id} />`; remove the now-unused `handleDownload`/`isDownloading`/`downloadImage` usage from `RightPanel.tsx`.
- [x] 2.3 Add `DownloadButton.test.tsx` covering the default/loading button states and the download-trigger call.

## 3. FolderPanelContent autosave dedup

- [x] 3.1 Replace `FolderPanelContent`'s local `name`/`description` state, `origName`/`origDescription` refs, and `handleNameBlur`/`handleDescriptionBlur` with two `useFieldAutosave` instances (name with `isEmpty: (v) => v.trim() === ''`, description with no `isEmpty`), saving via the existing `saveMutation`.
- [x] 3.2 In `RightPanel.tsx`, add `key={props.folder.id}` to `<FolderPanelContent />` so switching folders remounts the component and resets both autosave instances.
- [x] 3.3 Update `FolderPanelContent`'s tests (or `RightPanel.test.tsx` scenarios covering folder mode) to verify name/description autosave-on-blur and the empty-name revert, matching prior coverage.

## 4. Unify FolderInput/TagInput on TokenInput

- [x] 4.1 Create `frontend/src/features/right-panel/components/TokenInput.tsx`: generic `TokenInput<T extends { id: string; name: string }>` with `items`, `onChange`, `disabled`, `suggestions`, `placeholder`, and optional `createFromText` props, covering the shared chip rendering, dropdown filtering/keyboard navigation, blur-timer, and reference-based removal described in design.md.
- [x] 4.2 Reimplement `FolderInput.tsx` as a thin wrapper over `TokenInput<Folder>` with placeholder `"Add to folder…"` and no `createFromText`, preserving its existing `FolderInputProps` (`folders`, `onChange`, `disabled`, `suggestions`).
- [x] 4.3 Reimplement `TagInput.tsx` as a thin wrapper over `TokenInput<Tag>` with placeholder `"Add tags…"` and a `createFromText` that reproduces `commitRaw`'s normalization (lowercase, strip commas, trim, dedup-by-name), preserving its existing `TagInputProps` (`tags`, `onChange`, `disabled`, `suggestions`).
- [x] 4.4 Add `TokenInput.test.tsx` covering chip add/remove, dropdown keyboard navigation, blur-close behavior, and `createFromText` add/dedup/empty-skip.
- [x] 4.5 Adapt `FolderInput.test.tsx` and `TagInput.test.tsx` to exercise the wrappers, removing scenarios now covered by `TokenInput.test.tsx` and keeping only wrapper-specific assertions (placeholder text, item type, free-text behavior presence/absence).

## 5. Relocate useVisionSuggestion

- [x] 5.1 Move `features/right-panel/hooks/useVisionSuggestion.ts` and its test to `app-shell/useVisionSuggestion.ts` / `app-shell/useVisionSuggestion.test.ts`, with no internal changes.
- [x] 5.2 Update `AppLayout.tsx`'s import path for `useVisionSuggestion`.

## 6. Final checks

- [x] 6.1 Prune `RightPanel.test.tsx` of scenarios now covered by `DetailsGrid.test.tsx` and `DownloadButton.test.tsx`.
- [x] 6.2 Run `npm run build` and fix any resulting issues.
