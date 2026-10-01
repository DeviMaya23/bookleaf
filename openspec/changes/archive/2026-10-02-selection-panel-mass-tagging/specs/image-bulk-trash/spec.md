## MODIFIED Requirements

### Requirement: POST /images/bulk/trash — Bulk Move Images to Trash

The system SHALL expose a `POST /images/bulk/trash` endpoint on the protected route group that soft-deletes many images in a single request.

Request body:
```json
{
  "image_ids": ["uuid", ...]
}
```

- Every entry in `image_ids` MUST be a well-formed UUID. If any entry is malformed, the entire request SHALL fail with `400 Bad Request` and no images SHALL be processed.
- The system SHALL filter the submitted IDs to those that exist, are not already trashed, and belong to the authenticated user. Unowned, non-existent, or already-trashed IDs SHALL be silently ignored and SHALL NOT be counted in `succeeded_count`.
- The system SHALL soft-delete all qualifying images in a single atomic database statement (`UPDATE images SET deleted_at = NOW() WHERE id IN (?) AND user_id = ? AND deleted_at IS NULL`). `succeeded_count` SHALL be derived from the number of rows affected.
- On completion, the system SHALL return `200 OK` with `{"succeeded_count": <n>}`, where `n` is the number of images successfully soft-deleted.

#### Scenario: All images trashed successfully

- **WHEN** a user submits `image_ids` for three images they own that are not already trashed
- **THEN** the system soft-deletes all three in a single statement and returns `200 OK` with `{"succeeded_count": 3}`

#### Scenario: One of the images is already trashed

- **WHEN** a user submits `image_ids` including one image that is already soft-deleted
- **THEN** that image is excluded from the update, the remaining images are trashed, and `succeeded_count` reflects only the newly-trashed images

#### Scenario: One image ID does not belong to the authenticated user

- **WHEN** a user submits `image_ids` where one ID belongs to a different user
- **THEN** that image is excluded from the update, the remaining valid images are trashed, and `succeeded_count` reflects only the valid images

#### Scenario: image_ids contains a malformed UUID

- **WHEN** a user submits an `image_ids` entry that is not a valid UUID
- **THEN** the system returns `400 Bad Request` and no images in the request are processed

#### Scenario: All submitted image IDs are invalid

- **WHEN** every entry in `image_ids` fails ownership/existence/not-already-trashed validation
- **THEN** the system returns `200 OK` with `{"succeeded_count": 0}`
