## 1. Backend: accept client-supplied dimensions on CompleteUpload

- [x] 1.1 Add a `completeUploadRequest` DTO to `internal/handler/image_upload.go` with optional `width *int`, `height *int`, `file_size *int64` fields, bound from the JSON request body
- [x] 1.2 Update the `CompleteUpload` usecase signature to accept the client-supplied `width`, `height`, `file_size` values and pass them through to `CompleteUpload`
- [x] 1.3 In `imageUploadUsecase.CompleteUpload`, persist a value only when it is a positive integer (`> 0`); otherwise set the corresponding `domain.Image` field to `nil` (`NULL`)
- [x] 1.4 Remove `extractImageMetadata` and its now-unused imports (`stdimage "image"`, `_ "image/jpeg"`, `_ "image/png"`) from `internal/usecase/image_upload_usecase.go`

## 2. Backend: tests

- [x] 2.1 Update/add usecase tests in `internal/usecase/image_upload_usecase_test.go` covering: positive values persisted, non-positive values stored as NULL, omitted/absent values stored as NULL, and that `extractImageMetadata`/R2 fetch is no longer invoked during `CompleteUpload`
- [x] 2.2 Update/add handler tests in `internal/handler/image_upload_test.go` covering: request body with valid `width`/`height`/`file_size` is decoded and forwarded to the usecase, and a missing/empty request body completes successfully with all three values treated as absent

## 3. Backend: Bruno collection

- [x] 3.1 Update `bruno/images/complete-upload.bru` to include an example JSON request body with `width`, `height`, and `file_size`

## 4. Frontend: send dimensions and file size on completeUpload

- [x] 4.1 In `frontend/src/lib/thumbnail.ts`, change `generateThumbnail` (or add a sibling helper reusing its `createImageBitmap` decode) so the caller receives the thumbnail blob alongside the source bitmap's `width`/`height`
- [x] 4.2 In `frontend/src/lib/images.ts`, extend `completeUpload` to accept and send `width`, `height`, and `file_size` in the request body
- [x] 4.3 In `frontend/src/components/UploadModal.tsx`, capture `width`/`height` from the decode performed in 4.1 and `file_size` from the uploaded blob (`uploadBlob.size`), and pass them to `completeUpload`
- [x] 4.4 In `frontend/src/components/BatchUploadModal.tsx` and `frontend/src/lib/dragHandlers.ts`, apply the same change as 4.3 — both already call `generateThumbnail`/`completeUpload` for their own upload sequences and would otherwise permanently persist `NULL` dimensions now that the backend no longer derives them server-side

## 5. Frontend: tests and build

- [x] 5.1 Update existing upload-flow tests (or add new ones) asserting `completeUpload` is called with `width`, `height`, and `file_size` matching the decoded bitmap and uploaded blob, covering `UploadModal`, `BatchUploadModal`, and `dragHandlers`
- [x] 5.2 Run `npm run build` and fix any issues that arise

## 6. Extension: send dimensions and file size on complete

- [x] 6.1 In `extensions/src/background/index.ts`, change `generateThumbnail` (or restructure around it) so the caller can also obtain the decoded bitmap's `width`/`height` when `OffscreenCanvas` is available
- [x] 6.2 In `saveImage`, send a JSON body on `POST /images/:id/complete` containing `file_size` (always) and `width`/`height` (only when captured in 6.1)

## 7. Backend: lint

- [x] 7.1 Run `golangci-lint run` and fix any issues that arise
