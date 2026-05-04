## MODIFIED Requirements

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
- Image moved to folder: INFO, `event: "image.mutated"`, `image_id`, `user_id`, `operation: "moved_to_folder"`, `folder_id` — emitted by `ImageUsecase.UpdateImage` ONLY when `folder_id` is present in the update params AND the new value differs from the current value
- Image moved to trash: INFO, `event: "image.mutated"`, `image_id`, `user_id`, `operation: "trashed"`

Title-only updates SHALL NOT emit a domain event log. The span and the `LoggingMiddleware` request log are sufficient for title edits.

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

#### Scenario: Folder change emits moved_to_folder event

- **WHEN** `ImageUsecase.UpdateImage` is called with a `folder_id` that differs from the image's current `folder_id`
- **THEN** an INFO log is emitted with `event: "image.mutated"`, `operation: "moved_to_folder"`, `image_id`, `user_id`, and `folder_id`

#### Scenario: Title-only update emits no domain event log

- **WHEN** `ImageUsecase.UpdateImage` is called with only `title` provided
- **THEN** no `image.mutated` log event is emitted
- **AND** the request is still visible in the `LoggingMiddleware` request log and the usecase span

#### Scenario: folder_id unchanged emits no event

- **WHEN** `ImageUsecase.UpdateImage` is called with a `folder_id` equal to the image's current `folder_id`
- **THEN** no `image.mutated` log event is emitted
