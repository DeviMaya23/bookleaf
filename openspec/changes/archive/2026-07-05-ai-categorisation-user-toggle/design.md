## Context

`ai_categorisation_enabled` is stored on the `users` table and checked in the image upload usecase before enqueuing the River categorisation job. The feature is fully functional but gated — the flag defaults to `false` and can only be flipped via direct DB access.

The `ai_categorisation_logs` table already records every categorisation run with `user_id` and `created_at`, which is all that's needed to count monthly usage. No schema migration is required.

The `CategorisationUsecase` is constructed in `main.go` before `MeHandler`, making it straightforward to inject as an additional dependency.

## Goals / Non-Goals

**Goals:**
- Let users toggle `ai_categorisation_enabled` themselves via `PATCH /me`
- Expose the monthly usage count in `GET /me` so the UI can show "X / 50 this month"
- Silently enforce the 50-image monthly limit in the `CategoriseImage` usecase before calling the agent
- Show a dismissable red badge on the profile avatar when the limit is hit

**Non-Goals:**
- Per-user configurable limits (hardcoded at 50 for now)
- Notifying users via SSE or toast when the limit is hit
- Tracking token counts (only run counts)
- Resetting or carrying over unused quota

## Decisions

### Limit enforcement in `CategoriseImage`, not the worker

The River worker calls `usecase.CategoriseImage`. The limit check belongs in the usecase — it's business logic, not job orchestration. The worker returns `nil` regardless (job is considered "done" whether it ran or was skipped), which prevents River from retrying what is intentionally a no-op.

**Alternative considered**: Check at job enqueue time in `image_upload_usecase`. Rejected — the count at enqueue time may differ from the count when the job runs (jobs can queue up), and the usecase is the authoritative enforcement point.

### `CountThisMonth` added to `CategorisationUsecase`, injected into `MeHandler`

`MeHandler` needs the monthly count for the `GET /me` response. The count logic belongs on `CategorisationUsecase` (it already owns `logRepo`). A minimal interface `CategorisationCountUsecase` with a single `CountThisMonth(ctx, userID) (int, error)` method is injected into `MeHandler` as a new dependency.

`categorisationUsecase` is already constructed before `meHandler` in `main.go` (line 262 vs 310), so no ordering changes are needed — just pass it in.

**Alternative considered**: Add the count to `userUsecase.GetByID` by injecting `logRepo` there. Rejected — it crosses the domain boundary; user usecase shouldn't know about categorisation logs.

### `CountByUserAndMonth` uses year + month int params

The repository method signature:
```
CountByUserAndMonth(ctx context.Context, userID string, year, month int) (int, error)
```
Using explicit year/month ints (rather than a `time.Time`) makes the method directly testable without time mocking.

`CountThisMonth` in the usecase calls it with `time.Now().UTC().Year()` and `int(time.Now().UTC().Month())`.

### Limit hardcoded as a constant

```go
const categorisationMonthlyLimit = 50
```

Defined in `categorisation_usecase.go`. Adjusting requires a redeploy, which is acceptable until there's a reason for per-user or configurable limits.

### Badge dismissal via localStorage

The red dot on the profile avatar is a frontend-only concern. Dismissal is stored in `localStorage` under the key `categorisation_limit_dismissed_<YYYY-MM>`. On mount, if the key exists for the current month, the badge is hidden. On trigger click (opening the dropdown), the key is written. Keys from prior months are naturally ignored — no cleanup needed.

**Alternative considered**: A server-side `dismissed_at` field on the user. Rejected — the badge is ephemeral UI state, not a meaningful preference to persist to the DB.

## Risks / Trade-offs

- **Count query on every GET /me**: A `SELECT COUNT(*)` against `ai_categorisation_logs` on each profile fetch. The table is small and scoped to `user_id + created_at`. Acceptable at current scale; add an index on `(user_id, created_at)` if it becomes a concern.

- **Silent skip is invisible without the counter**: If a user is at the limit and uploads an image, categorisation silently doesn't run. The "X / 50 this month" counter in settings is the sole signal. Users who don't check settings may be confused. Acceptable given the app's low-noise philosophy.

- **Monthly boundary is UTC-based**: The count window is the current UTC calendar month. Users in other timezones may see the count reset at an unexpected local time. Acceptable for now.

## Migration Plan

No database migrations. Feature is additive — existing users with `ai_categorisation_enabled = false` (all of them) see a new toggle in settings, defaulting to off. No backfill needed.

Deploy order: backend first (new PATCH field + GET count field are additive and backwards-compatible), then frontend.
