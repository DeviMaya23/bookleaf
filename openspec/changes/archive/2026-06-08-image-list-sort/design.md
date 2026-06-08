## Context

`GET /images` (`internal/repository/image_repository.go:34-80`) forks on `folderID != nil`:

- **Folder branch**: returns the full folder contents, hardcoded `Order("image_folders.position ASC")`, ignores cursor/limit/name entirely (per `image-endpoints` spec).
- **Non-folder branch** (All/Unsorted/Trash): keyset-paginated, hardcoded `Order("images.created_at DESC, images.id DESC")` and `Where("(images.created_at, images.id) < (?, ?)", cursor.CreatedAt, cursor.ID)`.

`ImageCursor`/`cursorPayload` (`internal/usecase/image_pagination.go:12-45`) is a fixed shape: `{CreatedAt, DeletedAt *time.Time, ID uuid.UUID}`. `DeletedAt` is already an optional field populated only for trash cursors — the struct already tolerates "shape varies by context."

The cursor is opaque to the frontend — it's a base64url string the client decodes nothing from and simply resends. That opacity is what makes a backend-only phase possible: as long as omitted `sort`/`direction` reproduce today's exact ordering and cursor semantics, the existing FE needs zero changes.

## Goals / Non-Goals

**Goals:**
- Add `sort`/`direction` query params to `GET /images`, validated against an allow-list (`400` on invalid values).
- Make keyset pagination correct under `title` ordering as well as `created_at` ordering — comparison column(s), operator, and tiebreak direction all need to track the active sort.
- Apply the chosen sort uniformly to both repository branches: folder views can opt into `title`/`created_at` ordering as an alternative to `position`, and non-folder views can opt into `title` ordering as an alternative to `created_at`.
- Preserve today's behavior exactly when `sort`/`direction` are omitted, so the cursor stays opaque and the existing FE is unaffected.

**Non-Goals:**
- Frontend sort control (placement, persistence, interaction) — phase 2.
- Deciding whether/how "manual" order is exposed outside folder views — already settled as folder-view-only; not re-litigated here.
- Sort fields beyond `created_at`/`title` (e.g. `file_size`, dimensions) — the allow-list is structured so adding one later is additive, not a redesign.

## Decisions

### 1. Two flat query params (`sort`, `direction`), not a combined value

`sort=created_at|title` and `direction=asc|desc`, each independently optional and validated. This matches the existing flat query-param style (`folder_id`, `tag_id`, `unfiled`, `name`) rather than introducing a combined-value convention (`sort=title:asc`) that nothing else in the API uses.

### 2. Allow-list maps each sort field to its column(s) and default direction

A small lookup (e.g. `map[string]sortConfig{column string; defaultDirection string}`) drives both validation and query construction:

| `sort` value | column        | default direction |
|--------------|---------------|-------------------|
| `created_at` | `created_at`  | `desc` (newest first — matches today) |
| `title`      | `title`       | `asc` (A→Z) |

The handler validates `sort` against this map's keys and `direction` against `{asc, desc}`; anything else is `400`. When `sort` is provided but `direction` is not, the field's default direction applies — so `?sort=title` alone yields A→Z without the caller needing to spell out `asc`.

**Alternative considered**: a single global default direction (e.g. always default to `desc`). Rejected — `created_at desc` (newest first) and `title asc` (A→Z) are the conventional defaults for those fields respectively; forcing one global default would mean `?sort=title` alone produces the *less* expected ordering (Z→A).

### 3. Cursor gains an optional typed field per sortable column, no discriminator

Extend `ImageCursor`/`cursorPayload` with `Title *string`, alongside the existing `CreatedAt time.Time` and `DeletedAt *time.Time`. No "which field is this cursor for" discriminator is added to the payload, because the caller already must resend the same `sort`/`direction` on every subsequent page request — exactly the same implicit contract that already governs `folder_id`/`tag_id`/`name` (the cursor "has no opinion about which WHERE clauses produced the page"; the request's params are the source of truth). The repository simply reads `cursor.CreatedAt` or `cursor.Title` based on the *current* request's `sort` value.

**Alternative considered**: a generic/polymorphic payload (`SortField string; SortValue any`). Rejected — `any` forces type assertions and loses the compile-time guarantees the typed-field approach keeps; it would also be a new pattern in a codebase that already solved "cursor shape varies by context" with plain optional typed fields (`DeletedAt`).

### 4. One dispatch table drives both `ORDER BY` and the keyset `WHERE`, keyed on (field, direction)

The rule is mechanical and the same for every field:
- `ORDER BY <column> <DIR>, id <DIR>`
- `WHERE (<column>, id) <op> (?, ?)`, where `<op>` is `>` for `asc` and `<` for `desc`

This is exactly today's `created_at DESC` logic generalized — `id` always tiebreaks in the same direction as the primary column, and the comparison operator is the mechanical inverse of the sort direction (asc → "give me what comes after", desc → "give me what comes before"). One small function `(field, direction) → (column, orderClause, whereOp)` replaces the two hardcoded strings in `imageRepository.List`.

### 5. Explicit `sort` overrides the folder branch's `position ASC` default

Today the folder branch ignores `cursor`/`limit`/`name` and always orders by `position ASC`. This change makes it also honor an explicit `sort` (swapping the `Order()` clause to `title`/`created_at` — no cursor concerns since the branch is unpaginated). When `sort` is omitted, `position ASC` remains the default — drag-reorder and the manual-order contract (`image-position-reorder`) are untouched. This gives folder views a `title`/`created_at` alternative to manual order "for free," using the same dispatch table as the non-folder branch, rather than leaving folder views permanently locked to position.

**Alternative considered**: keep the folder branch hardcoded to `position ASC` regardless of `sort`. Rejected — it would mean the same query parameter means different things depending on which branch handles it (honored vs. silently ignored), which is a sharper edge than "folder views also support explicit title/created_at sort, defaulting to position."

### 6. `direction` has no effect when `sort` is absent

If `direction` is supplied without `sort`, it is accepted but ignored — there is no "default field" for it to modify in a way a user would expect, and rejecting it would mean an unrelated, harmless param triggers a `400`. This mirrors the existing "extra params composed loosely" treatment (the folder branch already silently ignores `cursor`/`limit`/`name`).

## Risks / Trade-offs

- **[Risk]** `title` is not unique, so `ORDER BY title ASC, id ASC` alone could produce unstable pagination across pages with duplicate titles. → **Mitigation**: the dispatch table always appends `id` as a tiebreaker in the same direction as the primary column — identical in shape to today's `created_at, id` tiebreak, just generalized.
- **[Risk]** Growing `ImageCursor` with another optional field could be read as the struct becoming an unstructured grab-bag. → **Mitigation**: it follows the exact precedent already set by `DeletedAt *time.Time` (optional, populated only in specific contexts); this is consistent with the current shape, not a new pattern.
- **[Trade-off]** Folder views now have two ways to be ordered (`position` vs. explicit `sort`), which is a slightly larger behavioral surface than "always position." → Accepted: it's opt-in, defaults to today's behavior, and reuses the same dispatch table — the alternative (locking folder views to position regardless of `sort`) creates a worse inconsistency (a query param that's honored in one branch and silently dropped in the other).
- **[Trade-off]** This phase ships a backend capability the frontend won't exercise yet. → Accepted by design: it's verified via Bruno requests / integration tests against the handler and repository directly, and the opacity of the cursor guarantees the existing FE is unaffected — there's no "half-wired" state to worry about.

## Migration Plan

No database migration — sorting uses existing `images.created_at`/`images.title` columns, and cursor shape changes are purely an application-layer encoding concern (cursors are short-lived, session-scoped, never persisted, so there's no "old cursor meets new decoder" compatibility issue).

Deployable as a single backend change: add the params, generalize the cursor, generalize the query dispatch, validate via Bruno + integration tests. Rollback is a plain revert — no persisted state changes shape, so there's nothing to migrate backward.

## Open Questions

None outstanding — the frontend access pattern and the manual-vs-other-sorts UI scoping are intentionally deferred to the phase-2 (frontend) change, where they'll be explored and specified against a backend contract that already exists.
