## ADDED Requirements

### Requirement: Upload pipeline and file validation are shared via lib/upload.ts
The "validate file type → convert HEIC to JPEG if needed → initiate upload →
generate thumbnail → upload to R2 → complete upload" pipeline, and the
file-type/HEIC-support validation it depends on, SHALL be implemented once
in `frontend/src/lib/upload.ts` and reused by `app-shell` and
`features/upload`, rather than each call site duplicating
`ACCEPTED_TYPES`, `fileBaseName`, the HEIC/Safari check, and the
upload-pipeline steps.

#### Scenario: Shared upload module exists
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/lib/upload.ts` and
  `frontend/src/lib/upload.test.ts` exist, and export
  `validateImageFile`, `fileBaseName`, and `uploadImageFile`

#### Scenario: Call sites delegate to the shared pipeline
- **WHEN** `frontend/src/app-shell/lib/dragHandlers.ts`,
  `frontend/src/features/upload/components/UploadModal.tsx`, and
  `frontend/src/features/upload/components/BatchUploadModal.tsx` are
  inspected
- **THEN** none of them declare their own `ACCEPTED_TYPES` constant or
  HEIC/Safari validation check, and each calls `uploadImageFile` from
  `frontend/src/lib/upload.ts` to run the upload pipeline

## MODIFIED Requirements

### Requirement: Shared domain modules remain in lib/
Domain types and API wrappers used across multiple features (`images`, `folders`, `tags`, `thumbnail`, `upload`, `view`) SHALL remain in `frontend/src/lib/` and SHALL NOT be duplicated or relocated into a single feature's directory.

#### Scenario: Domain modules remain in lib/
- **WHEN** the codebase is inspected
- **THEN** `frontend/src/lib/images.ts`, `frontend/src/lib/folders.ts`,
  `frontend/src/lib/tags.ts`, `frontend/src/lib/thumbnail.ts`,
  `frontend/src/lib/upload.ts`, and `frontend/src/lib/view.ts` exist
