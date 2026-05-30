## MODIFIED Requirements

### Requirement: CompleteUpload Response Body

The `POST /images/:id/complete` handler SHALL return `200 OK` with a JSON body on success.

Response shape:
```json
{
  "image_id": "<uuid>",
  "suggested_folder_name": "<string | null>",
  "warning": "<string>"
}
```

- `image_id` SHALL always be present
- `suggested_folder_name` SHALL be `null` when the user does not have `vision_enabled`, when the Vision API returns no labels, or when Vision is not configured
- `warning` SHALL be omitted from the response when empty (`omitempty`)
- If thumbnail generation fails, the handler SHALL return a non-2xx error response. The `warning` field SHALL NOT be used for thumbnail failures.

#### Scenario: Vision enabled and suggestion resolved

- **WHEN** `CompleteUpload` succeeds and Vision returns at least one label
- **THEN** the response is `200 OK`
- **AND** `suggested_folder_name` is the top label description string
- **AND** `warning` is absent from the response body

#### Scenario: Vision enabled but API call fails

- **WHEN** `CompleteUpload` succeeds but the Vision API returns an error
- **THEN** the response is still `200 OK`
- **AND** `suggested_folder_name` is `null`
- **AND** `warning` is a non-empty string describing the failure

#### Scenario: Vision not enabled

- **WHEN** the image owner has `vision_enabled = false`
- **THEN** the response is `200 OK`
- **AND** `suggested_folder_name` is `null`
- **AND** `warning` is absent

#### Scenario: Thumbnail generation fails

- **WHEN** `prepareThumbnail` returns an error during `CompleteUpload`
- **THEN** the response is `500 Internal Server Error`
- **AND** `is_uploaded` remains false on the image record
