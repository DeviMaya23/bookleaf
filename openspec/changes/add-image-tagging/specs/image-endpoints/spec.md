## MODIFIED Requirements

### Requirement: Image Usecase Interface

The system SHALL define an `ImageUsecase` interface in `internal/usecase/` with methods corresponding to the HTTP operations, including `UpdateImage`. The `CompleteUpload` method SHALL return a result struct alongside the error.

```go
CompleteUpload(ctx context.Context, id uuid.UUID, userID string) (*CompleteUploadResult, error)
AcceptSuggestion(ctx context.Context, imageID uuid.UUID, userID string, suggestedFolderName string) error
```

`InitiateUpload` SHALL accept a `description *string` parameter:

```go
InitiateUpload(ctx context.Context, userID, title, mimeType string, sourceURL *string, folderID *uuid.UUID, description *string) (*UploadInitResult, error)
```

`UpdateImageParams` SHALL include a `Tags` field:

```go
type UpdateImageParams struct {
    Title       *string
    FolderID    **uuid.UUID
    Description *string
    SourceURL   **string
    Tags        *[]uuid.UUID  // nil = no change; non-nil (including empty slice) = replace tag set
}
```

`CompleteUploadResult` is defined in `internal/usecase/`:

```go
type CompleteUploadResult struct {
    ImageID             uuid.UUID
    SuggestedFolderName *string
    Warning             string
}
```

`ListImagesParams` SHALL include a `TagID` field:

```go
type ListImagesParams struct {
    FolderID *uuid.UUID
    Unfiled  bool
    TagID    *uuid.UUID
    Cursor   *ImageCursor
    Limit    int
}
```

When `UpdateImageParams.Tags` is non-nil, `UpdateImage` SHALL call `tagRepo.ReplaceImageTags` with the image ID and the given tag IDs after the scalar field update.

#### Scenario: UpdateImage replaces tags when Tags param is non-nil

- **WHEN** `UpdateImage` is called with `Tags` set to a slice of tag UUIDs
- **THEN** the image's tag associations are replaced with that set

#### Scenario: UpdateImage does not touch tags when Tags param is nil

- **WHEN** `UpdateImage` is called without a `Tags` param (nil)
- **THEN** the image's existing tag associations are unchanged

#### Scenario: UpdateImage clears all tags when Tags is an empty slice

- **WHEN** `UpdateImage` is called with `Tags` set to an empty slice
- **THEN** all tag associations for the image are removed

#### Scenario: ListImages passes TagID to repository

- **WHEN** `ListImages` is called with a non-nil `TagID` param
- **THEN** only images associated with that tag are returned

### Requirement: Image Response Types Include Tags

The system SHALL include tag data in image response types returned by the image handler.

```go
type tagResponse struct {
    ID   uuid.UUID `json:"id"`
    Name string    `json:"name"`
}

type imageResponse struct {
    ID           uuid.UUID     `json:"id"`
    Title        string        `json:"title"`
    Description  *string       `json:"description"`
    MIMEType     string        `json:"mime_type"`
    SourceURL    *string       `json:"source_url"`
    FolderID     *uuid.UUID    `json:"folder_id"`
    ThumbnailURL *string       `json:"thumbnail_url"`
    Width        *int          `json:"width"`
    Height       *int          `json:"height"`
    FileSize     *int64        `json:"file_size"`
    Tags         []tagResponse `json:"tags"`
    CreatedAt    time.Time     `json:"created_at"`
    UpdatedAt    time.Time     `json:"updated_at"`
}

type imageDetailResponse struct {
    // all fields of imageResponse, plus:
    ImageURL string `json:"image_url"`
}
```

`toImageResponse` SHALL map `item.Image.Tags` to `[]tagResponse`. If `Tags` is nil, it SHALL return an empty slice (never `null` in JSON).

#### Scenario: ListImages response includes tags array

- **WHEN** `GET /images` is called and images have associated tags
- **THEN** each image object in the response contains a `tags` array with `id` and `name` for each tag

#### Scenario: ListImages response has empty tags array for untagged images

- **WHEN** `GET /images` is called and an image has no tags
- **THEN** the image object contains `"tags": []`

#### Scenario: GetImage response includes tags

- **WHEN** `GET /images/:id` is called for an image with tags
- **THEN** the response contains a `tags` array with `id` and `name` for each associated tag

### Requirement: UpdateImage Accepts Tags Field

The system SHALL accept an optional `tags` field in `PATCH /images/:id` request body to replace the image's tag set.

```go
type updateImageRequest struct {
    Title       *string         `json:"title"`
    Description *string         `json:"description"`
    FolderID    json.RawMessage `json:"folder_id"`
    SourceURL   json.RawMessage `json:"source_url"`
    Tags        json.RawMessage `json:"tags"`
}
```

Parsing rules for `Tags`:
- Field absent or JSON `null` → `params.Tags` is nil (no change to tags)
- `[]` → `params.Tags` is a pointer to an empty slice (clear all tags)
- `["uuid1", "uuid2"]` → `params.Tags` is a pointer to the parsed UUID slice

The handler SHALL return `400 Bad Request` if a tag ID in the array is not a valid UUID.

#### Scenario: PATCH with tags replaces image tag set

- **WHEN** `PATCH /images/:id` is sent with `{"tags": ["<uuid>"]}`
- **THEN** the image's tags are replaced with the specified tag
- **AND** the response body contains the updated tags array

#### Scenario: PATCH without tags does not modify tags

- **WHEN** `PATCH /images/:id` is sent without a `tags` field
- **THEN** the image's existing tags are unchanged

#### Scenario: PATCH with empty tags array clears all tags

- **WHEN** `PATCH /images/:id` is sent with `{"tags": []}`
- **THEN** all tags are removed from the image

#### Scenario: PATCH with invalid tag UUID returns 400

- **WHEN** `PATCH /images/:id` is sent with `{"tags": ["not-a-uuid"]}`
- **THEN** the response is `400 Bad Request`

### Requirement: ListImages Accepts tag_id Filter

The system SHALL accept an optional `tag_id` query parameter on `GET /images` to filter images by a single tag.

#### Scenario: GET /images with tag_id returns only tagged images

- **WHEN** `GET /images?tag_id=<uuid>` is called
- **THEN** only images associated with that tag are returned

#### Scenario: GET /images with invalid tag_id returns 400

- **WHEN** `GET /images?tag_id=not-a-uuid` is called
- **THEN** the response is `400 Bad Request`

#### Scenario: GET /images without tag_id returns all images

- **WHEN** `GET /images` is called without a `tag_id` parameter
- **THEN** images are returned regardless of their tag associations
