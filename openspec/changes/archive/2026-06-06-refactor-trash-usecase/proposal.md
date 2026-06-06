## Why

`imageUsecase` has grown to own two distinct concerns — image querying/editing and trash lifecycle management — making it harder to reason about and a poor fit for the `JobEnqueuer` dependency that async trash deletion requires. Separating trash into its own usecase gives the lifecycle a clean home and enables `EmptyTrash` to return fast by pushing R2 deletion to a retryable background worker.

## What Changes

- New `TrashUsecase` interface and `trashUsecase` implementation extracted from `imageUsecase`, owning the full trash state machine: `SoftDelete`, `ListTrashed`, `Restore`, `PurgeExpiredTrash`, `DeleteFromTrash`, `EmptyTrash`
- `imageUsecase` retains only image querying and editing: `ListImages`, `GetImage`, `DownloadImage`, `UpdateImage`, `MoveImageFolder`, `UpdateImagePosition`
- New `TrashHandler` extracted from `ImageHandler`, handling all trash lifecycle routes including `DELETE /images/:id` (SoftDelete) and `POST /images/:id/restore` (Restore)
- `EmptyTrash` changed: hard-deletes all DB records synchronously, then enqueues a River job per image for async R2 deletion — response returns after DB commit, not after R2 cleanup
- New `R2DeleteWorker` River worker handles async R2 object deletion with retry
- `TrashPurgeWorker` registration shifts from `imageUsecase` to `trashUsecase`
- `JobEnqueuer` dependency added to `trashUsecase` only

## Capabilities

### New Capabilities

None — this change is a structural refactor. No new user-facing capabilities are introduced.

### Modified Capabilities

- `manual-trash-delete`: `EmptyTrash` behavior changes — R2 deletion becomes asynchronous (River job per image, with retry); hard-delete of DB records remains synchronous and determines the 204 response. A new `R2DeleteWorker` handles R2 cleanup. `DeleteFromTrash` remains synchronous (single image, single R2 call).
- `trash-purge`: `PurgeExpiredTrash` moves from `ImageUsecase` to a new `TrashUsecase` interface; behavior unchanged.

## Impact

- `backend/internal/usecase/image_usecase.go` — remove trash methods; remove `ImageUsecase` interface entries for trash operations
- `backend/internal/usecase/trash_usecase.go` — new file; `TrashUsecase` interface + `trashUsecase` implementation
- `backend/internal/usecase/image_repository.go` — `ImageRepository` interface is shared; no change needed
- `backend/internal/handler/image.go` — remove trash handler methods and `ImageUsecase` interface entries for trash
- `backend/internal/handler/trash.go` — new file; `TrashHandler` + `TrashUsecase` interface for handler layer
- `backend/internal/worker/r2_delete.go` — new River worker for async R2 deletion
- `backend/internal/worker/periodic.go` — `TrashPurgeWorker` shifts its usecase dependency to `TrashUsecase`
- `backend/cmd/server/main.go` — wire `trashUsecase`, `TrashHandler`; register `R2DeleteWorker`; update `TrashPurgeWorker` dependency
- No API surface changes — all existing routes and response shapes are preserved
