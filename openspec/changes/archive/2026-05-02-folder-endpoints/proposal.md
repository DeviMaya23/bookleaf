## Why

The folder domain struct and migration already exist, but there are no HTTP endpoints to manage folders. Users need a way to create, read, update, and delete their folders through the API.

## What Changes

- Add CRUD HTTP endpoints for folders under `/api/folders`, protected by Kinde auth middleware
- Add folder service layer with business logic for create, list, get, update, and delete
- Add folder SQL repository implementing the data access layer
- **BREAKING** Remove soft delete from the `Folder` domain struct and migration — folders are hard-deleted; on delete, child folders have their `parent_id` set to null and (when implemented) images in the folder have their `folder_id` set to null

## Capabilities

### New Capabilities

- `folder-endpoints`: HTTP CRUD endpoints for folders, auth-gated, with cascading null-out on delete

### Modified Capabilities

- `folder-domain`: Remove `DeletedAt` soft-delete field from the `Folder` struct and its corresponding DB migration column; switch to hard delete

## Impact

- `internal/domain/folder.go` — remove `DeletedAt gorm.DeletedAt` field
- `internal/migrations/` — update folders migration to drop `deleted_at` column
- New files: `internal/folder/service.go`, `internal/folder/service_test.go`, `internal/folder/handler.go`, `internal/folder/handler_test.go`, `internal/folder/repository.go`
- Router wiring in `internal/server/` to register folder routes with auth middleware
