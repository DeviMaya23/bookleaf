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
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

- `folder_ids` SHALL be a non-null array of UUIDs — empty (`[]`) when the image has no folder memberships, populated with all folder IDs the image belongs to otherwise
- `position` and `folder_id` (singular) fields are removed from the response shape
- `GET /images/:id` (`imageDetailResponse`) follows the same shape and includes an additional `image_url` field

The `toImageResponse` function in `internal/handler/image.go` SHALL be updated to populate `FolderIDs []uuid.UUID` by iterating `item.Image.ImageFolders`. The `firstFolderID` and `firstFolderPosition` helpers SHALL be removed.

#### Scenario: Image list response returns paginated envelope

- **WHEN** an authenticated `GET /images` request is made
- **THEN** the response is an object with an `images` array and a `next_cursor` field
- **AND** each item in `images` includes a `folder_ids` array (never null)

#### Scenario: Image detail response includes folder_ids array

- **WHEN** an authenticated `GET /images/:id` request is made for an existing image
- **THEN** the response includes a `folder_ids` field containing all folder UUIDs the image belongs to

#### Scenario: folder_ids is empty for an unfiled image

- **WHEN** `GET /images/:id` is called for an image with no folder memberships
- **THEN** `folder_ids` in the response is an empty array `[]`

#### Scenario: folder_ids contains all memberships for a multi-folder image

- **WHEN** `GET /images/:id` is called for an image belonging to two folders
- **THEN** `folder_ids` contains both folder UUIDs

---

### Requirement: GET /images and GET /images/:id — folder_id in Response

**Reason for modification**: Replaced by `folder_ids[]` to support multi-folder membership.

The `folder_id` (singular) and `position` fields are removed from `imageResponse` and `imageDetailResponse`. The `ImageRepository` interface, `toImageResponse` helper, and all related handler structs SHALL reflect this change.

`toImageResponse` SHALL populate `FolderIDs []uuid.UUID` from `item.Image.ImageFolders` (all entries, not just index 0). The `firstFolderID` and `firstFolderPosition` helpers are removed.

#### Scenario: Image response includes folder_ids from ImageFolders

- **WHEN** an image has one or more folder memberships
- **THEN** the response includes a non-empty `folder_ids` array containing all folder UUIDs

#### Scenario: Image response has empty folder_ids for unfiled image

- **WHEN** an image has no folder membership
- **THEN** `folder_ids` in the response is `[]`
