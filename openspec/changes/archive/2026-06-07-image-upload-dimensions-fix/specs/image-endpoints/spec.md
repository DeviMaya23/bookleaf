## MODIFIED Requirements

### Requirement: CompleteUpload Populates Dimensions and File Size

The `POST /images/:id/complete` handler SHALL accept an optional JSON request body containing client-supplied `width`, `height`, and `file_size` (integers), and the usecase SHALL persist them on the `domain.Image` record being created — instead of deriving them server-side from the uploaded bytes.

- The request body is optional; any of `width`, `height`, `file_size` MAY be omitted
- A supplied value SHALL be persisted only if it is a positive integer (`> 0`); otherwise the corresponding field SHALL be stored as `NULL`
- No image bytes SHALL be fetched from R2 or decoded by the backend to derive these fields (the prior `extractImageMetadata` decode path, including the `image.DecodeConfig`/`image/jpeg`/`image/png` stdlib usage, is removed)
- The `CompleteUpload` response body is unchanged

#### Scenario: Client-supplied dimensions and size are persisted

- **WHEN** `CompleteUpload` is called with a request body of `{ "width": 1920, "height": 1080, "file_size": 245760 }`
- **THEN** the image record is created with `width = 1920`, `height = 1080`, and `file_size = 245760`

#### Scenario: Implausible values are stored as NULL

- **WHEN** `CompleteUpload` is called with a request body containing a non-positive value, e.g. `{ "width": -1, "height": 0, "file_size": 1024 }`
- **THEN** the image record is created with `width = NULL` and `height = NULL`
- **AND** `file_size = 1024` is persisted
- **AND** the upload completes successfully (no error is returned)

#### Scenario: Omitted fields are stored as NULL

- **WHEN** `CompleteUpload` is called with no request body (or a body omitting `width`, `height`, and `file_size`)
- **THEN** the image record is created with `width = NULL`, `height = NULL`, and `file_size = NULL`
- **AND** the upload completes successfully (no error is returned)
