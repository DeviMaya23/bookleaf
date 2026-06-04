## Why

`imageUsecase` conflates two distinct concerns: image management (CRUD, listing, folders, trash) and the upload lifecycle (initiate → complete → accept suggestion, plus stale cleanup). These have different dependencies, different infrastructure concerns, and grow independently — keeping them together makes both harder to reason about and test.

## What Changes

- Extract `InitiateUpload`, `CompleteUpload`, `AcceptSuggestion`, and `CleanupStaleUploads` from `imageUsecase` into a new `imageUploadUsecase`
- Move upload-specific dependencies (`pendingUploadRepo`, `thumbnails`, `visionService`, `userRepo`, upload metrics) to the new usecase
- Create a new `UploadHandler` in `internal/handler/` for the three upload endpoints
- Remove upload methods from the `ImageUsecase` interface in the handler layer; add a new `UploadUsecase` interface
- Register both handlers under the same `/images/...` path prefix in the router — no client-visible change

## Capabilities

### New Capabilities

None. This is a pure internal refactor. No new API endpoints, no behavior changes.

### Modified Capabilities

None. Existing API contracts are unchanged. No spec-level behavior changes.

## Impact

- `backend/internal/usecase/image_usecase.go` — upload methods removed
- `backend/internal/usecase/image_usecase_test.go` — upload test cases removed
- `backend/internal/usecase/image_upload_usecase.go` — new file
- `backend/internal/usecase/image_upload_usecase_test.go` — new file
- `backend/internal/handler/image.go` — `ImageUsecase` interface trimmed; upload handler methods removed
- `backend/internal/handler/image_test.go` — upload handler test cases removed
- `backend/internal/handler/image_upload.go` — new file
- `backend/internal/handler/image_upload_test.go` — new file
- `backend/internal/server/` — router wiring updated to register `UploadHandler`
