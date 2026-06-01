## 1. Move platform packages

- [ ] 1.1 Create `internal/platform/` directory and move `internal/config/` to `internal/platform/config/`; update all import paths; verify `go build ./...` passes
- [ ] 1.2 Move `internal/observability/` to `internal/platform/observability/`; update all import paths; verify `go build ./...` passes

## 2. Move middleware under handler

- [ ] 2.1 Move `internal/middleware/` to `internal/handler/middleware/`; update all import paths; verify `go build ./...` passes

## 3. Move thumbnail to pkg

- [ ] 3.1 Create `pkg/thumbnail/` and move `internal/thumbnail/thumbnail.go` there; update all import paths; verify `go build ./...` passes

## 4. Move interfaces to usecase (consumer side)

- [ ] 4.1 Add `StorageService` interface to a new file `internal/usecase/image_storage.go`; remove the interface declaration from `internal/storage/storage.go`; update all references; verify `go build ./...` passes
- [ ] 4.2 Inline `ThumbnailService` and `VisionService` interface declarations at the top of `internal/usecase/image_usecase.go`; remove them from `pkg/thumbnail/thumbnail.go` and `internal/vision/vision.go`; update all references; verify `go build ./...` passes

## 5. Verify

- [ ] 5.1 Run `go build ./...` from `backend/` — zero errors
- [ ] 5.2 Run `go test ./...` from `backend/` — all tests pass
