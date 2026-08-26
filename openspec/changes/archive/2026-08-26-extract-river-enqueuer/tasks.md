## 1. Reorganize Worker Files by Domain

- [x] 1.1 Create `internal/worker/account_deletion.go` — define `AccountWipeArgs`, `BookletUserDeletionArgs`, `AccountWipeReconcileArgs`, `PurgedAccountSweepArgs` and their four workers (move from `account_wipe.go`, `booklet_user_deletion.go`, and `periodic.go`)
- [x] 1.2 Create `internal/worker/media.go` — define `VisionArgs`, `CategoriseImageArgs`, `R2DeleteArgs`, `BackfillPhashArgs` and their four workers (move from `vision.go`, `categorise.go`, `r2_delete.go`, and `periodic.go`)
- [x] 1.3 Create `internal/worker/maintenance.go` — define `CleanupStaleUploadsArgs`, `TrashPurgeArgs` and their two workers (move from `periodic.go`)
- [x] 1.4 Delete `internal/worker/account_wipe.go`, `booklet_user_deletion.go`, `r2_delete.go`, `vision.go`, `categorise.go`, `periodic.go`

## 2. Replace Generic Enqueuer with Per-Usecase Interfaces

- [x] 2.1 In `account_usecase.go`: add `accountJobEnqueuer` interface with `EnqueueAccountWipe`, `EnqueueAccountWipeUnique`, `EnqueueBookletUserDeletion`; update all three `enqueuer.Insert` / `enqueuer.InsertUnique` call sites to use the typed methods; update constructor signature
- [x] 2.2 In `image_upload_usecase.go`: add `imageJobEnqueuer` interface with `EnqueueVision`, `EnqueueCategoriseImage`; update both `enqueuer.Insert` call sites; update constructor signature
- [x] 2.3 In `trash_usecase.go`: add `trashJobEnqueuer` interface with `EnqueueR2Delete`; update the `enqueuer.Insert` call site; update constructor signature
- [x] 2.4 Delete `internal/usecase/job_args.go` and `internal/usecase/job_enqueuer.go`

## 3. Create River Adapter File

- [x] 3.1 Create `cmd/server/enqueuer.go` (`package main`) — define `riverEnqueuer` struct with `client *river.Client[pgx.Tx]`; implement one method per job type constructing worker Args and calling `river.Client.Insert`; implement `EnqueueAccountWipeUnique` with `UniqueOpts` matching the previous `InsertUnique` behavior
- [x] 3.2 Remove the `riverEnqueuer` type definition and its `Insert` / `InsertUnique` methods from `main.go`; verify deferred init (`enqueuer.client = riverClient`) and usecase wiring calls are unchanged

## 4. Update Tests

- [x] 4.1 Update `account_usecase_test.go` — replace `JobEnqueuer` mock with typed `accountJobEnqueuer` mock; remove type assertions on inserted args; assert directly on captured method arguments
- [x] 4.2 Update `image_upload_usecase_test.go` — replace `JobEnqueuer` mock with typed `imageJobEnqueuer` mock; remove type assertions; assert on captured arguments
- [x] 4.3 Update `trash_usecase_test.go` — replace `JobEnqueuer` mock with typed `trashJobEnqueuer` mock; remove type assertions; assert on captured arguments
- [x] 4.4 Update `internal/worker/categorise_test.go` — remove `usecase` import if it was only needed for Args; confirm `CategoriseImageArgs` is now referenced from the same package

## 5. Lint

- [x] 5.1 Run `golangci-lint run ./...` from the backend directory and fix any issues
