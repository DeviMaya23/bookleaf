## Why

`frontend/src/` is currently flat: `components/`, `lib/`, `hooks/`, and `pages/`
each hold files for every UI area side by side, with no convention for where
new feature work belongs. This has let several files grow into god-components
that mix multiple features' state and UI — `AppLayout.tsx` (~650 lines) owns
gallery search/sort/filter, image-selection, viewer, and upload-drop state on
top of its actual job as the app shell; `FolderSidebar.tsx` (~495 lines) and
`RightPanel.tsx` (~387 lines) each bundle several internally-repetitive
concerns. As more agent-driven changes land, this flat structure makes it
harder to find where code belongs and easier for these files to keep growing.

This change reorganizes `frontend/src/` around features and, in the same
pass, splits the three god-files into smaller cohesive units along feature
lines. It is a structural refactor: no user-visible behavior changes.

## What Changes

- Introduce `frontend/src/features/<feature>/` for: `gallery`, `viewer`,
  `folders`, `image-details`, `upload`, `auth`. Each feature gets its own
  `components/`, `hooks/`, and `lib/` as needed.
- Introduce `frontend/src/app-shell/` for `AppLayout` and cross-cutting
  orchestration (drag-and-drop coordination, view routing).
- Keep `frontend/src/lib/` as the shared domain layer (`images.ts`,
  `folders.ts`, `tags.ts`, `thumbnail.ts`, `view.ts`, `utils.ts`, `browser.ts`,
  `fracdex.ts`, `api.ts`) — these are used across nearly every feature and are
  not feature-scoped.
- Keep `frontend/src/components/ui/`, `frontend/src/hooks/` (generic hooks
  only), `frontend/src/pages/` (thin route entry points), and
  `frontend/src/assets/` as shared/top-level.
- Split `AppLayout.tsx` by extracting:
  - `useAppView` + view/sort/filter default helpers → `app-shell/`
  - `useAppDragAndDrop` (sensors, drag overlay, drag start/end handlers) →
    `app-shell/`
  - `useGalleryControls` (search/sort/filter state) + `GalleryToolbar`
    (the corresponding toolbar UI) → `features/gallery/`
- Split `FolderSidebar.tsx` by extracting:
  - `buildFolderTree` / `filterFolderTree` → `features/folders/lib/folderTree.ts`
    (pure functions, no React)
  - `FolderItem`, `UnsortedEntry`, `RootDropZone` → their own component files
    in `features/folders/components/`
  - folder create/rename/delete mutations → `useFolderMutations` hook
  - collapse the three near-identical `FolderNameDialog` usages (new root
    folder, new subfolder, rename) into a single dialog driven by one piece
    of state
- Split `RightPanel.tsx` by extracting:
  - tag resolve-or-create logic (`handleTagsChange` body) →
    `resolveOrCreateTags` in `lib/tags.ts`
  - the three repeated blur/dirty-check field handlers (title, description,
    source URL) → a shared `useFieldAutosave` hook
  - the image/folders/tags query bundling + selected-folders sync →
    `useImageDetailsData` hook
- Update all import paths accordingly; relocate and split test files to match
  their new co-located unit (e.g. `AppLayout.test.tsx` is pruned to
  shell-composition concerns, with new focused test files for
  `useGalleryControls`, `useFolderMutations`, etc.)
- Document the new directory structure and conventions in `CONVENTIONS.md` so
  future feature work follows this layout.

## Capabilities

### New Capabilities

- `frontend-structure`: defines the `frontend/src/` directory conventions
  (feature-based `features/<feature>/`, `app-shell/` for the shell and
  cross-cutting orchestration, shared `lib/`/`components/ui/`/`hooks/`/
  `pages/`) that this change establishes and that future frontend work
  follows. Mirrors the existing `backend-structure` capability.

### Modified Capabilities

None — no user-facing behavior changes. All existing functionality (gallery
search/sort/filter, folder tree CRUD and drag-and-drop, image detail editing,
upload, viewer, focus mode, auth) is preserved exactly as specified in
`fe-gallery-*`, `fe-folder-panel`, `fe-right-panel`, `fe-image-tagging`,
`fe-image-upload-flow`, `fe-image-viewer*`, `focus-mode`, and `fe-kinde-auth`.

## Impact

- All files under `frontend/src/components/`, `frontend/src/hooks/`, and
  `frontend/src/lib/` move or are split; all import paths across
  `frontend/src/` are updated accordingly.
- `frontend/src/components/AppLayout.tsx` → `frontend/src/app-shell/AppLayout.tsx`,
  shrinks substantially; its test file is split across the new units.
- `frontend/src/components/FolderSidebar.tsx` and `RightPanel.tsx` are split
  into multiple smaller files within `features/folders/` and
  `features/image-details/` respectively.
- `CONVENTIONS.md` gains a "Frontend Architecture / Directory Structure"
  section analogous to the existing backend one.
- No backend, API, database, or browser-extension changes.
- No new dependencies.
