## Context

Two lists in the gallery need name filtering, and they sit on opposite ends of a "is the full list already in memory?" divide:

- **Folders** (`GET /folders`): unpaginated — `folderUsecase.List` returns the user's entire folder tree in one call (`folder_repository.go:32-42`, `ORDER BY created_at ASC`, no `LIMIT`/cursor). The FE already holds the complete tree.
- **Images** (`GET /images`): the repository's `List` (`image_repository.go:34-80`) forks into two structurally different query paths:
  - **Folder-view branch** (`folderID != nil`): returns *all* images in that folder, ordered by `image_folders.position ASC`, ignoring cursor/limit entirely. The FE holds the complete per-folder list.
  - **Cursor-paginated branch** (`folderID == nil`, covers All/Unsorted/Trash): keyset pagination on `(created_at DESC, id DESC)`, `LIMIT n+1`, composes with `unfiled`/`tagID` WHERE clauses and a cursor WHERE clause. The FE holds only the fetched pages, fetched incrementally via `useInfiniteQuery`.

Where the full list is already client-side, a server round trip for filtering would be strictly worse (network latency, debounce complexity, and a "flash" as the paginated query resets) than filtering the in-memory array directly — which is exactly the pattern `FolderInput`/`TagInput` already use for their suggestion lists (`frontend/src/components/FolderInput.tsx:60-64`, instant `onChange` filtering, no debounce).

Where the list is paginated, filtering must happen server-side — the FE simply doesn't have the full data set to filter over.

## Goals / Non-Goals

**Goals:**
- Let users filter the folder sidebar by name, instantly, client-side.
- Let users search images by title within the view they currently have open (folder / All / Unsorted / Trash), matching Eagle's "search scoped to current view, clears on view switch" behavior.
- Keep the backend change minimal: one optional query parameter, one WHERE clause, only on the query path that actually needs it.

**Non-Goals:**
- Searching by tag (placeholder text says "by name or tag" but this change is title-only; tag search is a follow-up).
- Persisting or restoring search terms across view switches or page reloads.
- Fuzzy matching, ranking, or highlighting matched substrings — plain case-insensitive substring match (`ILIKE '%term%'` server-side, `toLowerCase().includes()` client-side) is sufficient.
- Changing how folders are paginated (they remain unpaginated; out of scope to introduce pagination here).

## Decisions

### 1. Folder filter and folder-view image filter are entirely client-side

Both already have their complete list in memory (`GET /folders` returns the whole tree; the folder-view branch of `GET /images` returns the whole folder). Filtering them is a pure array `.filter()` over already-fetched data — no debounce, no loading state, no backend change, and no risk of the "flash" that a server round trip would introduce. This mirrors the existing `FolderInput`/`TagInput` instant-filter pattern, so it introduces no new client-side pattern.

**Alternative considered**: route folder-view image search through the same backend `name` parameter as the paginated views, for symmetry. Rejected — it would add an unnecessary network round trip and debounce delay to a case where instant local filtering is strictly better, and it would require touching the folder-view branch of the repository query (which currently bypasses WHERE-clause composition entirely) for no UX benefit.

### 2. Paginated views (All/Unsorted/Trash) get a new `name` query parameter on `GET /images`

Added to `ListImagesParams` (`image_pagination.go:18-24`) alongside the existing `FolderID`/`Unfiled`/`TagID`/`Cursor`/`Limit`, threaded through the handler → usecase → repository, and applied as `WHERE images.title ILIKE ?` only in the cursor-paginated branch of `imageRepository.List` (`image_repository.go:54-79`). It composes with the existing `unfiled`/`tagID` WHERE clauses and the cursor WHERE clause exactly the way `tagID` already does — the cursor only encodes position `(created_at, id)` and has no opinion about which WHERE clauses produced the page, so adding another WHERE clause is orthogonal to pagination correctness. The client must resend the same `name` value on every subsequent page request, same as it already must for `folder_id`/`tag_id`/`unfiled`.

**Alternative considered**: prefix match (`ILIKE 'term%'`) instead of substring match, which could use a btree index. Rejected for now — substring match is the more useful "search as you type" behavior and matches user expectation (e.g. typing "tia" should find "heartopia"); at gallery scale, a sequential scan on `ILIKE '%term%'` is acceptable. If performance becomes a concern later, `pg_trgm` + a GIN index is the natural upgrade path without changing the API contract.

### 3. Search input state lives in `AppLayout`, passed down to `ImageGrid` as props

The spec requires the search input to render in `AppLayout`'s toolbar row (beside the "Image" upload button), not inside `ImageGrid`. That visual placement makes it impossible to also keep the term's state local to `ImageGrid` — the `<input>` and its `onChange` handler must live where they're rendered. So `searchTerm` and its debounced value (`useDebouncedValue`) are state in `AppLayout`, cleared via a `useEffect` keyed on `viewKey` (mirroring the existing pattern that resets `viewerImage`/`selectedImage`/`autoFocusTitle` on view change), and passed down to `ImageGrid` as `searchTerm`/`debouncedSearchTerm` props. `ImageGrid` already receives `view: AppView` as a prop and branches its query/fetch logic on it (`queryKeyFor`/`fetcherFor`), so it consumes the two extra string props the same way.

**Originally considered**: keep the term as local state inside `ImageGrid` to avoid prop drilling. Rejected once the toolbar placement requirement was confirmed — the input can't be rendered by `AppLayout` while its state lives in a different component, and `AppLayout` already lifts and passes down comparable per-view UI state (`view`, `selectedImage`, `viewerImage`), so two more string props follow an established pattern rather than introducing a new one.

### 4. Debounce the server-bound search, not the client-side filters

The client-side filters (folder sidebar, folder-view images) operate on in-memory arrays and re-render instantly — debouncing them would only add perceived latency for no benefit. The server-bound search (All/Unsorted/Trash) needs debouncing to avoid firing a request on every keystroke; ~300ms after the last keystroke is the conventional choice. This is a new pattern for the codebase (no existing debounce utility — `FolderInput`/`TagInput` filter in-memory data with no debounce), so it'll be implemented as a small local hook (e.g. `useDebouncedValue`) rather than pulling in a new dependency, per the "propose new dependencies before adding them" rule — a `useEffect` + `setTimeout` debounce is a handful of lines and doesn't warrant a library.

### 5. Use `placeholderData: keepPreviousData` to avoid grid flicker during debounced search

Each time the debounced `name` changes, it becomes part of the React Query key (alongside `view`), which triggers a fresh `useInfiniteQuery` fetch from page one. Without intervention, the grid would empty out and show a spinner between the old results and the new ones — a worse "flash" than the one the client-side filtering avoids. TanStack Query v5's `placeholderData: keepPreviousData` keeps the previous page's results rendered while the new query resolves, so the grid updates in place.

## Risks / Trade-offs

- **[Risk]** `ILIKE '%term%'` cannot use a standard index and forces a sequential scan on the `images` table. → **Mitigation**: acceptable at current/expected gallery scale (personal moodboarding tool, not a multi-tenant SaaS with millions of rows per user); if it becomes a problem, `pg_trgm` + GIN index is a drop-in upgrade that doesn't change the API.
- **[Risk]** Forgetting that `imageRepository.List` has two independent query-builder branches could lead to the `name` filter being added only to the paginated branch, silently no-op'ing for folder views (though folder views are filtered client-side in this design, so this isn't actually a gap — but it's worth calling out explicitly so a future change to "also filter folder-view images server-side" doesn't miss the second branch). → **Mitigation**: this design doc and the spec scenarios make explicit which branch the filter applies to.
- **[Trade-off]** Client-side filtering for folder views means the search box behaves slightly differently depending on which view is active (instant vs. debounced-with-network-latency) — a user moving from a folder view to "All" might notice the search becomes less instantaneous. → Accepted: this is inherent to the pagination divide, matches Eagle's behavior (which has the same characteristic), and the alternative (forcing folder views through the network too) would make the common case worse to paper over an edge case.

## Open Questions

None outstanding — the proposal and design were settled through exploration before drafting.
