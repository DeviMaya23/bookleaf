## Why

Users currently can only organise images by folder, which forces a single-hierarchy structure. Tags give users a second, free-form axis for filtering and organising their library without having to restructure their folders.

## What Changes

- New `tags` table: user-scoped tags with a unique name per user
- New `image_tags` junction table: many-to-many relationship between images and tags
- New tag CRUD endpoints: create, list, rename, delete
- `GET /images` gains an optional `tag_id` query filter
- `GET /images` and `GET /images/:id` responses include the image's associated tags
- `PATCH /images/:id` accepts an optional `tags` field to replace the image's tag set

## Capabilities

### New Capabilities

- `tag-domain`: Tag GORM struct, image_tags junction, DB migrations, TagRepository interface and SQL implementation
- `tag-endpoints`: Tag handler (create, list, update/rename, delete), usecase, routes wiring

### Modified Capabilities

- `image-domain`: Image struct gains a `Tags []Tag` GORM association; ImageRepository `List` and `GetByID` signatures updated to preload tags
- `image-endpoints`: `imageResponse` and `imageDetailResponse` include tags; `UpdateImageParams` gains a `Tags` field; `ListImages` gains `tag_id` filter

## Impact

- New DB migrations (tags table + image_tags junction)
- `internal/domain/image.go` — add `Tags []Tag` association field
- `internal/domain/tag.go` — new file
- `internal/usecase/image_repository.go` — updated `List` and `GetByID` signatures (non-breaking: only adds preload behaviour)
- `internal/usecase/image_usecase.go` — `UpdateImageParams` gains `Tags **[]uuid.UUID`; `UpdateImage` calls tag association replacement
- `internal/repository/image_repository.go` — add `.Preload("Tags")` to fetch paths
- `internal/handler/image.go` — `imageResponse`/`imageDetailResponse` gain `Tags []tagResponse`; `updateImageRequest` gains `Tags` field; `ListImages` reads `tag_id` query param
- New files: `internal/domain/tag.go`, `internal/usecase/tag_repository.go`, `internal/usecase/tag_usecase.go`, `internal/repository/tag_repository.go`, `internal/handler/tag.go`
- `cmd/server/main.go` — wire tag repository, usecase, handler, routes
- Bruno collection — new request files for tag endpoints
