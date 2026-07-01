## Context

`GET /images` already supports `sort`/`direction` end-to-end via a shared dispatch helper, `ResolveSort(sortField, direction) SortDispatchResult{Column, OrderClause, WhereOperator, DefaultDirection}` (`usecase/image_pagination.go`). It's consumed by:

- `handler/image.go` — validates `sort` against `{created_at, title}` and `direction` against `{asc, desc}`, 400s on bad input, defaults `direction` via `ResolveSort(sortField, nil).DefaultDirection` when omitted.
- `usecase/image_usecase.go` (`ListImages`) — passes `Sort`/`Direction` through to the repository untouched; uses `ResolveSort` again only to decide whether the next-page cursor needs a `Title`.
- `repository/image_repository.go` (`List`) — calls `ResolveSort`, uses `dispatch.OrderClause` for `ORDER BY`, and branches the keyset `WHERE` on `dispatch.Column` between a `(title, id)` and `(created_at, id)` tuple comparison using `dispatch.WhereOperator`.

The FE sort UI (`GalleryToolbar`, `useGalleryControls`) already renders and manages sort state for Trash, offering `created_at`/`title` (no `manual`); `defaultSortForViewType` initializes Trash to `{ sortBy: 'created_at', sortDir: 'desc' }`, the same default as All/Unsorted.

**Revision history on this design**: this change's first pass wired Trash's `created_at` option straight to `ResolveSort`'s existing `created_at` column — reusing it unmodified, with no new case. On manual verification this was found to be semantically wrong: a trash listing's "Date" sort should order by *when each item was deleted* (`deleted_at`), not when it was originally created. `created_at` for a trashed image answers "how old is this image," which isn't what a user sorting trash by date is asking. This revision corrects that: `deleted_at` comes back as a real, explicit, allow-listed sort field — but only for `GET /images/trash`, never for `GET /images` (where `deleted_at` is always null and has no meaning).

## Goals / Non-Goals

**Goals:**
- Make Trash view sort selection actually affect the returned order and pagination, with the *correct* fields: `Date deleted` (`deleted_at`) and `Name` (`title`).
- Reuse `ResolveSort` as the single dispatch function for both endpoints — adding one new case to it is acceptable (unlike the first pass's stance), since `deleted_at` is now a real, explicit, user-facing value, not a dead branch.
- Keep `GET /images`'s behavior, defaults, and allow-list (`created_at`/`title`) completely unchanged — `deleted_at` is reachable only via `GET /images/trash`.
- Trash's default ordering becomes `deleted_at DESC` (newest-deleted-first), matching the "newest first" convention `created_at`/Date-added uses everywhere else in the app, while ordering by the semantically correct column.

**Non-Goals:**
- Offering `deleted_at` as a sort option on `GET /images` — it has no meaning there (always null on non-deleted rows).
- Removing `created_at` as a valid explicit value for `GET /images/trash` — it stays allow-listed for direct API/bruno use (a trashed image's creation date remains meaningful), even though the FE no longer offers it as a UI choice for Trash.
- Adding new sort fields (e.g. file size, dimensions) to either endpoint.
- Touching `ListAllTrashed` or `ListExpiredTrash` (unpaginated, no sort concept).

## Decisions

**1. `ResolveSort` gains a `"deleted_at"` case (default direction `desc`), alongside unchanged `"title"` and default `"created_at"` cases.**

```go
case "deleted_at":
    if dir == "" { dir = "desc" }  // newest-deleted-first by default
    if dir == "asc" {
        return SortDispatchResult{Column: "deleted_at", OrderClause: "deleted_at ASC, id ASC", WhereOperator: ">", DefaultDirection: "desc"}
    }
    return SortDispatchResult{Column: "deleted_at", OrderClause: "deleted_at DESC, id DESC", WhereOperator: "<", DefaultDirection: "desc"}
```
Unlike the first pass's rejected `deleted_at` case, this one is reachable explicitly (`GET /images/trash?sort=deleted_at`) and is the real default for Trash — not a dead branch only reachable via an unused nil-default path. `GET /images`'s behavior is unaffected because its handler's allow-list (`{created_at, title}`) never passes `"deleted_at"` to `ResolveSort`.

**2. `GET /images/trash`'s default is decided in the handler, not by `ResolveSort`'s own nil-field default.**

`ResolveSort(nil, ...)` resolves to `created_at` — that's `GET /images`'s default and stays unchanged. Trash needs a *different* default (`deleted_at`). Rather than teach `ResolveSort` "if no field given AND this is the trash endpoint, use deleted_at" (which would make it endpoint-aware, a new and unwanted coupling), the trash handler itself substitutes a concrete default field (`"deleted_at"`) before calling `ResolveSort`, whenever the `sort` query param is absent. This mirrors how `GET /images/in-folder` already decides its own default (position-based ordering) entirely within its own handler path, independent of `ResolveSort`. `ResolveSort` itself remains a pure (field, direction) → dispatch function with no knowledge of which endpoint is calling it.

This is deterministic per-request (absence of `sort` always resolves to the same default), so pagination remains consistent across pages exactly like `GET /images`'s existing nil-default behavior already is — the client doesn't need to echo back `sort=deleted_at` on page 2 for correctness, any more than it needs to echo `sort=created_at` today for `GET /images`.

**3. `GET /images/trash`'s allow-list is `{created_at, title, deleted_at}` — a superset of `GET /images`'s `{created_at, title}`.**

All three fields are semantically meaningful on a trashed image record (when it was created, its name, when it was deleted), so all three are allow-listed for direct API use, even though the FE's Trash sort dropdown only ever offers two (`deleted_at`, `title`) post-change. This is the same kind of asymmetry the first pass already accepted in the opposite direction (a allow-listed-but-FE-unreachable value) — just inverted: `created_at` is allow-listed but no longer FE-reachable for Trash, instead of `deleted_at` being allow-listed but FE-unreachable.

**4. `ListTrashed`'s keyset `WHERE` and next-cursor construction branch three ways on `dispatch.Column` (`title` / `deleted_at` / `created_at`), reinstating `ImageCursor.DeletedAt`.**

```go
switch dispatch.Column {
case "title":
    // (images.title, images.id) <op> (?, ?)
case "deleted_at":
    // (images.deleted_at, images.id) <op> (?, ?)
default: // created_at
    // (images.created_at, images.id) <op> (?, ?)
}
```
The usecase's next-cursor builder mirrors this: populates `Title`, `DeletedAt`, or relies on the always-present `CreatedAt` field depending on `dispatch.Column`. `ImageCursor.DeletedAt` — marked "unused" by this change's first pass — comes back into active use, consistent with its original doc comment ("non-nil only for trash cursors").

**5. Frontend: `deleted_at` becomes a real `SortBy` variant, not a relabeling of `created_at`.**

The wire value sent as `sort=` must actually be `deleted_at` for the BE to order by the right column — a label-only fix (keeping `sort=created_at` but displaying "Date deleted") would be incorrect, since it wouldn't change what's actually being ordered by. `SortBy` (`useGalleryControls.ts`) widens to `'manual' | 'created_at' | 'title' | 'deleted_at'`; Trash's `sortFieldOptions` becomes `['deleted_at', 'title']`; its default becomes `deleted_at`/`desc`. `getTrashedImages`'s sort parameter type changes to `'deleted_at' | 'title'`, diverging from `getImages`/`getAllImages`'s `'created_at' | 'title'` — they're no longer the same type, since Trash's selectable fields have diverged from theirs.

## Risks / Trade-offs

- **[Risk]** This is the second behavior change to Trash's default ordering in one change (first pass: `deleted_at ASC` → `created_at DESC`; this revision: → `deleted_at DESC`). → **Accepted**: this is a within-change correction caught before merge/deploy, not a second production behavior change — only one default ordering change reaches users.
- **[Risk]** `ResolveSort` now has a case (`deleted_at`) that's meaningless for one of its two callers (`List`/`GET /images`). → **Mitigation**: `GET /images`'s handler allow-list physically prevents `"deleted_at"` from ever reaching `ResolveSort` via that path; this is enforced by existing validation, not convention.
- **[Trade-off]** `GET /images/trash`'s allow-list (`created_at, title, deleted_at`) is asymmetric with the FE's actual UI options (`deleted_at, title`) — `created_at` is API-reachable but not FE-reachable for Trash. → **Accepted**: `created_at` remains a meaningful field on a trashed image; excluding it from the allow-list would save nothing (no caller needs the protection) while removing a legitimately useful API capability.

## Migration Plan

No data migration. Deploy as a single change; no feature flag. Same additive-optional-params safety as before: BE-then-FE or FE-then-BE deploy ordering is safe. As with the first pass, the BE default-ordering change for Trash is immediately user-visible independent of FE wiring. Rollback is a straight revert.
