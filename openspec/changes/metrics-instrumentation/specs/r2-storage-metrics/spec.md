## ADDED Requirements

### Requirement: Presigned URL Generation Duration Metrics

The system SHALL record a `r2.presigned_url.duration` Float64Histogram instrument (unit: `ms`) in `r2Storage`. The histogram SHALL be initialised once in `NewR2Storage` and stored as a field on the struct. It SHALL be recorded at the end of both `GeneratePresignedPutURL` and `GeneratePresignedGetURL`, measuring the elapsed time of the underlying AWS SDK presign call. The `r2.operation` attribute SHALL be set to `"presigned_put"` or `"presigned_get"` respectively. The `r2.status` attribute SHALL be set to `"success"` or `"error"`.

#### Scenario: Presigned PUT URL generation duration recorded on success

- **WHEN** `GeneratePresignedPutURL` succeeds
- **THEN** `r2.presigned_url.duration` is recorded with `r2.operation="presigned_put"` and `r2.status="success"`

#### Scenario: Presigned PUT URL generation duration recorded on failure

- **WHEN** `GeneratePresignedPutURL` returns an error
- **THEN** `r2.presigned_url.duration` is recorded with `r2.operation="presigned_put"` and `r2.status="error"`

#### Scenario: Presigned GET URL generation duration recorded on success

- **WHEN** `GeneratePresignedGetURL` succeeds
- **THEN** `r2.presigned_url.duration` is recorded with `r2.operation="presigned_get"` and `r2.status="success"`

#### Scenario: Presigned GET URL generation duration recorded on failure

- **WHEN** `GeneratePresignedGetURL` returns an error
- **THEN** `r2.presigned_url.duration` is recorded with `r2.operation="presigned_get"` and `r2.status="error"`

### Requirement: Upload Completion Count Metric

The system SHALL record an `r2.upload.count` Int64Counter instrument in `imageUsecase`. The counter SHALL be initialised once in `NewImageUsecase` and stored as a field on the struct. It SHALL be incremented in `CompleteUpload` with a `r2.status` attribute set to `"success"` on successful completion or `"error"` if the repository call fails before the thumbnail goroutine is dispatched.

#### Scenario: Successful upload completion increments success counter

- **WHEN** `CompleteUpload` succeeds (image record found, goroutine dispatched)
- **THEN** `r2.upload.count` is incremented with `r2.status="success"`

#### Scenario: Failed upload completion increments error counter

- **WHEN** `CompleteUpload` returns an error (e.g. image not found)
- **THEN** `r2.upload.count` is incremented with `r2.status="error"`

### Requirement: Thumbnail Generation Metrics

The system SHALL record a `r2.thumbnail.duration` Float64Histogram instrument (unit: `ms`) and an `r2.thumbnail.count` Int64Counter instrument in `imageUsecase`. Both instruments SHALL be initialised once in `NewImageUsecase` and stored as fields on the struct. They SHALL be recorded in `generateThumbnail` at the point where the goroutine concludes (success or any failure step). The `r2.status` attribute SHALL be `"success"` or `"error"`.

#### Scenario: Successful thumbnail generation records duration and increments success counter

- **WHEN** `generateThumbnail` completes successfully
- **THEN** `r2.thumbnail.duration` is recorded with `r2.status="success"`
- **AND** `r2.thumbnail.count` is incremented with `r2.status="success"`

#### Scenario: Failed thumbnail generation records duration and increments error counter

- **WHEN** `generateThumbnail` fails at any step (fetch, generate, put, or DB update)
- **THEN** `r2.thumbnail.duration` is recorded with `r2.status="error"`
- **AND** `r2.thumbnail.count` is incremented with `r2.status="error"`
