## Context

Bookleaf's AI categorisation runs as a River background job. The job completes seconds after upload — outside any FE request/response cycle — and directly assigns the image to a folder in the DB. The FE currently has no mechanism to learn when this happens, so the image's folder assignment is invisible until the user refreshes.

The app uses Kinde JWT Bearer tokens for auth. The native browser `EventSource` API cannot send custom headers, so standard SSE with Bearer auth is not directly possible. The `@microsoft/fetch-event-source` library works around this by implementing the SSE protocol over `fetch`, which supports arbitrary headers.

## Goals / Non-Goals

**Goals:**
- Push a `categorisation_complete` event to all open tabs of the affected user when a categorisation job succeeds
- FE invalidates relevant React Query caches on receipt, making the folder assignment visible without a manual refresh
- Work for uploads originating from any source (web app, browser extension)
- Keep the broadcaster in-process; no external infrastructure

**Non-Goals:**
- Generalising SSE to other job types beyond categorisation (infrastructure is generic enough to extend, but no other events are wired now)
- Persisting events for offline/reconnect replay
- Horizontal scaling (broadcaster is in-memory; multi-instance deployments would need Redis pub/sub or Postgres LISTEN/NOTIFY)
- Making `ai_categorisation_enabled` user-editable

## Decisions

### D1: In-process broadcaster over Postgres LISTEN/NOTIFY

**Decision**: Use an in-memory `EventBroadcaster` struct with `map[userID][]chan Event`.

**Rationale**: The app runs as a single server instance. Postgres LISTEN/NOTIFY or Redis pub/sub would add operational complexity (extra connection, new dependency) for no benefit at current scale. A comment in the broadcaster documents the single-server assumption so it's visible when that changes.

**Alternative considered**: Postgres LISTEN/NOTIFY — self-contained but adds a persistent listener connection and complicates shutdown logic. Deferred.

### D2: Emit from worker, not usecase

**Decision**: `CategorisationWorker` calls `broadcaster.Publish` after `usecase.CategoriseImage` returns `nil`.

**Rationale**: SSE notification is an infrastructure side-effect of the async job succeeding — not a business rule. The usecase stays focused on domain logic. `CategoriseImage` signature is unchanged (`error` only), which also means no usecase change ripples to tests.

**Alternative considered**: Emitting from inside the usecase — would couple the domain layer to an infrastructure concern and make the usecase harder to test in isolation.

### D3: `@microsoft/fetch-event-source` for FE SSE

**Decision**: Use `@microsoft/fetch-event-source` instead of native `EventSource`.

**Rationale**: Native `EventSource` cannot send custom headers, so Bearer token auth is impossible. `fetch-event-source` implements the same SSE protocol over `fetch`, allowing the standard `Authorization: Bearer <token>` header. No changes needed to the BE auth model.

**Alternative considered**: Token in query string — simple but exposes the JWT in server logs and browser history. Rejected on security grounds.

### D4: Gate SSE connection on `ai_categorisation_enabled`

**Decision**: FE opens the EventSource connection only when `me.ai_categorisation_enabled` is `true`.

**Rationale**: Users without the feature enabled will never receive a `categorisation_complete` event. Holding an idle open connection for them wastes a connection slot. Gating on the flag avoids the idle connection with minimal complexity — `ai_categorisation_enabled` is already fetched as part of `GET /me`.

### D5: Buffered channels in broadcaster

**Decision**: Each subscriber channel is created with a small buffer (`cap = 16`).

**Rationale**: Prevents a slow SSE write (network hiccup, blocked flush) from blocking the River worker goroutine that calls `Publish`. If the buffer fills (extremely unlikely for low-frequency categorisation events), the send is dropped rather than blocking.

### D6: Event payload carries only `image_id`

**Decision**: `{ "type": "categorisation_complete", "payload": { "image_id": "..." } }` — no `folder_id`.

**Rationale**: The BE should emit facts about what completed, not prescribe which FE caches to invalidate. The FE owns cache management. With only `image_id`, the FE does a slightly broader invalidation (prefix-bust `['folder']` instead of targeting a specific folder detail), which is acceptable given only one folder detail panel is open at a time.

## Risks / Trade-offs

**Single-server only** → If the app scales to multiple instances, events emitted on server A won't reach SSE connections held on server B. Mitigation: the broadcaster is isolated to `internal/sse/` with a documented constraint; migrating to Postgres LISTEN/NOTIFY later is a contained change.

**Connection lifecycle on token expiry** → If the Kinde JWT expires while the SSE connection is open, the next reconnect attempt will fail auth. `fetch-event-source` retries automatically; the retry will re-fetch a fresh token via `getToken()` before opening the new request. No special handling needed.

**Broadcaster not wired when Anthropic key is absent** → The `CategorisationWorker` is only registered when `cfg.AnthropicAPIKey != ""`. The broadcaster and SSE handler must still be created unconditionally so the `GET /events` endpoint exists for all users (it will just never receive categorisation events for users without the key configured). The broadcaster is cheap to create with no active clients.

## Migration Plan

Deploy is additive — no DB migrations, no breaking API changes. `GET /events` is a new endpoint. `GET /me` gains a new field (`ai_categorisation_enabled`) which FE can start reading after deploy. Rollback: revert the deploy; FE falls back to the stale-until-refresh behaviour that exists today.
