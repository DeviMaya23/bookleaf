## Context

Three call sites implement the same "validate → convert HEIC if needed →
`initiateUpload` → `generateThumbnail` → `putToR2` ×2 → `completeUpload`"
sequence:

- `app-shell/lib/dragHandlers.ts`'s `handleFileAutoUpload` — validates first
  (throws `'unsupported_type'`/`'heic_safari_only'` Error), runs the
  pipeline, then calls `getImage(result.image_id)` for the full
  `ImageDetail` `RightPanel` needs.
- `UploadModal.tsx`'s `handleFile` validates eagerly on file select/drop
  (sets `typeError` state for inline display, doesn't throw), then later
  `uploadMutation.mutationFn` runs the pipeline on the already-validated
  file with `title`/`description`/`sourceUrl` from the form.
- `BatchUploadModal.tsx`'s `makeBatchFile` validates eagerly when files are
  added (marks `UNSUPPORTED`/`OVERSIZED` status, never queues them), then
  `runUpload` runs the pipeline on an already-validated file.

In all three, **validation already happens before the pipeline runs** —
there's no caller relying on the pipeline itself to reject an invalid file.

## Goals / Non-Goals

**Goals:**
- One implementation of the pipeline (`uploadImageFile`) and one
  implementation of the type/HEIC validity check (`validateImageFile`) in
  `frontend/src/lib/upload.ts`.
- Preserve each call site's current validation *timing* (eager-on-select for
  the modals, inline-first-step for `dragHandlers`) and current error
  surface (thrown `Error('unsupported_type')` /
  `Error('heic_safari_only')` for `dragHandlers`; non-throwing status/state
  for the modals).

**Non-Goals:**
- No `Dropzone` extraction (deferred, separate change).
- No change to `BatchUploadModal`'s queue/retry/concurrency logic
  (`scheduleNext`, `inFlightRef`, `filesRef`) beyond what `runUpload`'s body
  delegates to `uploadImageFile`.
- No change to `MAX_SIZE_BYTES`/oversize handling — stays batch-local.
- No change to any component prop, query key, or user-visible behavior.

## Decisions

### D1: `lib/upload.ts` exports

```ts
export type FileValidationError = 'unsupported_type' | 'heic_safari_only'

export function validateImageFile(file: File): FileValidationError | null

export function fileBaseName(name: string): string

export interface UploadImageFileParams {
  file: File
  folderId: string | null
  title?: string
  description?: string
  sourceUrl?: string
}

export async function uploadImageFile(
  getToken: GetToken,
  params: UploadImageFileParams,
): Promise<CompleteUploadResult>
```

- `title` defaults to `fileBaseName(params.file.name)` when omitted —
  matches `dragHandlers`' and `BatchUploadModal`'s current behavior, and
  lets `UploadModal` pass its resolved title explicitly.
- Returns `CompleteUploadResult` (`{ image_id, suggested_folder_name }`),
  the common denominator across all three callers.

### D2: `uploadImageFile` does not call `validateImageFile`

Since every call site already validates before reaching the pipeline (see
Context), `uploadImageFile` assumes a valid file and does not re-validate.
This keeps its failure modes to one kind (pipeline/network errors) instead
of mixing in a validation-error string union, and avoids the question of
"what does `uploadImageFile` do with a validation failure" (throw? return a
discriminated result?) — that's each caller's existing job via
`validateImageFile`.

- `dragHandlers.handleFileAutoUpload` becomes:
  ```ts
  const err = validateImageFile(file)
  if (err) throw new Error(err)
  // ...convert HEIC if needed stays here, OR moves into uploadImageFile —
  // see note below
  const result = await uploadImageFile(getToken, { file, folderId })
  return getImage(getToken, result.image_id)
  ```
  HEIC→JPEG conversion is part of `uploadImageFile` (it's pipeline work, not
  validation), so `handleFileAutoUpload` doesn't need to handle it itself.
- `UploadModal.handleFile` keeps its current `validateImageFile`-based
  `typeError` logic (message text unchanged); `uploadMutation.mutationFn`
  calls `uploadImageFile` directly (file is already known-valid by the time
  `mutate()` can be called, since the upload button is disabled without a
  valid `file`).
- `BatchUploadModal.makeBatchFile` keeps its current
  `validateImageFile`-based `UNSUPPORTED` status logic (plus its own
  oversize check); `runUpload` calls `uploadImageFile` directly.

### D3: Test migration

Pipeline-shaped scenarios currently in `dragHandlers.test.ts`'s
`describe('handleFileAutoUpload', ...)` (HEIC conversion, webp/avif accept,
thumbnail generation, R2/initiate/complete error propagation,
validation-error throwing) move to `frontend/src/lib/upload.test.ts`,
mocking `@/lib/images` and `@/lib/thumbnail`/`@/lib/browser` the same way
`dragHandlers.test.ts` does today.

`dragHandlers.test.ts`'s `handleFileAutoUpload` block shrinks to a couple of
integration-style scenarios confirming it wires `validateImageFile` →
`uploadImageFile` → `getImage` correctly (e.g. "returns full `ImageDetail`
on success", "rejects unsupported type without calling `uploadImageFile`"),
now mocking `@/lib/upload` instead of `@/lib/images`/`@/lib/thumbnail`/`@/lib/browser`.

`UploadModal.test.tsx` and `BatchUploadModal.test.tsx` switch their
`vi.mock('@/lib/images', ...)` + `vi.mock('@/lib/thumbnail', ...)` +
`vi.mock('@/lib/browser', ...)` to a single `vi.mock('@/lib/upload', ...)`
mocking `validateImageFile`/`fileBaseName`/`uploadImageFile`. Existing
scenarios (success, retry, oversize/unsupported display, HEIC-on-non-Safari
message) are preserved, just asserting against `uploadImageFile` calls
instead of the individual `initiateUpload`/`putToR2`/`completeUpload` calls.

### D4: Sequencing

Four independent steps, each leaving the app in a working state:
1. Add `lib/upload.ts` + `lib/upload.test.ts` (new file, nothing depends on
   it yet).
2. Migrate `dragHandlers.ts` + its test.
3. Migrate `UploadModal.tsx` + its test.
4. Migrate `BatchUploadModal.tsx` + its test.

Steps 2-4 are independent of each other and can land in any order.

## Risks / Trade-offs

- **[Trade-off]** `validateImageFile` is called separately by each of the
  three sites rather than being baked into `uploadImageFile`, so a future
  caller could forget to validate before uploading. → **Mitigation**: this
  matches today's behavior exactly (no caller currently skips validation);
  documented in D2 as a deliberate choice, not an oversight.
- **[Risk]** `BatchUploadModal`'s retry path (`FAILED_FINAL` → `PENDING` →
  `runUpload` again) must continue to retry the *pipeline only*, not
  re-validate — `uploadImageFile` not validating internally makes this
  automatic, but worth confirming in the retry test scenario.
