# Proposal

## Why

Image uploads have no enforced size ceiling — a user can upload arbitrarily large files, wasting R2 storage and degrading performance for themselves and the system. Adding a 50 MB per-image limit closes this gap with client-side feedback on every upload surface and a signed R2 enforcement layer that holds even if someone bypasses the API and uses a presigned URL directly.

## What Changes

- `POST /images` (initiate upload) accepts a new required `file_size` field and returns `413` if the value exceeds 50 MB
- `POST /images/:id/complete` adds a belt-and-suspenders `file_size ≤ 50 MB` validation and returns `413` if exceeded
- `StorageService.GeneratePresignedPutURL` gains a `contentLength int64` parameter; when non-zero, the value is signed into the presigned PUT URL so R2 rejects any PUT whose `Content-Length` header does not match
- `generateThumbnail` in the extension background does not sign ContentLength (size unknown at initiate time for the thumbnail)
- `validateImageFile()` in `frontend/src/lib/upload.ts` gains an `'oversized'` error variant and checks `file.size > 50 MB`; the 50 MB constant is exported from that module
- All frontend upload surfaces (single-file `UploadModal`, drag-to-main-area auto-upload, clipboard paste → `UploadModal`) reject oversized files with an inline error before initiating
- `BatchUploadModal` already validates 50 MB; its inline constant is replaced with the shared one from `upload.ts`, and it sends `file_size` in the `POST /images` initiate body (currently only sent in complete)
- The extension's `persistImage()` checks `blob.size > 50 MB` before attempting to save and sends the size in the `POST /images` initiate body; failure shows a "File too large" in-page toast

## Capabilities

### New Capabilities

- `upload-file-size-limit`: Cross-cutting 50 MB per-image enforcement — backend validation at both initiate and complete, frontend shared constant and `validateImageFile` extension, single-file upload surfaces (modal, auto-upload, paste), and extension save-before-upload guard

### Modified Capabilities

- `r2-storage`: `StorageService.GeneratePresignedPutURL` signature gains `contentLength int64`; the R2 implementation sets `ContentLength` on `PutObjectInput` when non-zero, causing the AWS SDK to include `content-length` in `X-Amz-SignedHeaders`
- `extension-save-image`: Save flow adds a size guard in `persistImage()` before the upload sequence; oversized blobs are rejected with a toast and no API call is made; `file_size` is added to the `POST /images` initiate request body
- `fe-batch-upload`: `file_size` is added to the `POST /images` initiate request body (currently only sent in complete); the inline `MAX_SIZE_BYTES` constant is replaced with the shared one from `upload.ts`

## Impact

- **Backend**: `handler/image_upload.go`, `usecase/image_upload_usecase.go`, `storage/r2.go`, `storage/storage.go` (interface)
- **Frontend**: `lib/upload.ts`, `features/upload/components/UploadModal.tsx`, `features/upload/components/BatchUploadModal.tsx`, `app-shell/AppLayout.tsx`, `app-shell/lib/dragHandlers.ts`
- **Extension**: `extensions/src/background/index.ts`
- **API contract**: `POST /images` request body gains required `file_size`; `POST /images/:id/complete` gets a new `413` response; `StorageService` interface changes (internal, compile-time only)
- **Bruno file**: `POST /images` request body needs `file_size` added
