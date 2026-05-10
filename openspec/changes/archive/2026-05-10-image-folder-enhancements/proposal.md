## Why

Images and folders currently lack descriptive metadata and dimensional information, limiting how users can annotate and browse their library. Surfacing image counts on folders also reduces the need for extra round-trips when rendering folder views.

## What Changes

- **Image domain**: add `description` (nullable text), `width`/`height` (integer pixels, nullable), and `file_size` (integer bytes, nullable) fields to the `Image` struct and `images` table; all three are calculated server-side at upload completion time
- **Image endpoints**: `POST /images` (create) and `PATCH /images/:id` (update) accept the new `description` field; `GET /images/:id` and `GET /images` responses include `description`, `width`, `height`, and `file_size`; `POST /images/:id/complete` populates dimensions and file size by reading the uploaded file from R2
- **Folder domain**: add `description` (nullable text) field to the `Folder` struct and `folders` table
- **Folder endpoints**: `POST /folders` and `PUT /folders/:id` accept `description`; `GET /folders/:id` and `GET /folders` responses include `description`; `GET /folders/:id` also returns an `image_count` field with the number of non-deleted images in that folder

## Capabilities

### New Capabilities

*(none)*

### Modified Capabilities

- `image-domain`: add `description`, `width`, `height`, and `file_size` fields to `Image` struct and migration
- `image-endpoints`: request/response shapes updated for description on create/update; response shapes updated to include description, width, height, file_size; complete-upload populates all three BE-side
- `folder-domain`: add `description` field to `Folder` struct and migration
- `folder-endpoints`: request/response shapes updated for description on create/update; GET /folders/:id response includes `image_count`; GET /folders list response includes `description`

## Impact

- New SQL migration(s) adding nullable columns to `images` and `folders`
- `domain.Image` and `domain.Folder` structs updated
- `ImageRepository` and `FolderRepository` interfaces may need additional methods (e.g., count images by folder)
- `imageUsecase.CompleteUpload` reads image bytes from R2, extracts width/height from image metadata, and measures byte length for file_size — all persisted in one update
- Request/response DTOs and handler logic updated for both domains
- Existing unit tests for image and folder handlers/usecases will need to be extended to cover new fields
