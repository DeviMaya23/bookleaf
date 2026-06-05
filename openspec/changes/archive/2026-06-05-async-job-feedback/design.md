## Context

After the async job queue implementation, thumbnail generation and vision labelling run as River background jobs after `CompleteUpload` returns. This broke two user-facing flows:

1. **Thumbnail**: `MasonryCardContent` already handles `thumbnail_url: null` gracefully (renders an `ImageIcon` placeholder), but the card never updates to the real thumbnail without a page reload — the React Query cache is stale.
2. **Folder suggestion**: Vision results are stored in `Image.ai_labels` but never surfaced via the API. The `suggested_folder_name` field in `CompleteUploadResult` (FE type) is a dead field; the suggestion dialog in `UploadModal` never fires.

The backend's `ProcessVisionLabelling` also has an ambiguous signal problem: it writes nothing to `ai_labels` when vision is disabled or returns zero labels, making `null` mean both "not run yet" and "done, nothing to report."

## Goals / Non-Goals

**Goals:**
- Thumbnail card updates live (without reload) when the background job completes
- Vision suggestion reaches the user as a toast immediately after the job finishes, on a best-effort basis
- `null` `ai_labels` reliably signals "job not yet complete"
- Dead suggestion dialog code removed from `UploadModal`

**Non-Goals:**
- Real-time push (SSE, WebSockets) — polling a single known image ID is sufficient and avoids persistent connection complexity
- Surfacing `ai_labels` data beyond `suggested_folder_name` (raw labels are stored for future use)
- Modifying batch upload modal behaviour

## Decisions

### 1. Poll `GET /images/:id` as the status source

The single-image detail endpoint already returns `thumbnail_url`. Adding `suggested_folder_name` to it gives us one endpoint that answers both questions without a new route.

**Alternative considered:** Dedicated `/images/:id/async-status` endpoint. Rejected — adds a route and handler for a thin wrapper that duplicates data already in `GetImage`. The detail endpoint already requires auth and ownership checks; reusing it is simpler.

### 2. Two independent effects in one hook — setInterval for thumbnail, setTimeout for vision

A single `usePostUploadFeedback(imageId, visionEnabled)` hook runs after `CompleteUpload` returns. It uses two separate React effects:

- **Thumbnail** (`useEffect` with `setInterval`, 2s tick, 30s cap): polls `GET /images/:id` every 2s; on `thumbnail_url` non-null, calls `queryClient.invalidateQueries({ queryKey: ['images'] })` to trigger a refetch and clears the interval. Stops after 15 attempts (~30s).
- **Vision** (`useEffect` with `setTimeout`, 2.5s delay, single call): only runs when `visionEnabled` is true. After 2.5s, calls `GET /images/:id` once. Shows a suggestion toast if `suggested_folder_name` is present, otherwise shows a "Couldn't get folder suggestion" error toast.

The two effects are independent; each self-cleans via the effect's return function.

**Why `invalidateQueries` instead of `setQueriesData`:** The original plan was to patch the image in-place via `setQueriesData`. This introduced a closure/cleanup race: the `setInterval` callback is async, so `clearInterval` during cleanup does not cancel an already-in-flight `getImage` call. When that call resolved after cleanup, `setQueriesData` would run against stale closure state and could patch the wrong query key (e.g. if the user navigated to a different folder view). `invalidateQueries` avoids the key-matching problem entirely — it just marks the cache stale and lets React Query refetch cleanly.

**Why two effects instead of one shared interval:** The vision check fires once after a fixed delay (not on every tick), so a `setTimeout` is the natural primitive. Sharing an interval would require tracking elapsed time manually to know when to fire the one-shot vision check, adding complexity for no benefit. Two independent effects with separate lifecycles are simpler and easier to reason about.

**Alternative considered:** One shared `setInterval` with `isDone` flags. Rejected — requires manual elapsed-time tracking for the vision one-shot, and introduces shared mutable state between the two concerns.

### 3. `ai_labels = []` as the "job ran, nothing to report" sentinel

`ProcessVisionLabelling` currently early-returns without writing when `vision_enabled` is false or Vision returns zero labels. This makes `null` ambiguous.

Change: always call `UpdateAILabels` at the end of the method — with the labels array (possibly empty `[]`). `null` then exclusively means the River job has not yet reached `ProcessVisionLabelling`.

This is a pure backend change; no migration needed (existing `null` rows just mean the job hasn't run, which remains accurate for images uploaded before this change).

**Alternative considered:** A separate `vision_labelled_at` timestamp column. Rejected — schema change is heavier than a simple empty-array write.

### 4. `suggested_folder_name` derived inline in `GetImage` usecase

When `GetImage` is called, if `Image.AILabels` is non-null and non-empty, pick the first label (labels are stored ordered by Vision score descending) and return its `Description` as `suggested_folder_name`. No new DB column or join needed.

The FE only reads this field immediately post-upload via the polling hook; it's ignored in all other `GetImage` callers.

### 5. Suggestion delivered as a Sonner toast with action buttons

Sonner (already in use) supports `action` and `cancel` on toasts. The Accept action calls the existing `acceptSuggestion` API function directly from the toast callback. No new endpoint or modal state needed.

The existing inline suggestion view in `UploadModal` (the `SuggestionState` type, `suggestion` state, `acceptMutation`, and the suggestion JSX branch) is removed. The modal now always closes immediately on upload success.

## Risks / Trade-offs

- **Vision check miss**: If the River job completes after the 2.5s window (e.g. under queue load), the user gets an error toast even though a suggestion will eventually exist in the DB. Acceptable — the suggestion is best-effort, and the stored `ai_labels` will be available for future features (e.g. right-click context menu).
- **Extra refetch on thumbnail resolve**: `invalidateQueries` triggers a full list refetch rather than a targeted cache patch. At single-user scale this is negligible; a future optimisation could use `setQueriesData` once the key-matching logic is reliable.
- **Multiple simultaneous uploads**: Each upload creates its own `usePostUploadFeedback` instance polling independently. At typical usage volume (single user, personal app) this is negligible.

## Migration Plan

No schema changes. No data migrations. Deploy order: backend first (adds `suggested_folder_name` to response, changes `ai_labels` write behaviour), then frontend. The frontend polling hook gracefully handles the absence of `suggested_folder_name` (null = no suggestion) if backend deploys first while frontend is still old.
