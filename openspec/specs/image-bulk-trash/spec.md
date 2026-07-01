## Purpose

Defines the `POST /images/bulk/trash` endpoint for soft-deleting many images in a single request.

## Requirements

### Requirement: POST /images/bulk/trash — Bulk Move Images to Trash

The system SHALL expose a `POST /images/bulk/trash` endpoint on the protected route group that soft-deletes many images in a single request.

Request body:
```json
{
  "image_ids": ["uuid", ...]
}
```

- Every entry in `image_ids` MUST be a well-formed UUID. If any entry is malformed, the entire request SHALL fail with `400 Bad Request` and no images SHALL be processed.
- For each well-formed image ID that does not exist, does not belong to the authenticated user, or is already trashed: the system SHALL skip it, log it server-side, and NOT count it toward `succeeded_count`. It SHALL NOT fail the request.
- For each valid, owned, not-already-trashed image ID: the system SHALL soft-delete it using the same mechanism as the existing single-image trash deletion (setting `deleted_at`).
- One image's processing failure SHALL NOT prevent other images in the same request from being processed.
- On completion, the system SHALL return `200 OK` with `{"succeeded_count": <n>}`, where `n` is the number of image IDs successfully soft-deleted. No per-image failure detail SHALL be included in the response.
- There is no limit on the number of `image_ids` in a single request.

#### Scenario: All images trashed successfully

- **WHEN** a user submits `image_ids` for three images they own that are not already trashed
- **THEN** the system soft-deletes all three and returns `200 OK` with `{"succeeded_count": 3}`

#### Scenario: One of the images is already trashed

- **WHEN** a user submits `image_ids` including one image that is already soft-deleted
- **THEN** that image is skipped and logged server-side, the remaining images are trashed normally, and `succeeded_count` reflects only the newly-trashed images

#### Scenario: One image ID does not belong to the authenticated user

- **WHEN** a user submits `image_ids` where one ID belongs to a different user
- **THEN** that image is skipped and logged server-side, the remaining valid images are trashed normally, and `succeeded_count` reflects only the valid images

#### Scenario: image_ids contains a malformed UUID

- **WHEN** a user submits an `image_ids` entry that is not a valid UUID
- **THEN** the system returns `400 Bad Request` and no images in the request are processed

#### Scenario: All submitted image IDs are invalid

- **WHEN** every entry in `image_ids` fails ownership/existence/not-already-trashed validation
- **THEN** the system returns `200 OK` with `{"succeeded_count": 0}`
