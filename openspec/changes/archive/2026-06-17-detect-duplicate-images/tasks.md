## 1. Domain & Migration

- [x] 1.1 Add `PHash *string` field to `domain.Image` with GORM tags `column:phash;type:bit(64)`
- [x] 1.2 Write `golang-migrate` SQL migration adding `phash bit(64) NULL` to the `images` table (up and down)

## 2. Backend – Repository

- [x] 2.1 Add `go get github.com/corona10/goimagehash` to `backend/go.mod`
- [x] 2.2 Add `FindDuplicates(ctx context.Context, userID string, phash string, excludeID uuid.UUID, threshold int) ([]*domain.Image, error)` to the `UploadImageRepository` interface in `usecase/image_upload_usecase.go`
- [x] 2.3 Implement `FindDuplicates` in `repository/image_repository.go` using `bit_count(phash # $phash::bit(64)) <= $threshold` with `user_id`, `deleted_at IS NULL`, `phash IS NOT NULL`, and `id != $excludeID` filters
- [x] 2.4 Add `ListUnhashed(ctx context.Context, limit int) ([]*domain.Image, error)` and `UpdatePHash(ctx context.Context, id uuid.UUID, phash string) error` to a new `ImageHashRepository` interface in a new `usecase/image_hash_repository.go` file
- [x] 2.5 Implement `ListUnhashed` and `UpdatePHash` in `repository/image_repository.go`

## 3. Backend – CompleteUpload Usecase & Handler

- [x] 3.1 Add `PHash *string` param to `CompleteUpload` in `usecase/image_upload_usecase.go`; store it on the `Image` struct during the transaction
- [x] 3.2 Add `Duplicates []*domain.Image` to `CompleteUploadResult`; after committing the image, call `FindDuplicates` with threshold 10 when `PHash` is non-nil and populate the field (empty slice otherwise)
- [x] 3.3 Write unit tests for `CompleteUpload` in `usecase/image_upload_usecase_test.go`:
  - Phash provided, duplicates found → result includes duplicate entries with correct id/title/thumbnail_path
  - Phash provided, no duplicates → result has empty duplicates slice
  - Phash absent → `FindDuplicates` not called, result has empty duplicates slice
- [x] 3.4 Add `PHash *string` to `completeUploadRequest` in `handler/image_upload.go`; pass it through to the usecase
- [x] 3.5 Add `Duplicates` array to `completeUploadResponse` in `handler/image_upload.go`; map each `domain.Image` to `{id, title, thumbnail_path}`

## 4. Backend – Periodic Backfill

- [x] 4.1 Add `BackfillPhash(ctx context.Context, batchSize int) error` method to `imageUploadUsecase`, injecting `StorageService` (already present): fetch `ListUnhashed(ctx, batchSize)`, for each image download its thumbnail (fall back to `R2Path` if `ThumbnailPath` is nil), compute `goimagehash.PerceptionHash`, format as 64-character binary string, call `UpdatePHash`; skip and warn on fetch/decode error
- [x] 4.2 Add `BackfillPhashArgs` and `BackfillPhashWorker` to `worker/periodic.go` following the existing worker pattern; the worker calls `usecase.BackfillPhash(ctx, 20)`
- [x] 4.3 Register `BackfillPhashWorker` with `river.AddWorker` and add a `river.NewPeriodicJob` entry (interval: 5 minutes) in `cmd/server/main.go`
- [x] 4.4 Add a Bruno file for `POST /images/:id/complete` updated with the `phash` request field and `duplicates` response example

## 5. Frontend – pHash Computation

- [x] 5.1 Implement `computePHash(canvas: HTMLCanvasElement): string` in `src/lib/phash.ts`: draw to 64×64 greyscale canvas (matching goimagehash's input size), read pixel data using Rec. 601 luminance, apply separable 2D DCT, take top-left 8×8 coefficients, compare each to the median of all 64 values, return a 64-character binary string
- [x] 5.2 Update `generateThumbnail` (or the thumbnail step in `uploadImageFile`) in `src/lib/upload.ts` to call `computePHash` on the thumbnail canvas and return the hash alongside `blob`, `width`, and `height`
- [x] 5.3 Update `CompleteUploadResult` type in `src/lib/images.ts` to include `duplicates: Array<{ id: string; title: string; thumbnail_path: string | null }>`
- [x] 5.4 Update `completeUpload` in `src/lib/images.ts` to accept and send `phash` in the request body
- [x] 5.5 Update `uploadImageFile` in `src/lib/upload.ts` to pass `phash` to `completeUpload` and return the full `CompleteUploadResult` including `duplicates`

## 6. Frontend – Upload Modal Duplicate Toast

- [x] 6.1 In the single-file upload success flow (`UploadModal` or the drag-and-drop path in `AppLayout`), check if `result.duplicates` is non-empty after `uploadImageFile` resolves; if so, show a warning toast naming `result.duplicates[0].title` in addition to the standard success toast

## 7. Frontend – Batch Modal Duplicate Annotation

- [x] 7.1 Add optional `duplicateOf?: string` field (first duplicate's title) to the `BatchFile` type in `BatchUploadModal.tsx`
- [x] 7.2 In `runUpload`, after a successful `uploadImageFile`, check `result.duplicates`; if non-empty, call `updateFile` with `{ status: 'SUCCESS', duplicateOf: result.duplicates[0].title }`
- [x] 7.3 Update `StatusCell` to render a warning icon (e.g. `AlertTriangle`) next to the success icon when `batchFile.duplicateOf` is set; wrap it in a tooltip showing the duplicate's title

## 8. Quality

- [x] 8.1 Run `golangci-lint run ./...` in `backend/` and fix all reported issues
- [x] 8.2 Run `npm run build && npm run lint` in `frontend/` and fix all reported issues
