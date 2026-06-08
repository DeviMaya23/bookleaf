## Context

`GET /images` (`internal/handler/image.go:104`) currently serves two structurally different queries behind one signature:

- **Folder view** (`folder_id` present): full fetch, no pagination, ordered by `image_folders.position`, returns per-image `FolderPosition` — a traversal of the `image_folders` relationship, used by the drag-and-drop board view.
- **Gallery view** (`folder_id` absent): keyset-paginated, sortable, filterable by `name`/`tag_id`/`unfiled` — a search/filter query over `Image`.

`ImageRepository.List` (`internal/repository/image_repository.go:34`) branches on `folderID != nil` to pick between these two query shapes (lines 38-58 vs. 61-95), and `ImageUsecase.ListImages` (`internal/usecase/image_usecase.go:74`) mirrors that branch.

This design extracts the folder-view query into its own endpoint, then rewrites the gallery query's filters to support multi-value "match any" semantics for `folder_ids`, `tag_ids`, and a new `mime_types` filter — composing cleanly with existing pagination, sort, `name`, and `unfiled`.

## Goals / Non-Goals

**Goals:**
- Give the "folder contents in custom order" query its own endpoint (`GET /images/in-folder/:id`), in the image domain, so it stops occupying a branch of `GET /images`.
- Let `GET /images` accept `folder_ids`, `tag_ids`, `mime_types` as independent multi-value match-any filters that compose with each other and with `name`/`unfiled`/sort/pagination.
- Ensure multi-value filters cannot duplicate rows, since duplication would corrupt the keyset cursor's `limit + 1` probe (`image_usecase.go:121`) and the `(column, id)` comparison it relies on for "is there a next page."

**Non-Goals:**
- FE wiring (separate phase — the FE folder-view fetch path will need to be repointed to the new endpoint, but that work and its sequencing relative to this change is tracked elsewhere).
- Match-ALL semantics for any filter (e.g. "image has tag A AND tag B"). Only match-any is in scope, per the proposal.
- Any change to `image_folders`/`image_tags` schema, indexes, or write paths (`SetImageFolder`, `SyncImageFolders`, `ReplaceImageTags`, etc.) — this is a read-path change only.
- Changes to the sort feature or `ResolveSort` dispatch — filters and sort are orthogonal; this design must compose with sort, not modify it.

## Decisions

### 1. New route: `GET /images/in-folder/:id`

Lives on `imageHandler`, registered alongside the other `/images/*` routes in `cmd/server/main.go`. Keeping it under `/images/*` (rather than `/folders/:id/images`) keeps the route prefix aligned with the handler that actually owns it — consistent with the existing local pattern (`/images/:id/move-folder`, `/images/:id/position`), and avoids the internal mismatch of an `imageHandler` route living under a `/folders/*` prefix. The three-segment path (`/images/in-folder/:id`) does not collide with the existing two-segment `GET /images/:id`.

This endpoint:
- Accepts `:id` (folder UUID), validated and scoped to the authenticated user (folder ownership check via existing folder lookup, mirroring how `folder_id` is currently validated in the move-folder/position flows).
- Returns **all** non-deleted images in that folder, **ordered by `image_folders.position ASC`** (or by an explicit `sort`/`direction` if provided — preserving today's "explicit sort overrides position" behavior documented in `image-endpoints`'s "List honors an explicit sort field in folder views" scenario).
- Returns each image with its `FolderPosition` for that folder.
- Does **not** paginate — no `cursor`/`limit` params, matching today's folder-view contract (folders are bounded collections; this mirrors the existing implicit assumption rather than introducing a new one).
- Does **not** accept `name`/`tag_ids`/`mime_types`/`unfiled` — it answers one question only ("what's in this folder, in order"), keeping its contract narrow and avoiding re-introducing a mode-switch inside the new endpoint.

New usecase method (e.g. `ListFolderImages(ctx, userID, folderID, sort, direction)`) and repository method (e.g. `ListByFolder(ctx, userID, folderID, sortField, direction)`) carry over the extracted logic from `image_repository.go:38-58` essentially unchanged — this is a relocation, not a rewrite, of that branch.

### 2. `folder_id`/`tag_id` (singular) are removed, not deprecated-and-kept

`GET /images` drops `folder_id`/`tag_id` outright in favor of `folder_ids`/`tag_ids`. Per project convention (no compatibility shims for internally-controlled contracts), and because the FE is the only consumer, there's no value in carrying both shapes — it would only let the mismatch between BE and FE silently linger longer. The breaking change is called out in the proposal; the FE must be updated in the same change window.

### 3. Multi-value query param encoding: comma-separated (CSV)

`folder_ids=<uuid>,<uuid>`, `tag_ids=<uuid>,<uuid>`, `mime_types=image/jpeg,image/png`.

Considered repeated params (`?tag_ids=a&tag_ids=b`, via `c.QueryParams()["tag_ids"]`) — more conventional for arrays in some REST styles, but this codebase has no existing precedent for it (grep across `internal/handler/` turns up none), while every existing multi-part value here (e.g. the `Authorization` header split in `middleware/auth.go:146`) is parsed via simple delimiter splitting. CSV also matches how the FE would naturally serialize a multi-select filter into a single query string value with `URLSearchParams`. None of the three value domains (UUIDs, MIME types like `image/jpeg`) can contain a comma, so splitting is unambiguous. Each split value is validated individually (UUID parse for `folder_ids`/`tag_ids`; non-empty string for `mime_types`) — a malformed value in the list returns `400`, mirroring the existing single-value validation style (e.g. `image.go:115-118`).

### 4. Multi-value filter SQL: `EXISTS` subqueries, not `JOIN ... IN (...)`

A naive `JOIN image_tags ON image_tags.image_id = images.id AND image_tags.tag_id IN (...)` returns one row per matching join match — an image with 3 of 5 selected tags appears 3 times. That breaks the keyset pagination contract: the `len(rawImages) > limit` probe (`image_usecase.go:121`) and the `(created_at, id)`/`(title, id)` cursor comparison both assume one row per image.

Instead, each multi-value relational filter is expressed as a correlated `EXISTS`:

```sql
WHERE EXISTS (
  SELECT 1 FROM image_tags
  WHERE image_tags.image_id = images.id AND image_tags.tag_id IN (?, ?, ...)
)
```

and analogously for `folder_ids` against `image_folders`. This guarantees at most one row per image regardless of how many values match, composes via `AND` with every other filter (each `EXISTS`/`IN`/`ILIKE` clause is independent), and needs no `DISTINCT`/`GROUP BY` (which would otherwise complicate the `Preload("Tags")`/`Preload("ImageFolders")` + `Order` + keyset combination). Existing indexes — `idx_image_folders_image_id` and the equivalent on `image_tags` — support the correlated subquery's lookup by `image_id`.

`mime_types` needs no `EXISTS` — `MIMEType` is a plain column on `images`, so it's a direct `WHERE images.mime_type IN (?, ?, ...)`.

### 5. Cross-filter composition: AND across filters, OR (match-any) within each

`folder_ids=[A,B]&tag_ids=[X,Y]` reads as "images in folder A or B, AND tagged with X or Y" — each filter independently narrows the result set via `AND`, while the multiple values *within* a filter are unioned via `IN`/`EXISTS ... IN`. This is the standard faceted-search interpretation and requires no special-casing: it falls out naturally from each filter being its own independent `WHERE` clause.

### 6. `unfiled` and `folder_ids` together

Both remain independent boolean/list filters, ANDed in as today. `unfiled=true&folder_ids=[A]` is a contradiction (an unfiled image cannot be "in folder A"), but no special validation is added for it — it simply yields an empty result, the same way an impossible `name`+`tag_ids` combination would. Adding cross-filter contradiction validation would be new logic this feature doesn't need; trusting the SQL to "just answer the question literally asked" keeps the filter layer uniform.

## Risks / Trade-offs

- **[Risk]** `GET /images` contract change is breaking; if the BE ships before the FE folder-view fetch is repointed to `/images/in-folder/:id`, the existing FE folder view will silently receive paginated/unordered gallery results instead of full position-ordered folder contents (degraded, not crashing — arguably worse, since it could go unnoticed).
  → **Mitigation**: land the FE repointing in the same release window as this BE change; call this out explicitly in the rollout/PR description so it isn't deployed independently.
- **[Risk]** `EXISTS` subqueries may be planned differently than the `JOIN`s they replace, with unknown effect on query performance at scale.
  → **Mitigation**: the correlated subquery filters on an indexed `image_id` column in both `image_tags` and `image_folders`; this is a standard, well-optimized pattern in Postgres. Worth an `EXPLAIN ANALYZE` spot-check during implementation against a representative dataset, but no structural risk is anticipated.
- **[Trade-off]** Removing `folder_id`/`tag_id` outright (rather than accepting both shapes temporarily) means the BE and FE must deploy together — there's no graceful transition window. This is an intentional trade against carrying dead/duplicated parsing logic that would need its own removal later (see Decision 2).

## Migration Plan

- **No database migration** — `image_folders`, `image_tags`, and `images.mime_type` already exist with the needed columns and indexes; this is a read-path-only change.
- **Deploy**: the new `GET /images/in-folder/:id` endpoint is purely additive and can land independently. The `GET /images` rewrite is breaking and should be coordinated with the FE change that repoints the folder-view fetch — ideally the same PR/release, or BE-then-immediately-FE within the same deploy window.
- **Rollback**: a straight code revert; no schema state to unwind.

## Open Questions

- Bruno collection structure for the new endpoint — new file under `bruno/images/` (e.g. `list-folder-images.bru`), plus an update to the existing `bruno/images/list-images.bru` to reflect the new `folder_ids`/`tag_ids`/`mime_types` params and the removal of `folder_id`/`tag_id`.
- Exact naming for the new usecase/repository methods (`ListFolderImages`/`ListByFolder` used above are placeholders) — to be finalized during task breakdown in line with existing naming patterns on `ImageUsecase`/`ImageRepository`.
