## MODIFIED Requirements

### Requirement: CompleteUpload Response Body

The `POST /images/:id/complete` handler SHALL return `200 OK` with a JSON body on success.

Response shape:
```json
{
  "image_id": "<uuid>"
}
```

- `image_id` SHALL always be present
- `suggested_folder_name` and `warning` fields are removed; vision labelling is now asynchronous
- If thumbnail generation fails in the synchronous phase, the handler SHALL return a non-2xx error response

`CompleteUploadResult` in `internal/usecase/` SHALL be simplified to:
```go
type CompleteUploadResult struct {
    ImageID uuid.UUID
}
```

#### Scenario: Successful CompleteUpload returns only image_id

- **WHEN** `CompleteUpload` succeeds
- **THEN** the response is `200 OK`
- **AND** the response body contains `image_id`
- **AND** the response body does not contain `suggested_folder_name` or `warning`

#### Scenario: Thumbnail generation failure returns error

- **WHEN** `prepareThumbnail` returns an error during `CompleteUpload`
- **THEN** the response is `500 Internal Server Error`
- **AND** the `pending_uploads` row is not committed to `images`

## REMOVED Requirements

### Requirement: CompleteUpload Response — suggested_folder_name and warning fields

**Reason**: Vision labelling is now executed asynchronously via River. The response can no longer carry a synchronous vision result. The `SuggestedFolderName` and `Warning` fields are removed from `CompleteUploadResult`.

**Migration**: FE code reading `suggested_folder_name` or `warning` from the `POST /images/:id/complete` response must be updated. The `POST /images/:id/accept-suggestion` endpoint remains available for when vision results are surfaced through another mechanism (e.g., polling `GET /images/:id` for `ai_labels`).
