## ADDED Requirements

### Requirement: GET /images/in-folder/:id Endpoint

The system SHALL expose `GET /images/in-folder/:id` on the protected route group, handled by `imageHandler`. It returns the full, ordered set of images belonging to a single folder — the "board/reorder view" query that `GET /images` previously served via its `folder_id` mode-switch (now removed; see `image-endpoints`).

The handler SHALL:
1. Resolve the authenticated `userID` from context
2. Parse and validate `:id` as a UUID; return `400 Bad Request` if invalid
3. Optionally accept `sort`/`direction` query parameters, validated against the same allow-list as `GET /images` (`created_at`/`title`, `asc`/`desc`; see `image-endpoints`'s `GET /images sort and direction query parameters` requirement) — invalid values return `400 Bad Request`
4. Delegate to a dedicated usecase method scoped to the given folder and the authenticated user
5. Return `404 Not Found` if the folder does not exist or is not owned by the authenticated user
6. Return `200 OK` with the full set of images in the folder

This endpoint accepts NO `cursor`/`limit`/`name`/`tag_ids`/`mime_types`/`unfiled`/`folder_ids` parameters — it answers exactly one question ("what is the ordered contents of this folder") and does not paginate or filter. Folders are bounded collections; fetching the full set in one response mirrors the existing implicit assumption of the folder-view branch this endpoint replaces.

Response body (200): a JSON array of items in the same `imageResponse` shape used by `GET /images` (see `image-endpoints`'s Response Shape requirement), with one difference — `position` is non-null, populated from `image_folders.position` for the queried folder:

```json
[
  {
    "id": "uuid",
    "title": "string",
    "...": "... all other imageResponse fields ...",
    "position": "string"
  }
]
```

#### Scenario: Authenticated user fetches a folder's images in position order

- **WHEN** an authenticated `GET /images/in-folder/:id` request is made for a folder the user owns, with no `sort` parameter
- **THEN** the response is `200 OK`
- **AND** the images are ordered by `image_folders.position ASC`
- **AND** every image in the response has a non-null `position` matching its `image_folders.position` value for that folder

#### Scenario: Explicit sort overrides position ordering

- **WHEN** an authenticated `GET /images/in-folder/:id?sort=title&direction=asc` request is made
- **THEN** the response is `200 OK`
- **AND** the images are ordered by `title ASC, id ASC` instead of by position
- **AND** each image's `position` field is still populated from `image_folders.position` (ordering and the reported position are independent concerns)

#### Scenario: Invalid folder id returns 400

- **WHEN** `GET /images/in-folder/not-a-uuid` is called
- **THEN** the response is `400 Bad Request`

#### Scenario: Folder not found or not owned returns 404

- **WHEN** an authenticated `GET /images/in-folder/:id` request is made for a folder that does not exist or belongs to another user
- **THEN** the response is `404 Not Found`

#### Scenario: Invalid sort or direction returns 400

- **WHEN** `GET /images/in-folder/:id?sort=file_size` or `GET /images/in-folder/:id?sort=title&direction=descending` is called
- **THEN** the response is `400 Bad Request`

#### Scenario: Response contains no pagination envelope

- **WHEN** an authenticated `GET /images/in-folder/:id` request succeeds
- **THEN** the response body is a plain JSON array of images (not a `{images, next_cursor}` envelope) — there is nothing to paginate

#### Scenario: Unauthenticated request is rejected

- **WHEN** a `GET /images/in-folder/:id` request is made without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: ListFolderImages Usecase Method

The system SHALL add a `ListFolderImages` method to the `ImageUsecase` interface:

```go
ListFolderImages(ctx context.Context, userID string, folderID uuid.UUID, sort *string, direction *string) ([]ImageItem, error)
```

The implementation SHALL:
1. Verify the folder exists and is owned by `userID` (via the existing folder lookup used elsewhere in the image/folder flows); return a not-found error if it does not
2. Delegate to a dedicated repository method scoped to the single folder, passing through `sort`/`direction`
3. For each returned image, populate `ImageItem.FolderPosition` by locating the `ImageFolders` entry whose `FolderID` matches the requested folder and copying its `Position` — the same logic the removed folder-view branch of `ListImages` used to perform (see `image-endpoints`)
4. Populate `ImageItem.ThumbnailURL` via the existing `thumbnailURL` helper, exactly as `ListImages`/`GetImage`/`UpdateImage` do (see `image-endpoints`'s Thumbnail URL Generation requirement)
5. Return the items with no cursor — this query is never paginated

#### Scenario: ListFolderImages returns items with FolderPosition populated

- **WHEN** `ListFolderImages` is called for a folder containing images
- **THEN** each returned `ImageItem.FolderPosition` is non-nil and equals that image's `image_folders.position` value for the requested folder

#### Scenario: ListFolderImages returns not-found for a missing or unowned folder

- **WHEN** `ListFolderImages` is called with a `folderID` that does not exist or does not belong to `userID`
- **THEN** an error indicating "not found" is returned
- **AND** no repository list query is executed

#### Scenario: ListFolderImages generates thumbnail URLs

- **WHEN** `ListFolderImages` returns images that have a non-nil `thumbnail_path`
- **THEN** each corresponding `ImageItem.ThumbnailURL` is a non-nil presigned GET URL

---

### Requirement: ImageRepository ListByFolder Method

The system SHALL add a `ListByFolder` method to the `ImageRepository` interface:

```go
ListByFolder(ctx context.Context, userID string, folderID uuid.UUID, sortField *string, direction *string) ([]*domain.Image, error)
```

This method SHALL:
- Return all non-deleted images belonging to `folderID` and owned by `userID`, joined against `image_folders`
- Preload `Tags` and `ImageFolders` on each result (consistent with `List`/`GetByID`)
- Order by `image_folders.position ASC` when `sortField` is nil
- Order by the selected column and direction (with `id` as tiebreaker) when `sortField` is non-nil, exactly like `List`'s explicit-sort behaviour
- Apply no cursor, no limit, and no `name`/`tag_ids`/`mime_types`/`unfiled` filtering — this is the single-folder, full-fetch query extracted from the old `List` folder-view branch (`internal/repository/image_repository.go:38-58` prior to this change), relocated here essentially unchanged

This is functionally the same query the old `List` method ran when `folderID` was non-nil — relocated to its own method with a narrower, single-purpose signature rather than being one branch of a dual-purpose method.

#### Scenario: ListByFolder defaults to position ordering when sort is nil

- **WHEN** `ListByFolder` is called with a nil `sortField`
- **THEN** results are ordered by `image_folders.position ASC`

#### Scenario: ListByFolder honors an explicit sort field

- **WHEN** `ListByFolder` is called with `sortField = "title"` and `direction = "asc"`
- **THEN** results are ordered by `title ASC, id ASC` instead of by position

#### Scenario: ListByFolder preloads tags and image folders

- **WHEN** `ListByFolder` is called for a folder containing images with tags and folder memberships
- **THEN** each returned `domain.Image` has its `Tags` and `ImageFolders` slices populated

#### Scenario: ListByFolder returns only images in the specified folder

- **WHEN** `ListByFolder` is called for folder A
- **THEN** only images with a row in `image_folders` for folder A are returned
- **AND** images belonging only to other folders are excluded

#### Scenario: ListByFolder is satisfied by the SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `imageRepository` in `internal/repository/` implements the `ListByFolder` method on `usecase.ImageRepository` without compilation errors
