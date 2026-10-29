# Tasks

## 1. Storage layer — interface and R2 implementation

- [x] 1.1 Add `contentLength int64` parameter to `GeneratePresignedPutURL` in `StorageService` interface (`internal/storage/storage.go`) and update the `r2Storage` implementation to conditionally set `ContentLength` on `PutObjectInput` when the value is greater than zero; verify the package compiles without errors

## 2. Backend usecase — initiate and complete validation

- [x] 2.1 Add `ErrMissingFileSize` and `ErrFileSizeTooLarge` sentinel errors to `image_upload_usecase.go`; add a `MaxImageFileSizeBytes` constant set to `52_428_800`

- [x] 2.2 Add `FileSize int64` to `initiateImageUploadRequest` in `handler/image_upload.go` and update `InitiateUpload` in the usecase to validate: return `ErrMissingFileSize` when `file_size` is absent or zero, return `ErrFileSizeTooLarge` when it exceeds `MaxImageFileSizeBytes`, and pass `fileSize` as `contentLength` to `GeneratePresignedPutURL` for the main image URL only (thumbnail URL passes zero)

- [x] 2.3 Update `handler/image_upload.go` `InitiateUpload` to map the new errors: `ErrMissingFileSize` → `400`, `ErrFileSizeTooLarge` → `413`

- [x] 2.4 Update `CompleteUpload` in the usecase to return `ErrFileSizeTooLarge` when `file_size` exceeds `MaxImageFileSizeBytes`; update the handler to map `ErrFileSizeTooLarge` → `413` on the complete route

- [x] 2.5 Write usecase unit tests in `image_upload_usecase_test.go` covering: `InitiateUpload` returns `ErrMissingFileSize` when `file_size` is zero, returns `ErrFileSizeTooLarge` when `file_size` exceeds the limit (assert the specific error), succeeds and calls `GeneratePresignedPutURL` with the correct `contentLength` when `file_size` is valid; `CompleteUpload` returns `ErrFileSizeTooLarge` when `file_size` exceeds the limit (assert the specific error)

- [x] 2.6 Write handler unit tests in `image_upload_test.go` covering: `InitiateUpload` returns `400` for missing `file_size`, `413` for oversized `file_size`; `CompleteUpload` returns `413` for oversized `file_size`

## 3. Frontend — shared validation library

- [x] 3.1 In `frontend/src/lib/upload.ts`: export `MAX_FILE_SIZE_BYTES = 52_428_800`, add `'oversized'` to the `FileValidationError` union type, and update `validateImageFile()` to return `'oversized'` when `file.size > MAX_FILE_SIZE_BYTES`

- [x] 3.2 Update `frontend/src/lib/images.ts` `initiateUpload()` to accept and send `fileSize: number` in the `POST /images` request body

- [x] 3.3 Update `upload.test.ts` to cover: `validateImageFile` returns `'oversized'` for a file over the limit, returns `null` for a file exactly at the limit

## 4. Frontend — UploadModal (single file)

- [x] 4.1 Update `UploadModal.handleFile()` to handle the `'oversized'` error from `validateImageFile()` with an inline error message; verify the upload button remains disabled while the error is shown

- [x] 4.2 Update `uploadImageFile()` in `upload.ts` to pass `uploadBlob.size` as `fileSize` to `initiateUpload()`

- [x] 4.3 Update `UploadModal` tests to cover: oversized file shows inline error and disables upload button, selecting a valid file after an oversized one clears the error

## 5. Frontend — auto-upload and BatchUploadModal

- [x] 5.1 Update `handleFileAutoUpload()` in `dragHandlers.ts` to check `validateImageFile()` result for `'oversized'` and throw with message `'oversized'`; update AppLayout's `handleMainDrop` error handler to show a size-specific toast for the `'oversized'` case

- [x] 5.2 In `BatchUploadModal.tsx`, replace the inline `MAX_SIZE_BYTES` constant with the imported `MAX_FILE_SIZE_BYTES` from `upload.ts`; update `makeBatchFile()` to derive `'OVERSIZED'` status from the `validateImageFile()` result rather than a separate size check

- [x] 5.3 Update `BatchUploadModal` to pass `file.size` as `fileSize` to `uploadImageFile()`; verify this flows through to `initiateUpload()` via the updated `upload.ts`

- [x] 5.4 Update `dragHandlers.test.ts` to cover: `handleFileAutoUpload` throws `'oversized'` for a file over the limit; update `BatchUploadModal.test.tsx` to cover: oversized file receives `'OVERSIZED'` status badge via `validateImageFile()`

## 6. Extension

- [x] 6.1 Add a size check in `persistImage()` in `extensions/src/background/index.ts`: if `blob.size > 52_428_800`, send an error toast (`"File too large."` / `"Images must be under 50 MB."`) and return without calling `saveImage()`

- [x] 6.2 Update `saveImage()` to include `file_size: blob.size` in the `POST /images` request body

- [x] 6.3 Update `background/index.test.ts` to cover: `persistImage` sends an error toast and makes no API call when the blob exceeds 50 MB; `saveImage` includes `file_size` in the initiate request body

## 7. Bruno, lint, and build

- [x] 7.1 Update the Bruno request file for `POST /images` to include `file_size` in the request body

- [x] 7.2 Run `golangci-lint run ./...` from `backend/` and fix any issues introduced by this change

- [x] 7.3 Run `npm run build` and `npm run lint` from `frontend/` and fix any issues; run the same from `extensions/` and fix any issues

## Workflow follow-up

- Archive the change once review is complete.
