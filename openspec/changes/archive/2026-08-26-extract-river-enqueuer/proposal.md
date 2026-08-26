## Why

The async job layer has accumulated structural debt: job Args live in the `usecase` package to avoid a circular import, the generic `JobEnqueuer` interface leaks River's shape into the domain layer, worker files are split inconsistently between a `periodic.go` dumping ground and individual files, and `main.go` houses the River adapter inline alongside all other app wiring.

## What Changes

- Replace the generic `JobEnqueuer` / `JobArgs` interfaces in `usecase` with tight-scoped enqueuer interfaces per usecase (matching the repository pattern)
- Move job Args from `usecase/job_args.go` into the worker package, co-located with their workers
- Reorganize `internal/worker/` files by domain: `media.go`, `account_deletion.go`, `maintenance.go`
- Extract `riverEnqueuer` from `main.go` into `cmd/server/enqueuer.go`
- Delete `usecase/job_args.go` and `usecase/job_enqueuer.go`

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `async-job-queue`: Worker package structure changes (domain-grouped files, Args move to worker package); enqueuer interface contract changes from generic `JobEnqueuer` to per-usecase typed interfaces; `cmd/server/enqueuer.go` introduced as the River adapter.

## Impact

- `cmd/server/main.go` — `riverEnqueuer` struct and methods removed; wiring unchanged otherwise
- `cmd/server/enqueuer.go` — new file in `package main`; concrete River adapter implementing all per-usecase enqueuer interfaces
- `usecase/job_args.go` — deleted
- `usecase/job_enqueuer.go` — deleted
- `usecase/account_usecase.go` — `JobEnqueuer` replaced with `accountJobEnqueuer` interface; constructor signature updated
- `usecase/image_upload_usecase.go` — `JobEnqueuer` replaced with `imageJobEnqueuer` interface; constructor signature updated
- `usecase/trash_usecase.go` — `JobEnqueuer` replaced with `trashJobEnqueuer` interface; constructor signature updated
- `internal/worker/` — three domain files replace six individual/periodic files; Args defined inline per domain file
- Usecase unit tests — enqueuer mocks updated to match typed interfaces (no type assertions on inserted args)
