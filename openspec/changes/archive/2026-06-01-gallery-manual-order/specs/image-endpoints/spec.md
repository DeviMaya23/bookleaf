## MODIFIED Requirements

### Requirement: GET /images and GET /images/:id — Response Shape

The `GET /images` endpoint SHALL return a paginated envelope (see `image-list-pagination` spec). The per-item `imageResponse` shape uses a presigned GET URL (24h TTL) for `thumbnail_url`:

```json
{
  "id": "uuid",
  "title": "string",
  "description": "string|null",
  "mime_type": "string",
  "source_url": "string|null",
  "folder_ids": ["uuid"],
  "thumbnail_url": "string|null",
  "width": "integer|null",
  "height": "integer|null",
  "file_size": "integer|null",
  "tags": [{ "id": "uuid", "name": "string" }],
  "position": "string|null",
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

- `folder_ids` SHALL be a non-null array of UUIDs — empty (`[]`) when the image has no folder memberships
- `position` SHALL be the fracdex key from `image_folders.position` for the queried folder when `GET /images` is called with a `folder_id` parameter; `null` in all other list contexts (unfiled, all, trash)
- `GET /images/:id` (`imageDetailResponse`) follows the same shape as before and includes an additional `image_url` field; `position` is always `null` in the detail response (no folder context)

The `ImageItem` struct in `internal/usecase/image_usecase.go` SHALL add a `FolderPosition *string` field. In `ListImages`, when `params.FolderID` is non-nil, the implementation SHALL iterate the image's `ImageFolders` slice to find the entry matching `params.FolderID` and set `FolderPosition` to that entry's `Position`. The `toImageResponse` function in `internal/handler/image.go` SHALL map `item.FolderPosition` to `Position` on the response struct.

#### Scenario: Image list response returns paginated envelope

- **WHEN** an authenticated `GET /images` request is made
- **THEN** the response is an object with an `images` array and a `next_cursor` field
- **AND** each item in `images` includes a `folder_ids` array (never null)

#### Scenario: Folder-scoped list includes position

- **WHEN** an authenticated `GET /images?folder_id=<uuid>` request is made
- **THEN** each image in the response includes a non-null `position` string

#### Scenario: Non-folder-scoped list returns null position

- **WHEN** an authenticated `GET /images` request is made without `folder_id` (e.g. unfiled or all)
- **THEN** each image in the response has `position: null`

#### Scenario: Image detail response includes folder_ids array

- **WHEN** an authenticated `GET /images/:id` request is made for an existing image
- **THEN** the response includes a `folder_ids` field containing all folder UUIDs the image belongs to

#### Scenario: folder_ids is empty for an unfiled image

- **WHEN** `GET /images/:id` is called for an image with no folder memberships
- **THEN** `folder_ids` in the response is an empty array `[]`

#### Scenario: folder_ids contains all memberships for a multi-folder image

- **WHEN** `GET /images/:id` is called for an image belonging to two folders
- **THEN** `folder_ids` contains both folder UUIDs
