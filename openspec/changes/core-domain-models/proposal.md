## Why

Bookleaf needs core domain models before any feature work (upload, gallery, categorisation) can begin. Defining `Image`, `Folder`, and their ownership model now establishes the data contract that every subsequent layer — repository, handler, usecase — will depend on.

## What Changes

- GORM structs for `User`, `Folder`, and `Image` added to `internal/domain/`
- `golang-migrate` added as a dependency
- SQL migration files created in `migration/` for all three tables
- Folders support arbitrary nesting via a self-referencing parent FK
- Images reference their owning folder and user; actual file bytes live in the user's R2 bucket (no sidecar JSON in storage)

## Capabilities

### New Capabilities
- `image-domain`: GORM struct and DB schema for the `images` table — title, source URL, R2 path, folder, user ownership
- `folder-domain`: GORM struct and DB schema for the `folders` table — name, user ownership, optional parent folder (nested hierarchy)

### Modified Capabilities

## Impact

- `internal/domain/` — new `image.go`, `folder.go`, `user.go`
- `migration/` — new directory with SQL migration files
- `go.mod` / `go.sum` — `golang-migrate` added
