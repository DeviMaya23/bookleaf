## Why

`GET /images` has no user-controllable ordering — non-folder views are fixed to `created_at DESC` and folder views are fixed to `image_folders.position ASC`, so users can't sort their gallery alphabetically or by date. This is phase 1 (backend-only) of adding sort support: it builds the `sort`/`direction` contract and the cursor changes needed to paginate correctly under different orderings, without touching the frontend yet — FE sort UI and the manual-vs-other-sorts scoping decisions land in a follow-up phase once this contract exists.

## What Changes

- Add optional `sort` and `direction` query parameters to `GET /images`, supporting `created_at` and `title` as sort fields and `asc`/`desc` as direction, validated against an allow-list (invalid values return `400`).
- Generalize `ImageCursor` and its encode/decode functions to be aware of which field the keyset is built on, so cursor-based pagination stays correct when sorting by a field other than `created_at`.
- Generalize the `ORDER BY` / keyset `WHERE` clause construction in `imageRepository.List` to dispatch on the chosen sort field and direction (column(s), comparison operator, and `id` tiebreaker all vary by field/direction).
- When `sort`/`direction` are omitted, behavior is unchanged from today: non-folder views default to `created_at DESC`, folder views default to `image_folders.position ASC`. The cursor remains an opaque base64 string to the frontend, so the existing FE requires no changes for this phase.

**Out of scope for this phase:** frontend sort UI, the manual-vs-other-sorts scoping decision (folder views keep `Manual` as an option; other views won't), and additional sort fields beyond `created_at`/`title` (e.g. file size or dimensions).

## Capabilities

### New Capabilities
(none — this extends existing list/pagination capabilities)

### Modified Capabilities
- `image-endpoints`: `GET /images` gains optional `sort` and `direction` query parameters (allow-listed values, `400` on invalid input); the `ImageRepository.List` signature gains sort field/direction parameters that affect ordering on both the folder-view and cursor-paginated branches.
- `image-list-pagination`: `ImageCursor` and cursor encode/decode generalize from a fixed `(created_at, id)` shape to one that records which field the keyset is built on, and the keyset `WHERE` clause comparison (column(s), operator, direction) becomes sort-aware instead of hardcoded to `created_at DESC, id DESC`.

## Impact

- Backend: `internal/handler/image.go` (parse/validate `sort`/`direction` query params), `internal/usecase/image_pagination.go` (`ListImagesParams`, `ImageCursor`, cursor encode/decode generalization), `internal/usecase/image_usecase.go` + `internal/usecase/image_repository.go` (interface signature), `internal/repository/image_repository.go` (sort-aware `ORDER BY`/keyset `WHERE` dispatch in `List`).
- No database schema changes — sorting uses existing `images.created_at`/`images.title` columns.
- No frontend changes — the cursor stays opaque to the client and per-branch defaults preserve current behavior when `sort`/`direction` are omitted.
