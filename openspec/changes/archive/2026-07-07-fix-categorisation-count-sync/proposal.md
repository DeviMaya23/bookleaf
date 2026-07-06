## Why

When an AI categorisation completes, the `categorisation_complete` SSE event invalidates image and folder caches but not the `['me']` cache — so the monthly usage count shown in Settings → Advanced stays stale until the user hard-refreshes the page.

## What Changes

- On `categorisation_complete` SSE, also invalidate the `['me']` React Query cache so the count updates in real time without a page refresh.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `sse-events`: The FE cache invalidation requirement on `categorisation_complete` must also include the `['me']` query, as the event causes `ai_categorisation_count_this_month` to increment.

## Impact

- `frontend/src/app-shell/useSSEEvents.ts` — one additional `invalidateQueries` call
- `openspec/specs/sse-events/spec.md` — updated requirement and scenario
