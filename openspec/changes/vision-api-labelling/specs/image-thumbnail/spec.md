## MODIFIED Requirements

### Requirement: Async Thumbnail Storage

After `POST /images/:id/complete` is called, the system SHALL fetch the original image from R2 and generate a thumbnail synchronously within the `CompleteUpload` usecase method. The R2 upload and database update SHALL then be handed off to a background goroutine.

Synchronous steps (blocking the HTTP response):
1. Call `StorageService.GetObject` to fetch the original image bytes
2. Call `ThumbnailService.Generate` to produce the thumbnail; buffer the result as `[]byte`

If either synchronous step fails, the error SHALL be logged, `CompleteUploadResult.Warning` SHALL be set, and the method SHALL return without calling the goroutine. The HTTP response is still `200 OK`.

Goroutine steps (non-blocking):
3. Call `StorageService.PutObject` to store the pre-generated thumbnail bytes at `users/{kindeID}/thumbnails/{imageID}.jpg` with content type `image/jpeg`
4. Update `Image.ThumbnailPath` in the database

If either goroutine step fails, the error SHALL be logged and the goroutine SHALL exit without panicking. The HTTP response is NOT blocked on goroutine steps.

#### Scenario: Successful thumbnail flow updates ThumbnailPath

- **WHEN** GetObject, Generate, PutObject, and UpdateThumbnailPath all succeed
- **THEN** the `Image` record in the database has `thumbnail_path` set to `users/{kindeID}/thumbnails/{imageID}.jpg`

#### Scenario: GetObject failure sets warning and skips goroutine

- **WHEN** `StorageService.GetObject` returns an error
- **THEN** the error is logged
- **AND** `CompleteUploadResult.Warning` is set to a non-empty string
- **AND** no goroutine is launched
- **AND** `thumbnail_path` remains nil

#### Scenario: Goroutine PutObject failure is logged and does not crash

- **WHEN** GetObject and Generate succeed but `StorageService.PutObject` returns an error
- **THEN** the error is logged
- **AND** the goroutine exits without panicking
- **AND** `thumbnail_path` remains nil
