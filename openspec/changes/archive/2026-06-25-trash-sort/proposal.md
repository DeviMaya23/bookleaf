## Why

The gallery's sort control (`fe-gallery-sort`) already renders and manages sort state for the Trash view, and its own spec already lists `Date added`/`Name` as Trash's sort options — but the request-threading requirement that makes sort actually take effect was only ever written for `getImages`/`getAllImages`/`getFolderImages`. `getTrashedImages` never gained `sort`/`direction` parameters, and the backend's `ListTrashed` path never gained the `ResolveSort`-based dispatch that `List` already uses — it still issued a hardcoded `ORDER BY deleted_at ASC, id ASC`. The result: selecting a sort option in the Trash view was a no-op.

**Revision**: the first pass of this change wired Trash's "Date added" option to `created_at`, matching `List`'s default field. On manual verification, this was wrong: for a trash listing, "Date added" should mean "Date deleted" — the column a user actually cares about when looking at trashed items is when each item was deleted, not when it was originally created. This revision relabels and rewires that option to `deleted_at`, while keeping `Name`/`title` unchanged.

## What Changes

- `ResolveSort` (`backend/internal/usecase/image_pagination.go`) gains a `"deleted_at"` case (default direction `desc`, newest-deleted-first), alongside its existing `"title"` and default `"created_at"` cases. `GET /images`'s behavior is unaffected — it never requests `deleted_at`.
- `GET /images/trash` handler (`handler/trash.go`) allow-lists `sort` ∈ `{created_at, title, deleted_at}` (`created_at` kept for direct API/bruno use even though the FE no longer offers it for Trash — it remains a meaningful field on a trashed image). When `sort` is absent, the handler now defaults it explicitly to `"deleted_at"` (not left nil) — this is Trash's endpoint-specific default, decided in the handler exactly like `GET /images/in-folder`'s position-based default is decided in its own handler path; `ResolveSort` itself stays generic and endpoint-unaware.
- `TrashRepository.ListTrashed` (`repository/image_repository.go`) keyset `WHERE` branches three ways on `dispatch.Column` (`title` / `deleted_at` / `created_at`), reinstating use of `ImageCursor.DeletedAt` (previously added back to "unused" by this change's first pass).
- `TrashUsecase.ListTrashed`'s next-cursor construction (`usecase/trash_usecase.go`) branches three ways to populate `Title`/`DeletedAt`/`CreatedAt` on the cursor depending on `dispatch.Column`.
- Frontend: `SortBy` (`useGalleryControls.ts`) gains a `'deleted_at'` variant; Trash's `sortFieldOptions` becomes `['deleted_at', 'title']` (was `['created_at', 'title']`); Trash's default `sortBy` becomes `'deleted_at'`/`desc`. `GalleryToolbar.tsx`'s `SORT_FIELD_LABELS` gains `deleted_at: 'Date deleted'`; `DIR_LABELS` gains deleted_at-specific direction labels.
- `getTrashedImages`'s `sort` parameter type (`lib/images.ts`) changes from `'created_at' | 'title'` to `'deleted_at' | 'title'` — it no longer shares a sort-type alias with `getImages`/`getAllImages`, since Trash's selectable fields have diverged from theirs.
- **BREAKING (behavioral, supersedes this change's first pass)**: Trash's default ordering is `deleted_at DESC, id DESC` (newest-deleted-first) — not `created_at DESC` as this change's first pass shipped, and not the original pre-change `deleted_at ASC` either. There is no migration step; this is a direct behavior change taking effect on deploy.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `image-list-pagination`: `ListTrashed`'s sort order and pagination-cursor requirements use a three-way (`title`/`deleted_at`/`created_at`) `ResolveSort`-derived dispatch, with `deleted_at DESC` as the default.
- `image-endpoints`: `GET /images/trash`'s sort allow-list is `{created_at, title, deleted_at}` with a `deleted_at` default when `sort` is omitted — distinct from `GET /images`'s unchanged `{created_at, title}` allow-list and `created_at` default.
- `fe-gallery-sort`: Trash's sort field options, labels, and default change from `Date added` (`created_at`) to `Date deleted` (`deleted_at`); `Name`/`title` is unchanged.

## Impact

- **Backend**: `usecase/image_pagination.go` (`ResolveSort` — new case, but `title`/default `created_at` cases unchanged, so `GET /images` is unaffected), `repository/image_repository.go` (`ListTrashed` impl), `usecase/trash_usecase.go` (`ListTrashed` next-cursor logic), `handler/trash.go` (`ListTrashed` handler — allow-list and default).
- **Frontend**: `useGalleryControls.ts` (`SortBy`, `FIELD_DEFAULT_DIRECTION`, `defaultSortForViewType`, `sortFieldOptions`), `GalleryToolbar.tsx` (`SORT_FIELD_LABELS`, `DIR_LABELS`), `lib/images.ts` (`getTrashedImages`'s sort param type).
- **Tests**: BE repo/usecase/handler tests added for `deleted_at` sort and updated for the new default; existing `title`/`created_at` `ResolveSort`/`List` tests must keep passing unchanged. FE tests for `useGalleryControls`/`GalleryToolbar`/`useGalleryImages` updated for the new default and label.
- **Bruno**: `bruno/images/list-trash.bru`'s `~sort` example value should reflect a valid Trash sort field (`deleted_at` or `title`, not `created_at` as a UI-driven default, though `created_at` remains a valid explicit value to demonstrate too).
