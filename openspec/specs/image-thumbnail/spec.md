## ADDED Requirements

### Requirement: ThumbnailService Interface

The system SHALL define a `ThumbnailService` interface in `internal/thumbnail/` with a single `Generate` method. The image usecase SHALL depend on this interface.

Method:
- `Generate(ctx context.Context, src io.Reader) (io.Reader, error)` — reads the source image, returns a resized JPEG reader

#### Scenario: Interface is satisfied by imaging implementation

- **WHEN** the Go package is compiled
- **THEN** the concrete thumbnail implementation satisfies `ThumbnailService` without compilation errors

---

### Requirement: Thumbnail Generation

The system SHALL implement `ThumbnailService` using the `disintegration/imaging` library. The generated thumbnail SHALL:

- Fit within a 600×600 pixel bounding box while preserving the original aspect ratio (no distortion)
- Be encoded as JPEG regardless of the source image format
- Use `imaging.Lanczos` as the resampling filter

#### Scenario: Landscape image is resized to fit 600×600

- **WHEN** a 1200×600 image is passed to `Generate`
- **THEN** the output JPEG dimensions are 600×300

#### Scenario: Portrait image is resized to fit 600×600

- **WHEN** a 600×1200 image is passed to `Generate`
- **THEN** the output JPEG dimensions are 300×600

#### Scenario: Image already within bounds is not upscaled

- **WHEN** a 200×100 image is passed to `Generate`
- **THEN** the output JPEG dimensions are 200×100

---

### Requirement: Async Thumbnail Storage

After `POST /images/:id/complete` is called, the system SHALL fetch the original image from R2, generate a thumbnail, store it back to R2, and update `thumbnail_path` on the `Image` record — all in a background goroutine.

Steps (in the goroutine):
1. Call `StorageService.GetObject` to fetch the original image bytes
2. Call `ThumbnailService.Generate` to produce the thumbnail
3. Call `StorageService.PutObject` to store the thumbnail at `users/{kindeID}/thumbnails/{imageID}.jpg` with content type `image/jpeg`
4. Update `Image.ThumbnailPath` in the database

If any step fails, the error SHALL be logged and the goroutine exits without panicking. The HTTP response for `/complete` is NOT blocked on these steps.

#### Scenario: Successful thumbnail generation updates ThumbnailPath

- **WHEN** the goroutine completes all four steps without error
- **THEN** the `Image` record in the database has `thumbnail_path` set to `users/{kindeID}/thumbnails/{imageID}.jpg`

#### Scenario: Thumbnail generation failure is logged and does not crash

- **WHEN** `StorageService.GetObject` returns an error
- **THEN** the error is logged
- **AND** the goroutine exits without panicking
- **AND** `thumbnail_path` remains nil
