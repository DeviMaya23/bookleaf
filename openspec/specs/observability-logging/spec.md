# observability-logging Specification

## Purpose
Defines structured logging requirements using Zap, including logger initialisation and trace ID injection from OTel span context.

## Requirements

### Requirement: Zap Logger Initialisation

The system SHALL initialise a `*zap.Logger` in `cmd/server/main.go` before any other component. The logger SHALL be constructed by a `NewLogger(format string) (*zap.Logger, error)` function in `internal/observability/logging.go`. When `format` is `"json"`, it SHALL use `zap.NewProduction()` config. When `format` is `"console"` (or any other value), it SHALL use `zap.NewDevelopment()` config. The function SHALL return an error if Zap fails to build.

#### Scenario: Production logger in JSON format

- **WHEN** `LOG_FORMAT=json`
- **THEN** `NewLogger("json")` returns a `*zap.Logger` configured with production JSON encoding
- **AND** log output is valid JSON with `level`, `ts`, `msg` fields

#### Scenario: Development logger in console format

- **WHEN** `LOG_FORMAT=console` (or `LOG_FORMAT` is unset, defaulting to `console`)
- **THEN** `NewLogger("console")` returns a `*zap.Logger` with human-readable console output

### Requirement: Trace ID Field Injection

The system SHALL provide a `LoggerFromContext(ctx context.Context, base *zap.Logger) *zap.Logger` function in `internal/observability/logging.go`. If the context carries an active OTel span with a valid trace ID, the function SHALL return a child logger with a `trace_id` string field appended. If no active span exists or the trace ID is invalid, it SHALL return the base logger unchanged.

#### Scenario: Active span in context

- **WHEN** a context with a valid OTel span is passed to `LoggerFromContext`
- **THEN** the returned logger includes a `trace_id` field matching the span's trace ID in hex format

#### Scenario: No span in context

- **WHEN** a context without an active span is passed to `LoggerFromContext`
- **THEN** the returned logger is the base logger with no `trace_id` field added
### Requirement: Auth Event Logging

The system SHALL emit structured log events for authentication and authorisation failures.

The Kinde auth middleware SHALL emit a WARN log when it rejects an inbound request. The log SHALL include:
- `event: "auth.token_rejected"`
- `reason`: a short string identifying the rejection cause (e.g. `"missing_header"`, `"invalid_token"`, `"expired_token"`)
- `trace_id` (via `LoggerFromContext`)

No new user creation event is emitted by the auth middleware. The user usecase emits an INFO log when a new user record is persisted (see Usecase Domain Events below).

#### Scenario: Token rejection is logged at WARN

- **WHEN** the Kinde auth middleware rejects a request due to a missing or invalid token
- **THEN** a WARN log is emitted with `event: "auth.token_rejected"` and a `reason` field

### Requirement: Usecase Domain Event Logging

Usecase methods SHALL emit structured INFO or ERROR logs at the points listed below. All log calls SHALL pass the request context through `LoggerFromContext(ctx, tel.Logger)` to include `trace_id`.

**User events:**
- New user created: INFO, `event: "user.created"`, `user_id`

**R2 upload events (emitted from `ImageUsecase.InitiateUpload` and `ImageUsecase.CompleteUpload`):**
- Upload initiated: INFO, `event: "r2.upload.started"`, `image_id`, `user_id`, `mime_type`, `file_size`, `r2_key`
- Upload completed: INFO, `event: "r2.upload.completed"`, `image_id`, `user_id`, `duration_ms`

**Thumbnail events (emitted from the thumbnail processing path in `ImageUsecase`):**
- Job started: INFO, `event: "thumbnail.job.started"`, `image_id`, `user_id`
- Job completed: INFO, `event: "thumbnail.job.completed"`, `image_id`, `user_id`, `duration_ms`
- Job failed: ERROR, `event: "thumbnail.job.failed"`, `image_id`, `user_id`, `error`

**Image mutation events:**
- Image metadata edited: INFO, `event: "image.mutated"`, `image_id`, `user_id`, `operation: "edited"`
- Image moved to folder: INFO, `event: "image.mutated"`, `image_id`, `user_id`, `operation: "moved_to_folder"`, `folder_id`
- Image moved to trash: INFO, `event: "image.mutated"`, `image_id`, `user_id`, `operation: "trashed"`

**Folder mutation events:**
- Folder deleted: INFO, `event: "folder.mutated"`, `folder_id`, `user_id`, `operation: "deleted"`, `image_count` (number of images that were in the folder)

#### Scenario: Upload initiation emits start event

- **WHEN** `ImageUsecase.InitiateUpload` succeeds in creating the image record
- **THEN** an INFO log is emitted with `event: "r2.upload.started"`, `image_id`, `user_id`, `mime_type`, `file_size`, and `r2_key`

#### Scenario: Thumbnail failure emits error event

- **WHEN** the thumbnail generation job fails
- **THEN** an ERROR log is emitted with `event: "thumbnail.job.failed"`, `image_id`, `user_id`, and `error`
- **AND** no `duration_ms` field is present on the failure log

#### Scenario: Folder deletion log includes image count

- **WHEN** a folder is deleted
- **THEN** an INFO log is emitted with `event: "folder.mutated"`, `operation: "deleted"`, `folder_id`, `user_id`, and `image_count` reflecting the number of images that were associated with the folder

### Requirement: R2 Storage Operation Logging

The `r2Storage` implementation SHALL emit structured log events for presigned URL operations using the logger from its `*Telemetry` field. The logger SHALL be enriched with trace context via `LoggerFromContext(ctx, tel.Logger)` on each call.

Events:
- Presigned PUT URL generated successfully: INFO, `event: "r2.presigned_put.success"`, `image_id`
- Presigned PUT URL generation failed: ERROR, `event: "r2.presigned_put.failed"`, `image_id`, `error`
- Presigned GET URL generated successfully: INFO, `event: "r2.presigned_get.success"`, `image_id`
- Presigned GET URL generation failed: ERROR, `event: "r2.presigned_get.failed"`, `image_id`, `error`

The actual presigned URL string SHALL NOT appear in any log entry.

#### Scenario: Presigned PUT URL success is logged without URL value

- **WHEN** `GeneratePresignedPutURL` returns successfully
- **THEN** an INFO log is emitted with `event: "r2.presigned_put.success"` and `image_id`
- **AND** the log entry does not contain the presigned URL string

#### Scenario: Presigned PUT URL failure is logged with error

- **WHEN** `GeneratePresignedPutURL` returns an error
- **THEN** an ERROR log is emitted with `event: "r2.presigned_put.failed"`, `image_id`, and `error`

### Requirement: Logging Field Conventions

All structured log entries in application layers SHALL use the following standardised field names. Using alternate names (e.g. `imageID`, `userId`) is not permitted.

| Field | Type | Description |
|---|---|---|
| `event` | string | Dot-namespaced event identifier (e.g. `"r2.upload.started"`) |
| `user_id` | string | Kinde user ID |
| `image_id` | string | Image UUID as string |
| `folder_id` | string | Folder UUID as string |
| `file_size` | int64 | Size in bytes |
| `mime_type` | string | MIME type string |
| `r2_key` | string | Full R2 object key |
| `duration_ms` | float64 | Elapsed wall time in milliseconds |
| `http.request.method` | string | HTTP method — OTel semconv, consistent with `MetricsMiddleware` |
| `http.route` | string | Matched Echo route pattern — OTel semconv |
| `http.response.status_code` | int | HTTP response status code — OTel semconv |
| `image_count` | int | Number of images (folder deletion) |
| `operation` | string | Mutation type (e.g. `"edited"`, `"trashed"`) |
| `reason` | string | Rejection reason (auth token rejection) |
| `error` | string | Error message (failure events only) |

No image binary data, no presigned URL strings, and no raw JWT claims SHALL appear in any log entry.

#### Scenario: Field names are consistent across layers

- **WHEN** an image ID is logged by the usecase layer
- **THEN** the field key is `image_id` (not `imageID`, `image-id`, or `id`)
