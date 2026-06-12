## 1. Create the shared upload module

- [x] 1.1 Create `frontend/src/lib/upload.ts` exporting `FileValidationError`, `validateImageFile`, `fileBaseName`, `UploadImageFileParams`, and `uploadImageFile` per design.md D1. `uploadImageFile` ports the pipeline (HEIC conversion → `initiateUpload` → `generateThumbnail` → `Promise.all([putToR2 ×2])` → `completeUpload`) verbatim from the existing three copies; `title` defaults to `fileBaseName(params.file.name)` when omitted.
- [x] 1.2 Add `frontend/src/lib/upload.test.ts` covering `validateImageFile` (accepted type, unsupported type, HEIC on Safari vs. non-Safari) and `uploadImageFile` (happy path, HEIC conversion branch, webp/avif accept, R2/initiate/complete error propagation, default-title-from-filename), migrating the equivalent scenarios currently in `dragHandlers.test.ts`'s `handleFileAutoUpload` block per design.md D3.

## 2. Migrate dragHandlers

- [x] 2.1 Update `frontend/src/app-shell/lib/dragHandlers.ts`'s `handleFileAutoUpload` to call `validateImageFile` (throwing `Error(err)` on `'unsupported_type'`/`'heic_safari_only'` as today) then `uploadImageFile`, followed by the existing `getImage(getToken, result.image_id)` call. Remove its local `ACCEPTED_TYPES` and `fileBaseName`.
- [x] 2.2 Update `dragHandlers.test.ts`'s `handleFileAutoUpload` block per design.md D3: mock `@/lib/upload` instead of `@/lib/images`/`@/lib/thumbnail`/`@/lib/browser`, keeping a couple of integration-style scenarios (success returns full `ImageDetail` via `getImage`; unsupported type rejects without calling `uploadImageFile`).

## 3. Migrate UploadModal

- [x] 3.1 Update `UploadModal.tsx`'s `handleFile` to use `validateImageFile` for its `typeError` checks (same messages as today), and `uploadMutation.mutationFn` to call `uploadImageFile({ file, folderId, title: resolvedTitle, description, sourceUrl })`. Remove its local `ACCEPTED_TYPES`, `fileBaseName`, and `isValidType`.
- [x] 3.2 Update `UploadModal.test.tsx` per design.md D3: replace `vi.mock('@/lib/images', ...)` / `vi.mock('@/lib/thumbnail', ...)` / `vi.mock('@/lib/browser', ...)` with `vi.mock('@/lib/upload', ...)`, adapting existing scenarios to assert against `uploadImageFile`/`validateImageFile` calls.

## 4. Migrate BatchUploadModal

- [x] 4.1 Update `BatchUploadModal.tsx`'s `makeBatchFile` to use `validateImageFile` for its `UNSUPPORTED` status (oversize check via `MAX_SIZE_BYTES` stays local), and `runUpload` to call `uploadImageFile({ file: batchFile.file, folderId, title: fileBaseName(batchFile.file.name) })`. Remove its local `ACCEPTED_TYPES` and `fileBaseName`.
- [x] 4.2 Update `BatchUploadModal.test.tsx` per design.md D3: replace `vi.mock('@/lib/images', ...)` / `vi.mock('@/lib/thumbnail', ...)` / `vi.mock('@/lib/browser', ...)` with `vi.mock('@/lib/upload', ...)`, adapting existing scenarios (including retry) to assert against `uploadImageFile`/`validateImageFile` calls.

## 5. Final checks

- [x] 5.1 Run `npm run build` and fix any resulting issues.
