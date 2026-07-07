## Context

When a categorisation job completes, the backend broadcasts a `categorisation_complete` SSE event. The frontend handles this in `useSSEEvents.ts` by invalidating image and folder caches. The monthly usage count (`ai_categorisation_count_this_month`) lives on the `['me']` cache, which has `staleTime: Infinity` — it is never refetched unless explicitly invalidated. As a result, the count shown in Settings → Advanced stays frozen at its value from page load.

## Goals / Non-Goals

**Goals:**
- `ai_categorisation_count_this_month` updates in real time after each categorisation, without a page refresh.

**Non-Goals:**
- Changing how the count is computed or stored on the backend.
- Changing the `staleTime` on the `['me']` query globally.
- Polling or proactive refetch of `['me']` outside of this event.

## Decisions

### Invalidate `['me']` in the SSE handler (chosen)

Add `queryClient.invalidateQueries({ queryKey: ['me'] })` to the `categorisation_complete` branch in `useSSEEvents.ts`, alongside the existing invalidations.

**Why**: The SSE event is the precise signal that the count changed. `invalidateQueries` triggers a background refetch only if the query is currently observed (i.e. Settings modal is open or another component reads `me`), so it's cheap when the settings modal is closed. This is the same pattern already used for folders and images.

**Alternatives considered:**

- *Optimistic +1 on the cached value* — no extra network request, but diverges from the authoritative count on the server. If the user's cached count was already stale, or multiple categorisations fire close together, the displayed value would be wrong. The monthly limit makes accuracy meaningful.
- *Lower `staleTime` on `['me']`* — would cause unnecessary refetches on every query mount, not just when the count changes. Too broad.
- *Refetch `['me']` when settings modal opens* — only surfaces the stale count when the user happens to open settings; the count is still wrong the rest of the session.

## Risks / Trade-offs

- [Extra `/me` request per categorisation] → Low impact: the request is background, small, and only fires if a component currently subscribes to `['me']`. With `staleTime: Infinity`, it only re-fires once per invalidation, not on every render.

## Migration Plan

No migration needed. Frontend-only change; no API, schema, or backend modifications.
