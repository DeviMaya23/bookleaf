## 1. TrashUsecase

- [x] 1.1 Create `backend/internal/usecase/trash_usecase.go` — define `TrashUsecase` interface and `trashUsecase` struct with `imageRepo ImageRepository`, `store StorageService`, `enqueuer JobEnqueuer`, `tel *observability.Telemetry`; add `NewTrashUsecase` constructor
- [x] 1.2 Move `SoftDelete` from `image_usecase.go` to `trash_usecase.go` — no logic change
- [x] 1.3 Move `ListTrashed` from `image_usecase.go` to `trash_usecase.go` — no logic change
- [x] 1.4 Move `Restore` from `image_usecase.go` to `trash_usecase.go` — no logic change
- [x] 1.5 Move `PurgeExpiredTrash` from `image_usecase.go` to `trash_usecase.go` — no logic change
- [x] 1.6 Move `DeleteFromTrash` from `image_usecase.go` to `trash_usecase.go` — no logic change
- [x] 1.7 Rewrite `EmptyTrash` in `trash_usecase.go` — `ListAllTrashed` → hard-delete all DB records per image → enqueue `R2DeleteArgs` job per image → return nil; log warn on enqueue failure and continue
- [x] 1.8 Add `ProcessR2Delete(ctx context.Context, r2Path string, thumbnailPath *string) error` to `trashUsecase` — calls `store.DeleteObject` for `r2Path`; calls `store.DeleteObject` for `thumbnailPath` if non-nil; returns error on any failure
- [x] 1.9 Remove all trash methods (`SoftDelete`, `ListTrashed`, `Restore`, `PurgeExpiredTrash`, `DeleteFromTrash`, `EmptyTrash`) and `ListAllTrashed` from `image_usecase.go` and from the `ImageUsecase` interface in `handler/image.go`

## 2. Job Args

- [x] 2.1 Add `R2DeleteArgs` to `backend/internal/usecase/job_args.go` — `R2Path string`, `ThumbnailPath *string`; `Kind() = "r2_delete"`; `MaxAttempts() = 5`

## 3. R2DeleteWorker

- [x] 3.1 Create `backend/internal/worker/r2_delete.go` — define local `r2DeleteUsecase` interface with `ProcessR2Delete(ctx, r2Path string, thumbnailPath *string) error`; implement `R2DeleteWorker` using `river.WorkerDefaults[usecase.R2DeleteArgs]`; `Work` calls `usecase.ProcessR2Delete(ctx, job.Args.R2Path, job.Args.ThumbnailPath)`

## 4. TrashHandler

- [x] 4.1 Create `backend/internal/handler/trash.go` — define local `TrashUsecase` interface with `SoftDelete`, `ListTrashed`, `Restore`, `DeleteFromTrash`, `EmptyTrash`; implement `TrashHandler` struct with `NewTrashHandler` constructor
- [x] 4.2 Move `SoftDelete` handler from `handler/image.go` to `handler/trash.go` — no logic change
- [x] 4.3 Move `ListTrashed` handler from `handler/image.go` to `handler/trash.go` — no logic change
- [x] 4.4 Move `Restore` handler from `handler/image.go` to `handler/trash.go` — no logic change
- [x] 4.5 Move `DeleteFromTrash` handler from `handler/image.go` to `handler/trash.go` — no logic change
- [x] 4.6 Move `EmptyTrash` handler from `handler/image.go` to `handler/trash.go` — no logic change
- [x] 4.7 Remove moved handler methods and all trash-related entries from `ImageUsecase` interface in `handler/image.go`

## 5. Wiring

- [x] 5.1 In `backend/cmd/server/main.go` — instantiate `trashUsecase` via `NewTrashUsecase(imageRepository, storageService, enqueuer, tel)`; instantiate `TrashHandler` via `NewTrashHandler(trashUsecase, tel)`
- [x] 5.2 Register `R2DeleteWorker` with River — `river.AddWorker(workers, worker.NewR2DeleteWorker(trashUsecase))`
- [x] 5.3 Update `TrashPurgeWorker` registration — pass `trashUsecase` instead of `imageUsecase`
- [x] 5.4 Update route registrations — move `DELETE /images/:id`, `POST /images/:id/restore`, `GET /images/trash`, `DELETE /images/trash`, `DELETE /images/trash/:id` from `imageHandler` to `trashHandler`

## 6. Tests

- [x] 6.1 Create `backend/internal/usecase/trash_usecase_test.go` — unit tests for `trashUsecase`:
  - `EmptyTrash`: success (images found, DB deleted, jobs enqueued); no-op (no trashed images)
  - `DeleteFromTrash`: success; not-found (image not in trash); R2 failure does not block DB hard-delete
  - `ProcessR2Delete`: success with thumbnail; success without thumbnail; returns error on storage failure
- [x] 6.2 Remove trash method tests from `backend/internal/usecase/image_usecase_test.go`
- [x] 6.3 Create `backend/internal/handler/trash_test.go` — unit tests for `TrashHandler`:
  - `SoftDelete`: 204 on success; 404 on not-found
  - `ListTrashed`: 200 on success
  - `Restore`: 200 on success; 404 on not-found
  - `DeleteFromTrash`: 204 on success; 404 on not-found; 400 on malformed UUID
  - `EmptyTrash`: 204 on success
- [x] 6.4 Remove trash handler tests from `backend/internal/handler/image_test.go`
