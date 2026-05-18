## MODIFIED Requirements

### Requirement: Image Usecase Interface

`Restore` and `UpdateImage` SHALL return `*ImageItem` instead of `*domain.Image`:

```go
Restore(ctx context.Context, id uuid.UUID, userID string) (*ImageItem, error)
UpdateImage(ctx context.Context, id uuid.UUID, userID string, params UpdateImageParams) (*ImageItem, error)
```

`ImageItem` is defined in `internal/usecase/`:

```go
type ImageItem struct {
    Image        *domain.Image
    ThumbnailURL *string
}
```

`ListImagesResult.Images` and `ListTrashedResult.Images` SHALL be `[]ImageItem`.

`ImageDetail` SHALL include a `ThumbnailURL *string` field alongside `ImageURL`:

```go
type ImageDetail struct {
    Image        *domain.Image
    ImageURL     string
    ThumbnailURL *string
}
```

All other method signatures are unchanged.

#### Scenario: Usecase interface is satisfied by concrete implementation

- **WHEN** the Go package is compiled
- **THEN** `imageUsecase` implements `usecase.ImageUsecase` without compilation errors

---

### Requirement: Thumbnail URL Generation

The usecase SHALL generate presigned GET URLs for thumbnails with a 24h TTL. A private helper `thumbnailURL(ctx context.Context, path *string) *string` on `imageUsecase` SHALL:

- Return `nil` if `path` is `nil`
- Call `store.GeneratePresignedGetURL` with `presignedGetTTL` (24h)
- Return `nil` if presigning fails (non-fatal; thumbnail is cosmetic)

This helper SHALL be called by `ListImages`, `ListTrashed`, `GetImage`, `Restore`, and `UpdateImage`.

#### Scenario: Thumbnail URL is presigned when thumbnail path exists

- **WHEN** an image has a non-nil `thumbnail_path`
- **THEN** the response includes a non-nil `thumbnail_url` containing a presigned GET URL

#### Scenario: Thumbnail URL is nil when no thumbnail exists

- **WHEN** an image has a nil `thumbnail_path`
- **THEN** `thumbnail_url` in the response is `null`

#### Scenario: Thumbnail URL is nil when presigning fails

- **WHEN** `GeneratePresignedGetURL` returns an error for the thumbnail key
- **THEN** `thumbnail_url` in the response is `null`
- **AND** the overall request succeeds

---

### Requirement: GET /images and GET /images/:id — Response Shape

`thumbnail_url` in all image responses is now a presigned GET URL (24h TTL) rather than a public CDN URL. The field type and nullability are unchanged.

`GET /images/:id` response (`imageDetailResponse`) gains a `thumbnail_url` field sourced from `ImageDetail.ThumbnailURL`.

#### Scenario: thumbnail_url is a presigned URL

- **WHEN** an authenticated `GET /images` or `GET /images/:id` request is made for an image with a thumbnail
- **THEN** `thumbnail_url` is a non-null string containing a presigned R2 GET URL
- **AND** the URL is valid for 24 hours from the time of the request
