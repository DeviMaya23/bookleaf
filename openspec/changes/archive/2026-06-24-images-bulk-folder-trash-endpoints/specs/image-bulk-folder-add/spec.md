## ADDED Requirements

### Requirement: POST /images/bulk/add-to-folder — Bulk Add Images to a Folder

The system SHALL expose a `POST /images/bulk/add-to-folder` endpoint on the protected route group that adds many images to one folder in a single request.

Request body:
```json
{
  "image_ids": ["uuid", ...],
  "folder_id": "uuid"
}
```

- `folder_id` MUST exist and belong to the authenticated user. If it does not, the entire request SHALL fail with `404 Not Found` and no images SHALL be processed.
- Every entry in `image_ids` MUST be a well-formed UUID. If any entry is malformed, the entire request SHALL fail with `400 Bad Request` and no images SHALL be processed.
- For each well-formed image ID that does not exist or does not belong to the authenticated user: the system SHALL skip it, log it server-side, and NOT count it toward `succeeded_count`. It SHALL NOT fail the request.
- For each valid, owned image ID: the system SHALL ensure a row exists in `image_folders` for `(image_id, folder_id)`, appending a fracdex position after the current maximum position in that folder if no such row exists yet. If the row already exists, it SHALL be left unchanged (no error, no position change).
- An image that already belongs to the folder SHALL still count toward `succeeded_count`.
- No other folder memberships for the affected images SHALL be changed.
- One image's processing failure SHALL NOT prevent other images in the same request from being processed.
- On completion, the system SHALL return `200 OK` with `{"succeeded_count": <n>}`, where `n` is the number of image IDs successfully processed (including idempotent no-ops). No per-image failure detail SHALL be included in the response.
- There is no limit on the number of `image_ids` in a single request.

#### Scenario: All images added successfully to a folder they are not yet in

- **WHEN** a user submits `image_ids` for three images not currently in the target folder, with a `folder_id` the user owns
- **THEN** the system inserts an `image_folders` row for each of the three images and returns `200 OK` with `{"succeeded_count": 3}`

#### Scenario: One of the images is already in the target folder

- **WHEN** a user submits `image_ids` including one image that already belongs to the target folder
- **THEN** that image's existing `image_folders` row is left unchanged, no error occurs, and it is still counted in `succeeded_count`

#### Scenario: One image ID does not belong to the authenticated user

- **WHEN** a user submits `image_ids` where one ID belongs to a different user
- **THEN** that image is skipped and logged server-side, the remaining valid images are processed normally, and `succeeded_count` reflects only the valid images

#### Scenario: folder_id does not exist or is not owned by the user

- **WHEN** a user submits a `folder_id` that does not exist or belongs to another user
- **THEN** the system returns `404 Not Found` and no `image_folders` rows are inserted for any image in the request

#### Scenario: image_ids contains a malformed UUID

- **WHEN** a user submits an `image_ids` entry that is not a valid UUID
- **THEN** the system returns `400 Bad Request` and no images in the request are processed

#### Scenario: All submitted image IDs are invalid

- **WHEN** every entry in `image_ids` fails ownership/existence validation
- **THEN** the system returns `200 OK` with `{"succeeded_count": 0}`
