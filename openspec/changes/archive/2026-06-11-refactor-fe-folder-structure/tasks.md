## 1. Folder sidebar feature

- [x] 1.1 Create `frontend/src/features/folder-sidebar/{components,hooks,lib}/`
- [x] 1.2 Extract `buildFolderTree`/`filterFolderTree` into
      `features/folder-sidebar/lib/folderTree.ts` (pure functions) with unit
      tests
- [x] 1.3 Extract `FolderItem`, `UnsortedEntry`, `RootDropZone` into their own
      component files under `features/folder-sidebar/components/`
- [x] 1.4 Extract folder create/rename/delete mutations into
      `useFolderMutations` hook with unit tests
- [x] 1.5 Consolidate the three `FolderNameDialog` usages (new root folder,
      new subfolder, rename) into a single `NameDialogState`-driven dialog
      (per design.md D6); update `FolderSidebar.test.tsx` accordingly
- [x] 1.6 Move `FolderSidebar.tsx`, `FolderSidebar.test.tsx`, and
      `FolderNameDialog.tsx` into `features/folder-sidebar/components/`,
      updating all import paths (including the `AppLayout` import of
      `FolderSidebar`)
- [x] 1.7 Run `npm run test` for the folder-sidebar feature and fix any
      failures

## 2. Right panel feature

- [x] 2.1 Create `frontend/src/features/right-panel/{components,hooks,lib}/`
- [x] 2.2 Extract the tag resolve-or-create logic from `handleTagsChange`
      into `resolveOrCreateTags` in `lib/tags.ts` with unit tests
- [x] 2.3 Extract the three repeated blur/dirty-check field handlers
      (title, description, source URL) into `useFieldAutosave` (per
      design.md D7) with unit tests
- [x] 2.4 Extract the image/folders/tags query bundling and selected-folders
      sync effect into `useImageDetailsData` with unit tests
- [x] 2.5 Move `RightPanel.tsx`, `RightPanel.test.tsx`,
      `FolderPanelContent.tsx`, `TagInput.tsx`, `TagInput.test.tsx`,
      `FolderInput.tsx`, and `FolderInput.test.tsx` into
      `features/right-panel/components/`, updating all import paths
- [x] 2.6 Move `useVisionSuggestion.ts` (and its test, if any) into
      `features/right-panel/hooks/`, updating the `AppLayout` import
- [x] 2.7 Run `npm run test` for the right-panel feature and fix any
      failures

## 3. Viewer, upload, and auth features (pure moves)

- [x] 3.1 Move `ImageViewer.tsx`/`ImageViewer.test.tsx` into
      `frontend/src/features/viewer/components/`, updating imports
- [x] 3.2 Move `UploadModal.tsx`/`.test.tsx` and
      `BatchUploadModal.tsx`/`.test.tsx` into
      `frontend/src/features/upload/components/`, updating imports
- [x] 3.3 Move `AuthGuard.tsx`/`.test.tsx` and
      `ProfileMenu.tsx`/`.test.tsx` into
      `frontend/src/features/auth/components/`, and `lib/me.ts` into
      `frontend/src/features/auth/lib/`, updating imports (including the
      `ProfileMenu` import inside `features/folder-sidebar`)
- [x] 3.4 Run `npm run test` and fix any failures

## 4. Gallery feature

- [x] 4.1 Create `frontend/src/features/gallery/{components,hooks,lib}/`
- [x] 4.2 Move `ImageGrid.tsx`/`.test.tsx` and
      `MasonryLayout.tsx`/`.test.tsx` into `features/gallery/components/`,
      updating imports
- [x] 4.3 Move `lib/masonry.ts` into `features/gallery/lib/masonry.ts`,
      updating imports
- [x] 4.4 Extract `useGalleryControls` (search/sort/filter state, currently
      in `AppLayout`) into `features/gallery/hooks/` with unit tests
- [x] 4.5 Extract `GalleryToolbar` (search/sort/filter UI, currently in
      `AppLayout`) into `features/gallery/components/` with unit tests
- [x] 4.6 Wire `AppLayout` to use `useGalleryControls` + `GalleryToolbar`,
      removing the extracted state and JSX from `AppLayout`
- [x] 4.7 Update `AppLayout.test.tsx` to remove coverage now owned by
      `useGalleryControls`/`GalleryToolbar` tests
- [x] 4.8 Run `npm run test` and fix any failures

## 5. App shell

- [x] 5.1 Create `frontend/src/app-shell/`
- [x] 5.2 Extract `useAppView` into `app-shell/useAppView.ts` with unit tests.
      (The view/sort/filter default helpers `defaultSortForViewType`/
      `filterSectionsForViewType` were colocated with `useGalleryControls` in
      §4 — their only consumer — rather than app-shell, since `useAppView`
      does not use them.)
- [x] 5.3 Extract `useAppDragAndDrop` (sensors, `DragOverlay`,
      `handleDragStart`/`handleDragEnd`) into `app-shell/useAppDragAndDrop.tsx`
      (`.tsx` because it returns the `DragOverlay` element) with unit tests
- [x] 5.4 Move `lib/dragHandlers.ts`/`.test.ts` into `app-shell/lib/`,
      updating imports
- [x] 5.5 Move `AppLayout.tsx` into `app-shell/AppLayout.tsx`, wiring in
      `useAppView` and `useAppDragAndDrop`, updating the import in `App.tsx`
- [x] 5.6 Move and prune `AppLayout.test.tsx` to
      `app-shell/AppLayout.test.tsx`, retaining at least one integration
      test for cross-feature flows (double-click sets `selectedImage` and
      `viewerImage`; deleting the active image clears both) per design.md D5
- [x] 5.7 Run `npm run test` and fix any failures

## 6. Documentation

- [x] 6.1 Add a "Frontend Architecture / Directory Structure" section to
      `CONVENTIONS.md`, documenting `features/<feature>/`, `app-shell/`, and
      the shared `lib/`/`components/ui/`/`hooks/`/`pages/` layout

## 7. Final verification

- [x] 7.1 Run `npm run build` from `frontend/` and fix any issues
- [x] 7.2 Run the full frontend test suite and fix any issues
