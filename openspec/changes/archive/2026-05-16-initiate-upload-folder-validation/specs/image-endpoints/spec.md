## MODIFIED Requirements

### Requirement: POST /images — Initiate Upload Request and Response

The `POST /images` handler SHALL accept an optional `description` field and an optional `folder_id` field in the request body.

When `folder_id` is provided, the usecase SHALL look up the folder by ID scoped to the authenticated user via `folderRepo.GetByID`. If the folder is not found, the image SHALL be created with `folder_id = null`. No error SHALL be returned to the caller in this case.

Request body:
```json
{
  "title": "string (required)",
  "mime_type": "string (required)",
  "source_url": "string (optional)",
  "folder_id": "uuid (optional)",
  "description": "string (optional)"
}
```

Response body (201): `id`, `upload_url`, `r2_path`.

#### Scenario: Upload initiated with a valid folder_id

- **WHEN** an authenticated `POST /images` request includes a `folder_id` that exists and belongs to the user
- **THEN** the response is `201 Created`
- **AND** the created image record has `folder_id` set to the provided value

#### Scenario: Upload initiated with a folder_id that does not exist

- **WHEN** an authenticated `POST /images` request includes a `folder_id` that does not exist (or belongs to another user)
- **THEN** the response is `201 Created`
- **AND** the created image record has `folder_id` set to `null`

#### Scenario: Upload initiated with null or omitted folder_id

- **WHEN** an authenticated `POST /images` request omits `folder_id` or sets it to `null`
- **THEN** the response is `201 Created`
- **AND** the image record has `folder_id` as NULL

#### Scenario: Upload initiated with description

- **WHEN** an authenticated `POST /images` request includes a non-empty `description`
- **THEN** the response is `201 Created`
- **AND** the created image record has the supplied `description` value persisted

#### Scenario: Upload initiated without description

- **WHEN** an authenticated `POST /images` request omits `description`
- **THEN** the response is `201 Created`
- **AND** the image record has `description` as NULL
