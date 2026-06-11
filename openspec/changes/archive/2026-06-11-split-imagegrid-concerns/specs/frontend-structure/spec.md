## ADDED Requirements

### Requirement: Gallery grid concerns are split into focused hooks under features/gallery/hooks/
`ImageGrid.tsx`'s data-fetching, manual-reorder, and image-lifecycle concerns SHALL each be extracted into a named hook under `frontend/src/features/gallery/hooks/`, with a colocated test file, rather than bundled together in a single component file.

#### Scenario: Image lifecycle hook exists
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/features/gallery/hooks/useImageLifecycle.ts` and
  `frontend/src/features/gallery/hooks/useImageLifecycle.test.ts` exist

#### Scenario: Manual reorder hook exists
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/features/gallery/hooks/useManualReorder.ts` and
  `frontend/src/features/gallery/hooks/useManualReorder.test.ts` exist

#### Scenario: Gallery images data-fetching hook exists
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/features/gallery/hooks/useGalleryImages.ts` and
  `frontend/src/features/gallery/hooks/useGalleryImages.test.ts` exist

### Requirement: Delete-permanently confirmation is a separate presentational component
The "delete permanently" confirm dialog SHALL be its own presentational component under `frontend/src/features/gallery/components/`, rather than inline JSX within `ImageGrid.tsx`.

#### Scenario: DeleteImageDialog component exists
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/features/gallery/components/DeleteImageDialog.tsx`
  exists
