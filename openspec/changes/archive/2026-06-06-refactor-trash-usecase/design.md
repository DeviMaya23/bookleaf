## Context

`imageUsecase` currently owns two distinct concerns: image querying/editing (`ListImages`, `GetImage`, `DownloadImage`, `UpdateImage`, `MoveImageFolder`, `UpdateImagePosition`) and the full trash lifecycle (`SoftDelete`, `ListTrashed`, `Restore`, `PurgeExpiredTrash`, `DeleteFromTrash`, `EmptyTrash`). Trash is a coherent state machine — Active → Trash → Restored or Gone — with its own dependencies. Specifically, async R2 deletion via `JobEnqueuer` belongs with trash, not with image querying.

`EmptyTrash` currently performs sequential synchronous R2 deletes inside the request path, making it slow for large trash collections. R2 failures are logged but unrecoverable — the DB record is gone, the R2 path is lost.

`ImageHandler` similarly mixes both concerns. The handler split follows the usecase split.

Existing relevant pieces:
- `usecase.JobEnqueuer` interface with `Insert(ctx, JobArgs)` — used by `imageUploadUsecase`
- `usecase.VisionArgs` in `job_args.go` — pattern for River job args types
- `worker.TrashPurgeWorker` — currently depends on a local `trashUsecase` interface (coincidental naming); will shift its dependency to the new real `TrashUsecase`
- `riverEnqueuer` in `main.go` adapts `*river.Client[pgx.Tx]` to `JobEnqueuer`

## Goals / Non-Goals

**Goals:**
- Extract `trashUsecase` as a standalone usecase with its own interface and constructor
- Extract `TrashHandler` with its own `TrashUsecase` interface at the handler layer
- Add `JobEnqueuer` to `trashUsecase` only
- Make `EmptyTrash` return after DB commit; R2 deletion happens asynchronously via a new `R2DeleteWorker`
- `DeleteFromTrash` remains synchronous — single image, single R2 call, acceptable latency

**Non-Goals:**
- Transactional enqueueing (River `InsertTx`) — the GORM/pgx pool split makes this a new architectural pattern; the non-transactional gap (crash between DB commit and enqueue) is accepted
- Changing any API surface, response shapes, or route paths
- Frontend changes

## Decisions

### 1. `trashUsecase` takes the same `imageRepo` and `store` as `imageUsecase`

Both usecases depend on `ImageRepository` and `StorageService`. There is no new dependency — the interfaces are already defined in the usecase package. The constructor signature mirrors `imageUsecase`: `NewTrashUsecase(imageRepo ImageRepository, store StorageService, enqueuer JobEnqueuer, tel *observability.Telemetry)`.

**Alternatives considered:**
- *Single usecase with two constructors*: doesn't solve the coherence problem, just renames things.

### 2. `TrashHandler` owns `DELETE /images/:id` (SoftDelete) and `POST /images/:id/restore` (Restore)

These routes sit at `/images/:id` paths, not under `/images/trash`, but they are trash state transitions. If they stayed in `ImageHandler`, `ImageHandler` would need to depend on both `imageUsecase` and `trashUsecase` — recreating the mixing problem at the handler layer. Assigning them to `TrashHandler` keeps each handler to a single usecase dependency. The URL is a routing concern, not a domain boundary.

**Alternatives considered:**
- *Keep SoftDelete/Restore in ImageHandler, add trashUsecase dependency*: dual dependency on ImageHandler defeats the point of splitting.
- *Keep SoftDelete in imageUsecase*: imageUsecase would still touch `deleted_at` and trash state, leaking the boundary.

### 3. `EmptyTrash`: hard-delete DB records synchronously, enqueue River jobs for R2 deletion

Ordering: collect image paths via `ListAllTrashed` → hard-delete all DB records (GORM) → loop: `enqueuer.Insert(ctx, R2DeleteArgs{r2Path, thumbnailPath})` per image → return 204.

DB delete happens first so the user's data is removed immediately and authoritatively. R2 jobs are enqueued after commit. The gap — process crash between DB commit and enqueue — is accepted. It produces the same outcome as the current synchronous failure mode (orphaned R2 objects), but is far rarer than transient R2 network errors, which are now handled by River retry.

**Alternatives considered:**
- *Enqueue first, then DB delete*: if DB delete fails, River jobs run and delete R2 objects while DB records still exist — broken image references in DB. Worse outcome than the accepted gap.
- *River `InsertTx` for atomicity*: requires `pgx.Tx` which is incompatible with GORM's `database/sql` transaction. Would introduce pgx transactions at the repository layer — a new architectural pattern not currently in the codebase.
- *Outbox table*: GORM-native atomicity via a `pending_r2_deletes` table. Correct, but adds a migration and a periodic drainer. Heavier than needed at current scale.

### 4. `R2DeleteArgs` lives in `job_args.go`; `R2DeleteWorker` calls `trashUsecase.ProcessR2Delete`

Following the existing pattern: `VisionArgs` is defined in `usecase/job_args.go`; the worker in `internal/worker/` depends on a narrow usecase interface. `R2DeleteArgs` carries `R2Path string` and `ThumbnailPath *string` — just strings, no bytes. The worker calls a new `ProcessR2Delete(ctx, r2Path string, thumbnailPath *string) error` method on `trashUsecase`, which calls `store.DeleteObject` for each path. Max 5 attempts, default backoff.

**Alternatives considered:**
- *Worker calls storage directly*: bypasses the usecase layer, breaks the clean architecture boundary (handler → usecase → repo/storage).

### 5. `TrashPurgeWorker`'s local `trashUsecase` interface in `periodic.go` becomes the real dependency

`periodic.go` already defines a local `trashUsecase` interface with `PurgeExpiredTrash`. Once the real `TrashUsecase` interface exists in the usecase package, `TrashPurgeWorker` in `main.go` can be registered with the real `trashUsecase` instance. The local interface in `periodic.go` remains (it's the correct pattern — workers define their own narrow interface); the wiring in `main.go` simply passes the new `trashUsecase`.

## Risks / Trade-offs

- **Non-transactional DB delete → enqueue gap**: if the process crashes between `HardDelete` commit and `enqueuer.Insert`, R2 objects are orphaned with no recovery path. Accepted as a rare edge case; materially better than the current state where any transient R2 error in the request path produces the same outcome with no retry.
- **`EmptyTrash` no longer waits for R2 cleanup**: callers receive 204 before storage is reclaimed. This is invisible to the user (R2 is an implementation detail) and consistent with the async pattern already used for thumbnail generation and vision labelling.
- **Two usecases share `ImageRepository`**: not a problem — the interface is already defined in the usecase package and both usecases depend on the same interface, not a concrete type. GORM's connection pool handles concurrent access.
