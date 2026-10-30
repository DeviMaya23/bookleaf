### Requirement: InitiateUpload response includes thumbnail_upload_url

`POST /images` SHALL return a `thumbnail_upload_url` field alongside `upload_url` and `id`. The `thumbnail_upload_url` SHALL be a presigned PUT URL for `users/{kindeID}/thumbnails/{imageID}.jpg` with content type `image/jpeg`, valid for the same TTL as the original upload URL.

#### Scenario: Initiate response contains thumbnail_upload_url

- **WHEN** `POST /images` is called with valid parameters
- **THEN** the response body includes `thumbnail_upload_url` as a non-empty string
- **AND** the response body includes `upload_url` and `id` as before

---

### Requirement: CompleteUpload sets thumbnail path unconditionally

During `CompleteUpload`, the system SHALL compute the thumbnail key as `users/{kindeID}/thumbnails/{imageID}.jpg` and assign it to the image record's `thumbnail_path` at creation time. No R2 existence check is performed. The `ThumbnailUploadArgs` job is never enqueued. Only the vision labelling job is enqueued as before.

#### Scenario: CompleteUpload always sets thumbnail_path

- **WHEN** `POST /images/:id/complete` is called
- **THEN** the image record is created with `thumbnail_path` set to `users/{kindeID}/thumbnails/{imageID}.jpg`
- **AND** no `ThumbnailUploadJob` is inserted
- **AND** a vision labelling job is enqueued as normal

---

### Requirement: Thumbnail GET URLs are deterministic

All endpoints that return a `thumbnail_url` field SHALL generate it via `GenerateDeterministicPresignedGetURL`. The URL SHALL be stable for the full 24 h UTC window so the browser can cache the image across repeated API calls within the same window.

#### Scenario: Same thumbnail URL is returned on consecutive calls within a window

- **WHEN** any endpoint that returns `thumbnail_url` is called twice for the same image within the same 24 h UTC window
- **THEN** both responses contain the identical `thumbnail_url` value

#### Scenario: Endpoints affected

- **WHEN** any of the following endpoints responds with image data that includes a thumbnail
- **THEN** the `thumbnail_url` is produced by `GenerateDeterministicPresignedGetURL`
- **AND** the affected endpoints are: list images, get image detail, list folder images, list trash items, get trash item, list shared folder images, get shared image detail
