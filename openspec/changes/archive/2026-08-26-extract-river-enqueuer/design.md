## Context

The async job layer currently has two structural problems that compound each other:

1. **Circular import forcing Args into the wrong package**: `worker` imports `usecase` for Args types. Moving Args to `worker` would require `usecase` to import `worker`, creating a cycle. The current workaround is `usecase/job_args.go` — Args defined in the usecase layer, which shouldn't own them.

2. **Generic enqueuer leaks River shape into the domain**: `JobEnqueuer.Insert(ctx, JobArgs)` and `JobArgs` with `Kind()` / `MaxAttempts()` mirror River's own interfaces. The usecase layer is aware of the job queue's mechanics rather than just expressing what work it wants done.

A secondary problem: `main.go` embeds the `riverEnqueuer` River adapter inline alongside all other wiring, and worker files are inconsistently split between a `periodic.go` dumping ground and individual files.

## Goals / Non-Goals

**Goals:**
- Remove River concepts from the usecase layer entirely
- Give Args a natural home in the worker package, co-located with the workers that consume them
- Mirror the repository pattern: usecase declares a narrow interface, infrastructure implements it
- Extract the River adapter out of `main.go` without introducing a new internal package
- Reorganize worker files by domain for navigability

**Non-Goals:**
- Changing any job behavior, retry logic, or scheduling
- Adding new workers or job types
- Touching the handler or domain layers

## Decisions

### Decision 1: Per-usecase enqueuer interfaces (not a shared jobargs package)

Each usecase that enqueues jobs declares its own narrow interface:

```go
// in account_usecase.go
type accountJobEnqueuer interface {
    EnqueueAccountWipe(ctx context.Context, userID string) error
    EnqueueAccountWipeUnique(ctx context.Context, userID string) error
    EnqueueBookletUserDeletion(ctx context.Context, userID string) error
    EnqueueR2Delete(ctx context.Context, r2Path string, thumbnailPath *string) error
}

// in image_upload_usecase.go
type imageUploadJobEnqueuer interface {
    EnqueueVision(ctx context.Context, imageID uuid.UUID, userID string) error
    EnqueueCategoriseImage(ctx context.Context, imageID uuid.UUID, userID string) error
}

// in trash_usecase.go
type trashJobEnqueuer interface {
    EnqueueR2Delete(ctx context.Context, r2Path string, thumbnailPath *string) error
}
```

**Why over a shared `jobargs` package**: A shared package still requires `usecase` to import it to construct Args, keeping River-shaped types (`Kind()`, `MaxAttempts()`) visible to the domain layer. Typed interfaces remove that entirely — usecase expresses domain intent, infrastructure decides how to enqueue it.

**Why over a single shared typed interface**: Tight scoping matches the repository pattern already established. Each usecase sees only what it can enqueue.

### Decision 2: Args move to the worker package, grouped by domain

With the circular import broken (usecase no longer imports worker), Args can live next to the workers that consume them:

```
internal/worker/
├── image.go           ← VisionArgs, CategoriseImageArgs, BackfillPhashArgs
│                         + VisionWorker, CategorisationWorker, BackfillPhashWorker
├── account_deletion.go ← AccountWipeArgs, BookletUserDeletionArgs,
│                          AccountWipeReconcileArgs, PurgedAccountSweepArgs
│                          + all four workers
└── cleanup.go         ← CleanupStaleUploadsArgs, TrashPurgeArgs, R2DeleteArgs
                          + CleanupStaleUploadsWorker, TrashPurgeWorker, R2DeleteWorker
```

`Kind()` and `MaxAttempts()` stay on Args — they are River infrastructure concerns and belong in the worker package.

**Why domain grouping over one-file-per-worker**: Most workers are thin wrappers. A flat file per worker creates noise without navigability benefit. Domain grouping means when working on the account deletion flow, all related workers and Args are in one file.

### Decision 3: `riverEnqueuer` extracted to `cmd/server/enqueuer.go`

`cmd/server/` is `package main` and can span multiple files. `enqueuer.go` holds the concrete River adapter — a single struct implementing all three per-usecase interfaces:

```go
// cmd/server/enqueuer.go
type riverEnqueuer struct {
    client *river.Client[pgx.Tx]
}

func (e *riverEnqueuer) EnqueueAccountWipe(ctx context.Context, userID string) error {
    args := worker.AccountWipeArgs{UserID: userID}
    _, err := e.client.Insert(ctx, args, &river.InsertOpts{MaxAttempts: args.MaxAttempts()})
    return err
}
// ... one method per job type
```

The deferred init pattern (`enqueuer.client = riverClient` after `river.NewClient`) is preserved unchanged — it's still one pointer on one struct.

**Why one struct over separate structs per domain**: The implementations are identical in shape. Separate structs would be three copies of the same `*river.Client[pgx.Tx]` field with no meaningful distinction.

**Why `cmd/server/` over a new `internal/` package**: This is app wiring, not domain logic. A new internal package would be a package boundary without a corresponding abstraction boundary. Keeping it in `cmd/server/` is honest about what it is.

## Risks / Trade-offs

- **`async-job-queue` spec references the old structure explicitly** (Args in `usecase/job_args.go`, generic `JobEnqueuer`) → The spec delta covers this; spec must be updated alongside the code.
- **One concrete struct implements three interfaces** → Satisfying multiple interfaces from one type is idiomatic Go and carries no risk, but it means `cmd/server/enqueuer.go` grows with every new job type. Acceptable at current scale.
- **Usecase tests require mock updates** → The typed interfaces make mocks simpler (no type assertions on inserted args), but it is mechanical churn across three test files.
