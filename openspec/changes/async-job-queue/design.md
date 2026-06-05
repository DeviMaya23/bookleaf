## Context

The backend currently handles post-upload work in two ways: a fire-and-forget goroutine for thumbnail R2 upload (no retry, silent failure), and synchronous execution of vision labelling inside `CompleteUpload` (blocks the response, returns `suggested_folder_name` directly). Two maintenance workers run on tickers started in `main.go`, whose interface and composite wrapper also live there.

This design replaces all of that with a River-backed job queue using the existing Postgres database — no new infrastructure required.

## Goals / Non-Goals

**Goals:**
- Replace the thumbnail goroutine with a retryable River job (5 attempts)
- Move vision labelling out of the request path into a retryable River job (3 attempts, ~10s between retries)
- Replace the two ticker workers with River periodic jobs
- Clean up `main.go`: remove `compositeImageWorker`, `imageWorkerUsecase`, `startWorkers`
- Establish `internal/worker/` as the home for all background job workers

**Non-Goals:**
- Real-time FE notification of vision results (deferred; FE polling or SSE is a separate change)
- Dead-letter queue or alerting on exhausted jobs
- Job introspection UI

## Decisions

### 1. River over a hand-rolled job table

River provides Postgres-backed persistence, typed job args, retry with configurable backoff, periodic jobs, graceful shutdown, and advisory lock-based concurrency — all things a hand-rolled `jobs` table would require implementing. The only cost is a DB migration and a new dependency. Given the existing Postgres dependency, this is the right trade.

**Alternatives considered:**
- *Roll our own*: more code, same persistence guarantee, worse retry and scheduling primitives.
- *Redis/external queue*: introduces new infra with no benefit at this scale.

### 2. Dedicated `*pgxpool.Pool` for River, separate from GORM

River's pgx driver requires a `*pgxpool.Pool`. GORM's internal connection pool is not exposed as a `pgxpool.Pool`. Rather than pulling GORM internals out, a second pool is opened against the same `DB.URL` from config. River's pool is sized small (1–3 connections) since background workers are low-throughput.

**Alternatives considered:**
- *Extract from GORM*: fragile, couples to GORM internals.
- *Use `riverdatabasesql` driver with GORM's `*sql.DB`*: works but loses pgx-specific optimisations River is built around.

### 3. River migration as file `000012` in the existing migrations directory

River ships its schema as versioned SQL. Adding it as `000012_river_schema.up/down.sql` keeps all DB state changes in one place, consistent with the existing `golang-migrate` approach. The up migration runs River's schema creation; the down migration drops River's tables.

**Alternatives considered:**
- *Run River's migration programmatically at startup*: changes are harder to track, doesn't fit the existing pattern.

### 4. Workers call usecase methods, not repos directly

Workers in `internal/worker/` depend on narrow usecase interfaces — the same pattern as handlers. Domain logic stays in usecases; workers are just delivery mechanisms. Two new methods are added to `imageUploadUsecase`:
- `ProcessThumbnailUpload(ctx, imageID uuid.UUID, r2Path, thumbnailKey string) error`
- `ProcessVisionLabelling(ctx, imageID uuid.UUID, userID string) error`

The existing `uploadThumbnail` goroutine method and `runVisionFlow` method become the bodies of these (unexported logic promoted to exported usecase methods).

**Alternatives considered:**
- *Workers inject repos/storage directly*: flattens the clean architecture boundary for no gain.

### 5. Thumbnail job re-fetches and re-generates from R2; no bytes in job args

Job payloads are persisted to Postgres. Thumbnail bytes (potentially hundreds of KB) must not be stored there. The `ThumbnailUploadArgs` carries only `ImageID`, `UserID`, `R2Path`, and `ThumbnailKey`. The worker re-fetches the original from R2 and re-generates the thumbnail on each attempt. This is safe because the original is always present in R2 at this point.

### 6. Vision job uses the original R2 image, not the thumbnail

The vision and thumbnail jobs run concurrently after `CompleteUpload` enqueues both. The vision worker cannot rely on the thumbnail being available yet. It fetches the original image from R2 directly. The Vision API handles full-size images without issue; smaller thumbnails were an optimisation, not a requirement.

`VisionArgs` carries `ImageID`, `UserID`, and `R2Path`.

### 7. Vision retry schedule: fixed 10s delay, not exponential backoff

River's default retry uses exponential backoff. For vision, the failure mode is typically a transient API error or rate limit — a short fixed wait is appropriate rather than a growing backoff. The worker implements `NextAttemptScheduledAt` to return `now + 10s` regardless of attempt number. Max 3 attempts total.

### 8. `CompleteUpload` response simplified

`suggested_folder_name` and `warning` are removed from the `CompleteUploadResult`. The `CompleteUploadResult` struct retains only `ImageID`. `AcceptSuggestion` endpoint is unchanged — it remains available for when FE wires up to the async vision result.

### 9. River client shutdown added to `server.shutdown()`

The `server` struct gains a `riverClient` field. `server.shutdown()` calls `riverClient.Stop(ctx)` alongside Echo shutdown and telemetry shutdown, ensuring in-flight jobs drain before the process exits.

## Risks / Trade-offs

**Two Postgres connection pools** → slightly higher idle connection count. Mitigated by sizing River's pool to 2–3 connections; well within the existing `SetMaxOpenConns(5)` ceiling... actually River needs its own pool separate from GORM, so effective max open connections is `5 (GORM) + 3 (River) = 8`. The DB pool ceiling on GORM needs to remain independent. This is acceptable.

**Re-fetching original from R2 on each thumbnail retry** → a failed attempt costs an R2 GET. With max 5 attempts this is negligible.

**Vision job runs against full-size image** → slightly more data sent to the Vision API than before. Acceptable trade for removing the thumbnail-job dependency.

**River schema migration must run before deployment** → standard migration gate; no different from any other migration. Rollback drops River tables (jobs are lost, but in-flight jobs are low-risk since workers aren't running).

## Migration Plan

1. Add `000012_river_schema.up.sql` and `.down.sql` using River's published schema SQL
2. Run migration in the normal deploy flow (`golang-migrate up`) before the new binary starts
3. New binary starts River client; old goroutine-based workers are gone
4. Rollback: deploy previous binary, run `000012` down migration (drops River tables)
