## Why

The "convert HEIC → initiate upload → generate thumbnail → upload to R2 ×2 →
complete upload" pipeline is implemented three times: in
`features/upload/components/UploadModal.tsx` (`uploadMutation`), in
`features/upload/components/BatchUploadModal.tsx` (`runUpload`), and in
`app-shell/lib/dragHandlers.ts` (`handleFileAutoUpload`, the single-file
canvas-drop flow). Each copy also redeclares `ACCEPTED_TYPES`, `fileBaseName`,
and the HEIC/Safari support check. This is real logic duplication with drift
risk — a fix like the archived `image-upload-dimensions-fix` change would
need to land in three places — not just a file-size concern like the prior
`split-imagegrid-concerns`/`cleanup-*` structural refactors.

This change extracts the shared pipeline and validation into a new
`frontend/src/lib/upload.ts`, alongside the existing `lib/images.ts`,
`lib/thumbnail.ts`, and `lib/browser.ts` that it composes. The two upload
modals stay separate components (their UX flows — compose-then-submit with
metadata/vision vs. drop-and-watch queue with retry — are genuinely
different, not accidentally duplicated), and the dropzone UI duplication
between them is explicitly out of scope for this change, to be handled as a
feature-local extraction separately.

## What Changes

- Add `frontend/src/lib/upload.ts` exporting:
  - `validateImageFile(file): 'unsupported_type' | 'heic_safari_only' | null`
    — pure, synchronous check combining `ACCEPTED_TYPES` and the HEIC/Safari
    check, replacing the three duplicated inline versions.
  - `fileBaseName(name): string` — moved as-is, replacing the three
    duplicated copies.
  - `uploadImageFile(getToken, params): Promise<CompleteUploadResult>` — the
    shared pipeline (validate → HEIC convert if needed → `initiateUpload` →
    `generateThumbnail` → parallel `putToR2` ×2 → `completeUpload`).
    `params: { file, folderId, title?, description?, sourceUrl? }`, with
    `title` defaulting to `fileBaseName(file.name)` when omitted.
- `app-shell/lib/dragHandlers.ts`'s `handleFileAutoUpload`: replace its
  inline validation + pipeline with `validateImageFile` (for the
  `unsupported_type`/`heic_safari_only` errors it already throws) and
  `uploadImageFile`, then keep its existing `getImage(result.image_id)` call
  (the only caller needing the full `ImageDetail` for `RightPanel`).
- `UploadModal.tsx`: replace `handleFile`'s inline type/HEIC checks with
  `validateImageFile` (same error-to-message mapping as today), and replace
  `uploadMutation`'s pipeline body with `uploadImageFile({ file, folderId,
  title: resolvedTitle, description, sourceUrl })`.
- `BatchUploadModal.tsx`: replace `makeBatchFile`'s inline type/HEIC checks
  with `validateImageFile` (its oversize check via `MAX_SIZE_BYTES` stays
  local — a batch-only UX cap, not a general upload-validity concern), and
  replace `runUpload`'s pipeline body with `uploadImageFile({ file,
  folderId, title: fileBaseName(file.name) })`. The concurrency/retry queue
  (`scheduleNext`, `inFlightRef`, `filesRef`, retry-once-then-`FAILED_FINAL`)
  is unchanged.
- Add `frontend/src/lib/upload.test.ts` covering `validateImageFile` (valid
  type, unsupported type, HEIC on Safari vs. non-Safari) and
  `uploadImageFile` (happy path through the mocked pipeline, HEIC conversion
  branch, propagating a validation error).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

`frontend-structure` — adds a requirement describing the new
`frontend/src/lib/upload.ts` (+ test) location and its role as the shared
upload pipeline used by `app-shell` and `features/upload`. This is a
zero-functional-change extraction; no behavior changes.

## Impact

- New files: `frontend/src/lib/upload.ts`, `frontend/src/lib/upload.test.ts`.
- `frontend/src/app-shell/lib/dragHandlers.ts`,
  `frontend/src/features/upload/components/UploadModal.tsx`, and
  `frontend/src/features/upload/components/BatchUploadModal.tsx` shrink —
  each loses its inline `ACCEPTED_TYPES`/`fileBaseName`/HEIC-check and
  pipeline body, replaced by calls into `lib/upload.ts`.
- Existing test files (`UploadModal.test.tsx`, `BatchUploadModal.test.tsx`,
  `dragHandlers.test.ts` if present) are adapted to mock `lib/upload.ts`
  instead of `lib/images.ts`/`lib/thumbnail.ts` directly, where they
  currently do so.
- No change to any component's props, `AppLayout`'s usage, or any
  user-visible behavior. The `Dropzone` extraction discussed for
  `UploadModal`/`BatchUploadModal` is explicitly deferred to a future,
  separate change.
- No backend, API, database, or browser-extension changes. No new
  dependencies.
