## Why

Bookleaf has no backend structure yet. Before any domain models, handlers, or business logic can be written, the project needs a clean architecture directory layout and a working Go entry point that wires Echo together.

## What Changes

- `go.mod` and skeleton `go.sum` are initialized
- Clean architecture directory tree is created under `internal/`
- `cmd/server/main.go` bootstraps Echo and starts the HTTP server
- Project is runnable (`go run ./cmd/server`) with a health-check route

## Capabilities

### New Capabilities
- `server-bootstrap`: A runnable Go HTTP server with Echo, clean architecture folder structure in place

### Modified Capabilities

## Impact

- `cmd/server/main.go` — new, entry point
- `internal/domain/`, `internal/usecase/`, `internal/repository/`, `internal/handler/` — new directories
- `go.mod` — new
