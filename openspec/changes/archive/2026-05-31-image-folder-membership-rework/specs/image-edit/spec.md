## MODIFIED Requirements

### Requirement: PATCH /images/:id — Edit Image Metadata

The system SHALL expose a `PATCH /images/:id` endpoint on the protected route group that updates the `title`, `folder_ids`, and/or `source_url` of an existing image. No field is required; omitting a field means that field is left unchanged. The image binary (`r2_path`) SHALL NOT be modifiable via this endpoint.

Request body:
```json
{
  "title": "string (optional)",
  "folder_ids": "uuid[] | null (optional)",
  "source_url": "string | null (optional)"
}
```

- `folder_ids` is decoded as `json.RawMessage`. Three states are distinguished:
  - **Absent** (field not present in body): no folder changes are made
  - **Explicit `null`**: treated the same as absent; no folder changes are made
  - **Array** (including empty `[]`): the complete desired set of folder memberships; the BE diffs and reconciles
- An empty `folder_ids` array (`[]`) SHALL remove the image from all folders.
- A non-empty `folder_ids` array SHALL sync the image's folder memberships to exactly that set (diff against current; positions of unchanged memberships are preserved; departed memberships are deleted; new memberships are inserted with fracdex positions appended to end).
- `folder_id` (singular) is no longer accepted. Sending `folder_id` has no effect.
- A `null` `source_url` in the request body SHALL clear the source URL on the image.
- An absent `source_url` field SHALL leave the current `source_url` unchanged.
- The handler SHALL distinguish absent from null using a presence flag or pointer-of-pointer decoding — not `omitempty` alone.
- `title`, if present, MUST NOT be empty string.
- The image MUST be owned by the authenticated user.
- Returns `404 Not Found` if the image does not exist or belongs to another user.
- Returns `400 Bad Request` if the body is malformed or `title` is an empty string.

#### Scenario: Title is updated

- **WHEN** an authenticated `PATCH /images/:id` request is made with `{"title": "new name"}`
- **THEN** the response is `200 OK`
- **AND** the returned image has `title` set to `"new name"`
- **AND** the image's folder memberships and `source_url` are unchanged

#### Scenario: folder_ids absent leaves memberships unchanged

- **WHEN** an authenticated `PATCH /images/:id` request is made without a `folder_ids` field
- **THEN** the response is `200 OK`
- **AND** the image's folder memberships are not modified

#### Scenario: folder_ids null leaves memberships unchanged

- **WHEN** an authenticated `PATCH /images/:id` request is made with `{"folder_ids": null}`
- **THEN** the response is `200 OK`
- **AND** the image's folder memberships are not modified

#### Scenario: Empty folder_ids removes all memberships

- **WHEN** an authenticated `PATCH /images/:id` request is made with `{"folder_ids": []}`
- **THEN** the response is `200 OK`
- **AND** the image has no rows in `image_folders`

#### Scenario: folder_ids syncs memberships and preserves unchanged positions

- **WHEN** an authenticated `PATCH /images/:id` request is made with `{"folder_ids": ["B", "C"]}` and the image currently belongs to folders `[A, C]`
- **THEN** the response is `200 OK`
- **AND** the `(imageID, A)` row is deleted from `image_folders`
- **AND** a `(imageID, B)` row is inserted with a valid fracdex position
- **AND** the `(imageID, C)` row is unchanged (position preserved)

#### Scenario: Image is moved to root with empty folder_ids

- **WHEN** an authenticated `PATCH /images/:id` request is made with `{"folder_ids": []}`
- **THEN** the response is `200 OK`
- **AND** the image has no folder memberships

#### Scenario: Source URL is updated

- **WHEN** an authenticated `PATCH /images/:id` request is made with `{"source_url": "https://example.com"}`
- **THEN** the response is `200 OK`
- **AND** the returned image has `source_url` set to `"https://example.com"`
- **AND** the image's `title` and folder memberships are unchanged

#### Scenario: Source URL is cleared with null

- **WHEN** an authenticated `PATCH /images/:id` request is made with `{"source_url": null}`
- **THEN** the response is `200 OK`
- **AND** the returned image has `source_url` set to `null`

#### Scenario: Empty title is rejected

- **WHEN** an authenticated `PATCH /images/:id` request is made with `{"title": ""}`
- **THEN** the response is `400 Bad Request`

#### Scenario: Image not found or not owned

- **WHEN** an authenticated `PATCH /images/:id` request is made for a non-existent or unowned image
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `PATCH /images/:id` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: UpdateImage Usecase Method

The `ImageUsecase` interface SHALL include an `UpdateImage(ctx, id uuid.UUID, userID string, params UpdateImageParams) (*ImageItem, error)` method.

`UpdateImageParams` SHALL use pointer fields so the usecase can distinguish absent from provided values:
- `Title *string` — nil means unchanged; non-nil means update to this value
- `FolderIDs *[]uuid.UUID` — nil means unchanged; non-nil pointer to empty slice means remove all memberships; non-nil pointer to non-empty slice means sync to that set
- `SourceURL **string` — nil outer pointer means unchanged; non-nil outer pointer with nil inner pointer means clear source URL; non-nil inner pointer means set to that string
- `Tags *[]uuid.UUID` — nil means unchanged; non-nil means replace all tags with this set

`FolderID **uuid.UUID` is removed.

The usecase SHALL:
1. Fetch the existing image by `id` and `userID`; return `gorm.ErrRecordNotFound` if not found
2. Build a map of only the scalar fields that are non-nil in `params`
3. Delegate scalar updates to `ImageRepository.Update`
4. If `params.Tags` is non-nil: delegate to `TagRepository.ReplaceImageTags`
5. If `params.FolderIDs` is non-nil: delegate to `ImageRepository.SyncImageFolders`

#### Scenario: Only provided fields are updated

- **WHEN** `UpdateImage` is called with `params.Title = nil` and `params.FolderIDs` pointing to a slice
- **THEN** only folder memberships are written to the database
- **AND** `title` and `source_url` retain their previous values

#### Scenario: Not found returns error

- **WHEN** `UpdateImage` is called for an image that does not exist or is owned by another user
- **THEN** the method returns `gorm.ErrRecordNotFound`

---

### Requirement: SyncImageFolders Repository Method

The `ImageRepository` interface SHALL include a `SyncImageFolders(ctx context.Context, imageID uuid.UUID, folderIDs []uuid.UUID) error` method.

The implementation SHALL execute within a single database transaction:
1. Fetch current `image_folders` rows for `imageID`
2. Compute `toDelete` = current folder IDs not in `folderIDs`; compute `toAdd` = folder IDs in `folderIDs` not in current
3. Delete rows for `toDelete`: `DELETE FROM image_folders WHERE image_id = ? AND folder_id IN (?)`
4. For each folder ID in `toAdd`: compute fracdex position (append to end of that folder) and insert a new row
5. Rows in neither `toDelete` nor `toAdd` are not touched (positions preserved)

#### Scenario: Sync removes departed and adds new memberships

- **WHEN** `SyncImageFolders` is called with `folderIDs = [B, C]` and the image currently has memberships `[A, C]`
- **THEN** the `(imageID, A)` row is deleted
- **AND** a `(imageID, B)` row is inserted with a valid fracdex position
- **AND** the `(imageID, C)` row is unchanged

#### Scenario: Sync with empty slice removes all memberships

- **WHEN** `SyncImageFolders` is called with `folderIDs = []`
- **THEN** all `image_folders` rows for `imageID` are deleted

#### Scenario: Sync with same set is a no-op

- **WHEN** `SyncImageFolders` is called with a `folderIDs` set identical to the current memberships
- **THEN** no rows are deleted or inserted
