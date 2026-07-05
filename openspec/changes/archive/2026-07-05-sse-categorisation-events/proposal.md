## Why

When the AI categorisation job completes, the frontend has no signal to refresh — leaving the image's assigned folder invisible until the user manually reloads or navigates away and back. SSE gives the frontend a lightweight push notification so it can refresh the right caches the moment categorisation finishes, regardless of whether the upload originated from the web app or the browser extension.

## What Changes

- **New** `GET /events` endpoint (auth-protected, SSE stream) — keeps a long-lived connection open per browser tab and pushes events as they occur
- **New** in-process `EventBroadcaster` — pub/sub component that fans events out to all open tabs for a given user; lives at `internal/sse/`
- **New** `useSSEEvents` FE hook — opens an `EventSource` connection when `ai_categorisation_enabled` is true, listens for `categorisation_complete`, and invalidates affected React Query caches
- **Modified** `GET /me` response — exposes `ai_categorisation_enabled` flag (field exists in DB, not yet returned by the endpoint)
- **Modified** `CategorisationWorker` — emits a `categorisation_complete` event via the broadcaster after a successful job, carrying only `image_id`

## Capabilities

### New Capabilities

- `sse-events`: In-process SSE broadcaster and `GET /events` streaming endpoint; FE hook that connects, parses events, and dispatches cache invalidations on `categorisation_complete`

### Modified Capabilities

- `me-endpoint`: `GET /me` response gains `ai_categorisation_enabled` boolean field; `PATCH /me` response mirrors the same field

## Impact

- **BE**: new `internal/sse/` package; `internal/handler/events.go`; `CategorisationWorker` gains a broadcaster dependency; `initApp` wires broadcaster to both worker and handler; `me.go` updated to include `ai_categorisation_enabled` in responses
- **FE**: `lib/me.ts` Me type gains `ai_categorisation_enabled`; new `app-shell/useSSEEvents.ts`; `AppLayout.tsx` mounts the hook
- **Extension**: no changes — categorisation events fire for any upload origin; the FE tab picks them up passively
- **Scope note**: broadcaster is in-process only (single-server assumption); no Redis or Postgres LISTEN/NOTIFY needed
