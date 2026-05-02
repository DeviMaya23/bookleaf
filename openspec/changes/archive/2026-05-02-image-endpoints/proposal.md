## Why

The image domain struct and DB migration already exist but there are no HTTP endpoints, upload flow, or storage integration. Users need to upload images to R2, browse their gallery with thumbnails, access full-resolution images via presigned URLs, and soft-delete images with a recoverable trash.

## What Changes

- Add a two-step upload flow: `POST /images` initiates the upload and returns a presigned PUT URL; `POST /images/:id/complete` notifies the backend the upload finished and triggers async thumbnail generation
- Add gallery list endpoint `GET /images` (filterable by folder) returning metadata and thumbnail CDN URLs
- Add detail endpoint `GET /images/:id` returning full metadata and a fresh 24-hour presigned GET URL
- Add soft delete `DELETE /images/:id`, trash list `GET /images/trash`, and restore `POST /images/:id/restore`
- Add R2 storage service for generating presigned PUT/GET URLs and constructing CDN URLs
- Add server-side thumbnail generation using `disintegration/imaging` (max 600×600), run in a goroutine after upload completion and stored back to R2

## Capabilities

### New Capabilities

- `image-endpoints`: HTTP endpoints for image upload flow, gallery, detail, soft delete, trash, and restore — all auth-gated
- `r2-storage`: R2 storage service for presigned PUT/GET URL generation and CDN URL construction
- `image-thumbnail`: Server-side thumbnail generation service using `disintegration/imaging`, invoked asynchronously via goroutine after upload completion

### Modified Capabilities

- None

## Impact

- New files: `internal/handler/image.go`, `internal/usecase/image_usecase.go`, `internal/usecase/image_repository.go`, `internal/repository/image_repository.go`, `internal/storage/r2.go`, `internal/thumbnail/thumbnail.go`
- `internal/config/config.go` — add R2 config fields (`R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET_NAME`, `R2_PUBLIC_URL`)
- `cmd/server/main.go` — wire R2 client, image repository, image usecase, image handler, register routes
- New Go dependencies: `github.com/disintegration/imaging`, `github.com/aws/aws-sdk-go-v2` (S3-compatible client for R2)
- New environment variables: `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET_NAME`, `R2_PUBLIC_URL`
