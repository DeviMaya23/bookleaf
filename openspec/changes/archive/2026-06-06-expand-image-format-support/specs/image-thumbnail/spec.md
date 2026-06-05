## MODIFIED Requirements

### Requirement: Async Thumbnail Storage

After `POST /images/:id/complete` is called, the system SHALL check whether a thumbnail has already been uploaded to R2 by the client. If the thumbnail exists, `thumbnail_path` SHALL be set on the image record at creation time and no thumbnail worker job SHALL be enqueued. If the thumbnail does not exist, `thumbnail_path` SHALL remain null and a `ThumbnailUploadJob` SHALL be enqueued as a fallback.

The check SHALL be a `StorageService.HeadObject` call on `users/{kindeID}/thumbnails/{imageID}.jpg` performed during `CompleteUpload`, after the DB transaction succeeds.

The synchronous preflight (`ThumbnailService.Generate` called before the DB write) is removed. `CompleteUpload` SHALL NOT call `ThumbnailService.Generate` or `StorageService.GetObject` for thumbnail purposes.

Thumbnail-present path (client uploaded thumbnail):
1. `StorageService.HeadObject` on the thumbnail key — found
2. Image record created with `thumbnail_path` set to the thumbnail key
3. No `ThumbnailUploadJob` inserted

Thumbnail-absent path (fallback, e.g. extension):
1. `StorageService.HeadObject` on the thumbnail key — not found
2. Image record created with `thumbnail_path` = null
3. `ThumbnailUploadJob` inserted as before

River job steps (unchanged, non-blocking, executed by `ThumbnailUploadWorker`):
1. Call `StorageService.GetObject` to fetch the original image bytes from R2
2. Call `ThumbnailService.Generate` to produce the thumbnail
3. Call `StorageService.PutObject` to store the thumbnail at `users/{kindeID}/thumbnails/{imageID}.jpg`
4. Call `ImageRepository.UpdateThumbnailPath` to persist the path

#### Scenario: Client-uploaded thumbnail skips the worker

- **WHEN** the client uploads a JPEG thumbnail to `thumbnail_upload_url` before calling `/complete`
- **AND** `HeadObject` on the thumbnail key returns found
- **THEN** the image record is created with `thumbnail_path` set
- **AND** no `ThumbnailUploadJob` is inserted

#### Scenario: Missing thumbnail enqueues the worker

- **WHEN** no thumbnail has been uploaded before `/complete` is called
- **AND** `HeadObject` on the thumbnail key returns not found
- **THEN** the image record is created with `thumbnail_path` = null
- **AND** a `ThumbnailUploadJob` is inserted

#### Scenario: Successful worker run updates ThumbnailPath

- **WHEN** all job steps succeed
- **THEN** the `Image` record has `thumbnail_path` set to `users/{kindeID}/thumbnails/{imageID}.jpg`

#### Scenario: Job failure triggers retry

- **WHEN** the `ThumbnailUploadWorker` returns an error on an attempt
- **THEN** River retries the job up to 5 attempts total before discarding

---

### Requirement: InitiateUpload response includes thumbnail_upload_url

`POST /images` SHALL return a `thumbnail_upload_url` field alongside `upload_url` and `id`. The `thumbnail_upload_url` SHALL be a presigned PUT URL for `users/{kindeID}/thumbnails/{imageID}.jpg` with content type `image/jpeg`, valid for the same TTL as the original upload URL.

#### Scenario: Initiate response contains thumbnail_upload_url

- **WHEN** `POST /images` is called with valid parameters
- **THEN** the response body includes `thumbnail_upload_url` as a non-empty string
- **AND** the response body includes `upload_url` and `id` as before

---

## REMOVED Requirements

### Requirement: Synchronous thumbnail preflight in CompleteUpload

**Reason:** Thumbnail generation has moved to the client. The server no longer needs to decode the original image during `CompleteUpload`. Removing the preflight eliminates the dependency on Go image decoders and unblocks support for WebP, AVIF, and HEIC.

**Migration:** `CompleteUpload` no longer calls `ThumbnailService.Generate` or `StorageService.GetObject` synchronously. The preflight check is replaced by `StorageService.HeadObject` on the expected thumbnail path. See the updated Async Thumbnail Storage requirement above.
