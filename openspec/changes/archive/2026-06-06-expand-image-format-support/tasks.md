## 1. Backend: StorageService — HeadObject

- [x] 1.1 Add `HeadObject(ctx context.Context, key string) (bool, error)` to `StorageService` interface in `internal/usecase/image_storage.go`
- [x] 1.2 Implement `HeadObject` on the R2 storage client in `internal/storage/`
- [x] 1.3 Add `HeadObject` to `mockStorageService` in `internal/usecase/image_upload_usecase_test.go`

## 2. Backend: Format mappings

- [x] 2.1 Add `image/avif → ".avif"` case to `MimeTypeToExt` in `internal/storage/storage.go`
- [x] 2.2 Add `image/avif → "avif"` case to `downloadFileExtension` in `internal/usecase/image_usecase.go`

## 3. Backend: InitiateUpload — return thumbnail_upload_url

- [x] 3.1 Add `ThumbnailUploadURL string` and `ThumbnailKey string` to `UploadInitResult` struct
- [x] 3.2 Compute `thumbnailKey = "users/{userID}/thumbnails/{imageID}.jpg"` in `InitiateUpload` and generate a presigned PUT URL for it (`image/jpeg`, same TTL as original)
- [x] 3.3 Populate `ThumbnailUploadURL` and `ThumbnailKey` on the returned `UploadInitResult`
- [x] 3.4 Expose `thumbnail_upload_url` in the handler's JSON response for `POST /images`
- [x] 3.5 Update unit test for `InitiateUpload` success to assert `ThumbnailUploadURL` is non-empty in the result

## 4. Backend: CompleteUpload — remove preflight, add HEAD check

- [x] 4.1 Remove the synchronous `thumbnails.Generate` preflight call (and its associated `rawBytes` usage) from `CompleteUpload`; `extractImageMetadata` remains but only rawBytes is no longer passed to thumbnails
- [x] 4.2 Call `store.HeadObject` on `thumbnailKey` after `extractImageMetadata`, before the DB transaction
- [x] 4.3 Build `domain.Image` with `ThumbnailPath` set to `thumbnailKey` when `HeadObject` returns `true`; leave `ThumbnailPath` nil when `false`
- [x] 4.4 Skip `ThumbnailUploadArgs` enqueue when `HeadObject` returned `true`; keep the enqueue for the `false` path (fallback for extension)
- [x] 4.5 Delete the unit test scenarios that tested the removed synchronous preflight failure cases
- [x] 4.6 Add unit test: `CompleteUpload` — thumbnail present path (HEAD true → `thumbnail_path` set, no worker enqueued)
- [x] 4.7 Add unit test: `CompleteUpload` — thumbnail absent path (HEAD false → `thumbnail_path` nil, worker enqueued)

## 5. Backend: Bruno file

- [x] 5.1 Update `bruno/images/initiate-upload.bru` to show `thumbnail_upload_url` in the response body example

## 6. Frontend: Core utilities

- [x] 6.1 Update `InitiateUploadResult` type in `images.ts` to include `thumbnail_upload_url: string`
- [x] 6.2 Update `putToR2` signature to accept `Blob | File` (thumbnail is a `Blob`; `Blob.type` is used for `Content-Type`)
- [x] 6.3 Create `src/lib/thumbnail.ts` exporting:
  - `generateThumbnail(source: Blob): Promise<Blob>` — decodes via `createImageBitmap`, draws onto canvas fitting 600×600, exports as `image/jpeg` at quality `0.9`
  - `convertHeicToJpeg(file: File): Promise<Blob>` — decodes via `createImageBitmap`, draws onto canvas at full natural dimensions, exports as `image/jpeg` at quality `0.93`
- [x] 6.4 Create `src/lib/browser.ts` exporting `isSafari(): boolean` (user-agent check)

## 7. Frontend: Format acceptance and HEIC detection

- [x] 7.1 Update `ACCEPTED_TYPES` in `dragHandlers.ts` to `['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/avif', 'image/heic']`
- [x] 7.2 In `handleFileAutoUpload`, after the type check: if the file is `image/heic` and `!isSafari()`, throw `'unsupported_type'`
- [x] 7.3 In `handleFileAutoUpload`, if the file is `image/heic`, call `convertHeicToJpeg` to get an upload blob and use `'image/jpeg'` as the MIME type; otherwise use the original file

## 8. Frontend: 4-step upload flow

- [x] 8.1 Update `handleFileAutoUpload` in `dragHandlers.ts`: after initiating, run `generateThumbnail` and `putToR2` (original) in parallel via `Promise.all`, then PUT thumbnail to `thumbnail_upload_url`, then call `completeUpload`
- [x] 8.2 Update `UploadModal.tsx` inline upload sequence to the same 4-step flow (parallel PUTs, then complete)
- [x] 8.3 Update `BatchUploadModal.tsx` per-file upload sequence to the same 4-step flow

## 9. Frontend: Unit tests

- [x] 9.1 Update `dragHandlers.test.ts`: add scenarios for webp/avif accepted, heic rejected on non-Safari, heic accepted on Safari with conversion; update existing success test to include thumbnail PUT
- [x] 9.2 Update `UploadModal.test.tsx` mocks and assertions to reflect the 4-step flow (`thumbnail_upload_url` in initiate mock, thumbnail PUT mock, assertions on call order)
- [x] 9.3 Update `BatchUploadModal.test.tsx` mocks and assertions for the same 4-step flow
