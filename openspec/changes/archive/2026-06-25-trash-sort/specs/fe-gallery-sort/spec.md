## MODIFIED Requirements

### Requirement: Direction toggle is shown only for orderable (non-Manual) sort fields

The sort panel SHALL display a direction toggle (`↑`/`↓`, with field-specific labels — e.g. "Oldest first"/"Newest first" for `Date added`, "Oldest deleted first"/"Newest deleted first" for `Date deleted`, "A → Z"/"Z → A" for `Name`) whenever the selected sort field is `Date added`, `Date deleted`, or `Name`. The toggle SHALL be omitted entirely — not shown-but-disabled — when `Manual` is selected, since direction has no meaning for a manually-ordered list.

#### Scenario: Direction toggle visible for Date added or Name

- **WHEN** the user selects `Date added` or `Name` as the sort field
- **THEN** a direction toggle appears showing the field-appropriate label for the current direction
- **AND** clicking it flips the direction and updates the label and the active query

#### Scenario: Direction toggle visible for Date deleted in the Trash view

- **WHEN** the user selects `Date deleted` as the sort field while viewing Trash
- **THEN** a direction toggle appears showing "Oldest deleted first"/"Newest deleted first" for the current direction
- **AND** clicking it flips the direction and updates the label and the active query

#### Scenario: Direction toggle hidden for Manual

- **WHEN** the user selects `Manual` as the sort field
- **THEN** no direction toggle is shown in the panel

### Requirement: Sort field options differ by view type

The sort panel's radio list SHALL show different options depending on the active view:

- **Folder views**: `Manual`, `Date added`, `Name` — with `Manual` listed first
- **All / Unsorted views**: `Date added`, `Name` — `Manual` is not offered, since these views have no persisted manual order (`position` is folder-scoped)
- **Trash view**: `Date deleted`, `Name` — `Manual` is not offered; `Date added` is NOT offered here (unlike All/Unsorted) since a trash listing's relevant date is when each item was deleted, not when it was created

`file_size` and `dimensions` (shown in the design handoff) are NOT offered in any view, since the backend's sort allow-list (`image-list-sort`) does not support them.

#### Scenario: Folder view shows Manual as the first option

- **WHEN** the user opens the sort panel while viewing a folder
- **THEN** the radio list shows `Manual`, `Date added`, `Name`, in that order, with `Manual` selected by default

#### Scenario: Non-folder, non-trash views omit Manual

- **WHEN** the user opens the sort panel while viewing All or Unsorted
- **THEN** the radio list shows only `Date added` and `Name`

#### Scenario: Trash view shows Date deleted instead of Date added

- **WHEN** the user opens the sort panel while viewing Trash
- **THEN** the radio list shows only `Date deleted` and `Name` — `Date added` is not offered

### Requirement: Sort selection resets to the view's default when switching views

The system SHALL reset the active sort field and direction to the new view's default whenever the user navigates to a different view (a different folder, All, Unsorted, or Trash) — mirroring the existing search-term reset behavior (`fe-gallery-search`). Sort state is per-view and session-local; it is never carried over from one view to another nor persisted across sessions.

Per-view defaults:

| View | Default field | Default direction |
|---|---|---|
| Folder | `Manual` | n/a |
| All / Unsorted | `Date added` | `desc` (newest first) |
| Trash | `Date deleted` | `desc` (most recently deleted first) |

The All/Unsorted defaults reproduce that pre-change ordering exactly. Trash's default does NOT reproduce its original pre-this-change ordering (`deleted_at ASC`, oldest-deleted-first) — it intentionally uses `desc` to match the "newest first" convention every other date-based default in this app uses, while ordering by the now-corrected `deleted_at` column.

#### Scenario: Switching from a folder to All resets sort to Date added (newest first)

- **WHEN** the user has an explicit sort active in a folder (e.g. `Name`, ascending) and navigates to the "All" view
- **THEN** the sort control shows `Date added` / "Newest first" as selected
- **AND** the image list is requested with no explicit sort change needed beyond the view's own default (i.e. the All view's default request, matching pre-change behavior)

#### Scenario: Switching between folders resets sort to Manual

- **WHEN** the user has an explicit sort active in one folder and navigates to a different folder
- **THEN** the sort control shows `Manual` as selected for the new folder
- **AND** images are requested with no `sort`/`direction` parameters (the new folder's `position`-ordered list)

#### Scenario: Switching to Trash resets sort to Date deleted (newest first)

- **WHEN** the user navigates to the Trash view from any other view
- **THEN** the sort control shows `Date deleted` / "Newest first" as selected
- **AND** `GET /images/trash` is called with `sort=deleted_at&direction=desc`

### Requirement: Selected sort is threaded through to the image list query

`getImages` and `getAllImages` (`src/lib/images.ts`) SHALL accept optional `sort` and `direction` parameters and include them as query parameters on `GET /images` when present. `getFolderImages` (`src/lib/images.ts`) SHALL accept optional `sort` and `direction` parameters and include them as query parameters on `GET /images/in-folder/:id` when present (omitted entirely when `Manual` is selected — see above). `getTrashedImages` (`src/lib/images.ts`) SHALL accept optional `sort: 'deleted_at' | 'title'` and `direction` parameters and include them as query parameters on `GET /images/trash` when present — this is a narrower, distinct type from `getImages`/`getAllImages`'s `'created_at' | 'title'`, since Trash's selectable fields have diverged from theirs. `ImageGrid`'s query key SHALL include the active `sort`/`direction` so that changing the sort triggers a re-fetch, and `useInfiniteQuery`'s existing `placeholderData: keepPreviousData` SHALL keep the prior results visible during the transition (mirroring how `debouncedSearchTerm` changes are handled today).

#### Scenario: Selecting an explicit sort re-fetches with the new ordering

- **WHEN** the user selects `Name` / ascending while viewing the "All" view
- **THEN** `GET /images` is called with `sort=title&direction=asc`
- **AND** the displayed images update to the new ordering once the response arrives, with the previous results remaining visible during the fetch

#### Scenario: Selecting Manual re-fetches with no sort parameters

- **WHEN** the user selects `Manual` while viewing a folder that previously had an explicit sort active
- **THEN** `GET /images/in-folder/<id>` is called with no `sort`/`direction` parameters
- **AND** the displayed images update to `position`-ordered results

#### Scenario: Selecting an explicit sort in the Trash view re-fetches with the new ordering

- **WHEN** the user selects `Name` / ascending while viewing the Trash view
- **THEN** `GET /images/trash` is called with `sort=title&direction=asc`
- **AND** the displayed images update to the new ordering once the response arrives, with the previous results remaining visible during the fetch

#### Scenario: Trash view's default sort sends explicit deleted_at/desc parameters

- **WHEN** the Trash view is opened without the user having changed the sort control
- **THEN** `GET /images/trash` is called with `sort=deleted_at&direction=desc`
- **AND** the displayed images are ordered newest-deleted-first
