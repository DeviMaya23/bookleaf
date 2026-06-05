## MODIFIED Requirements

### Requirement: Async Thumbnail Storage

After `POST /images/:id/complete` is called, the system SHALL fetch the original image from R2 and generate a thumbnail synchronously within the `CompleteUpload` usecase method. The R2 upload and database update SHALL be handed off to a River job (`ThumbnailUploadJob`) immediately after the image DB write succeeds.

Synchronous steps (blocking the HTTP response):
1. Call `StorageService.GetObject` to fetch the original image bytes
2. Call `ThumbnailService.Generate` to produce the thumbnail; buffer the result as `[]byte`

If either synchronous step fails, the error SHALL be logged and `CompleteUpload` SHALL return an error. The HTTP response SHALL be non-2xx. The River job SHALL NOT be inserted.

River job steps (non-blocking, executed by `ThumbnailUploadWorker`):
3. Call `StorageService.GetObject` to re-fetch the original image bytes from R2
4. Call `ThumbnailService.Generate` to re-produce the thumbnail
5. Call `StorageService.PutObject` to store the thumbnail at `users/{kindeID}/thumbnails/{imageID}.jpg` with content type `image/jpeg`
6. Update `Image.ThumbnailPath` in the database

If any job step fails, River retries the job up to 5 attempts with exponential backoff. The thumbnail bytes are NOT passed via job args; the worker re-fetches and re-generates from R2 on each attempt.

#### Scenario: Successful thumbnail flow updates ThumbnailPath

- **WHEN** all job steps succeed
- **THEN** the `Image` record has `thumbnail_path` set to `users/{kindeID}/thumbnails/{imageID}.jpg`

#### Scenario: GetObject failure in synchronous phase returns error

- **WHEN** `StorageService.GetObject` fails during `CompleteUpload`
- **THEN** `CompleteUpload` returns an error
- **AND** no River job is inserted
- **AND** `thumbnail_path` remains nil

#### Scenario: Generate failure in synchronous phase returns error

- **WHEN** `ThumbnailService.Generate` fails during `CompleteUpload`
- **THEN** `CompleteUpload` returns an error
- **AND** no River job is inserted

#### Scenario: Job failure triggers retry

- **WHEN** the `ThumbnailUploadWorker` returns an error on an attempt
- **THEN** River retries the job up to 5 attempts total before discarding

## ADDED Requirements

### Requirement: ProcessThumbnailUpload Usecase Method

The system SHALL add `ProcessThumbnailUpload(ctx context.Context, imageID uuid.UUID, r2Path, thumbnailKey string) error` to `imageUploadUsecase`. This is the method called by `ThumbnailUploadWorker` on each attempt.

The method SHALL:
1. Call `StorageService.GetObject(ctx, r2Path)` to fetch the original image
2. Call `ThumbnailService.Generate(ctx, src)` to produce the thumbnail
3. Call `StorageService.PutObject(ctx, thumbnailKey, bytes, "image/jpeg")` to upload it
4. Call `ImageRepository.UpdateThumbnailPath(ctx, imageID, thumbnailKey)` to persist the path

Any failure at any step SHALL return a non-nil error so River can retry.

#### Scenario: All steps succeed

- **WHEN** `ProcessThumbnailUpload` is called and all steps complete without error
- **THEN** the image record has `thumbnail_path` set to `thumbnailKey`
- **AND** nil is returned

#### Scenario: Any step failure returns error

- **WHEN** any step in `ProcessThumbnailUpload` returns an error
- **THEN** `ProcessThumbnailUpload` returns a non-nil error
