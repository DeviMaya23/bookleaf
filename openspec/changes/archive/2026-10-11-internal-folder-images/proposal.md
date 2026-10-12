# Proposal

## Why

Booklet needs to display image thumbnails for a user's folder without requiring the user to navigate into the folder. A dedicated internal endpoint avoids round-tripping through the share layer and exposes exactly the data Booklet needs.

## What Changes

- Add `GET /internal/users/:user_id/folders/:folder_id/images` to the internal route group
- Add a new usecase interface and method on `InternalHandler` for listing folder image thumbnails
- Return a flat list of `{ image_id, thumbnail_url }` for all images in the folder that have a thumbnail; images without a thumbnail path are excluded
- Folder not found or not owned by the given user → `404`; empty folder (or all images lack thumbnails) → `200` with `{ "images": [] }`
- Add a Bruno request file for the new endpoint

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `internal-folder-api`: add the new `GET /internal/users/:user_id/folders/:folder_id/images` route, its handler method, usecase interface, and Bruno file to the existing internal API capability

## Impact

- `backend/internal/handler/internal.go` — new handler method `ListFolderImages`; new `InternalFolderImagesUsecase` interface wired into `InternalHandler`
- `backend/internal/usecase/` — new focused usecase (or method) implementing ownership check + thumbnail presign
- `backend/cmd/server/main.go` — one new route registration in the existing `/internal` group
- `bruno/internal/` — new `.bru` file
- No frontend or extension changes
