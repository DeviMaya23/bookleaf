# fe-gallery-sort

## Purpose

TBD — covers the gallery's sort control: the toolbar UI for choosing a sort field and direction, per-view field options and defaults, the Manual sort framing for folder views, the active-sort indicator, and threading the selected sort through to the image list query.

## Requirements

### Requirement: Sort control in gallery toolbar

The system SHALL display a sort icon button in the gallery toolbar (`AppLayout.tsx`'s toolbar row, alongside the search input — per Option A of the "Filter & Sort" design handoff) that opens a panel for choosing a sort field and direction. The control SHALL be built from this codebase's existing `DropdownMenu`/`DropdownMenuTrigger`/`DropdownMenuContent` and `DropdownMenuRadioGroup`/`DropdownMenuRadioItem` primitives (the same family already used for the upload split-button), not a bespoke floating-panel implementation. Below the `sm` breakpoint, the sort icon button SHALL NOT be rendered.

#### Scenario: Sort button opens the sort panel

- **WHEN** the user clicks the sort icon button in the toolbar at or above the `sm` breakpoint
- **THEN** a panel opens showing sort field options as a radio group, plus a direction toggle (when applicable)

#### Scenario: Sort panel closes on selection or outside click

- **WHEN** the user selects a sort field, toggles direction, or clicks outside the panel
- **THEN** the panel closes (or remains open per the underlying `DropdownMenu` primitive's standard interaction — selection does not require a separate "apply" step)

#### Scenario: Sort button is hidden below the breakpoint

- **WHEN** the viewport width is below the `sm` breakpoint
- **THEN** the sort icon button is not rendered in the gallery toolbar

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

### Requirement: Manual sort reproduces today's default ordering by sending no sort parameters

Selecting `Manual` SHALL cause the image list request to omit `sort` and `direction` entirely — reproducing exactly the request shape used before this change existed. The backend then applies its existing default (`image_folders.position ASC` for folder views). `Manual` is a frontend-only framing of "no explicit sort requested," not a value sent to or recognized by the backend.

#### Scenario: Selecting Manual omits sort parameters from the request

- **WHEN** the user selects `Manual` in a folder view
- **THEN** `GET /images/in-folder/<id>` is called with no `sort` or `direction` query parameters
- **AND** images are displayed in `image_folders.position ASC` order, matching pre-change behavior

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

### Requirement: Sort trigger shows an active indicator when a non-default sort is selected

The sort icon button SHALL visually distinguish itself when the active sort differs from the current view's default (i.e. the user has made an explicit, non-default choice), using this codebase's existing button-variant styling rather than introducing a new badge/indicator pattern. "Active" is computed as: the selected field differs from the view's default field, OR (the field is not `Manual` AND the selected direction differs from that field's default direction).

#### Scenario: Trigger appears active when a non-default sort is selected

- **WHEN** the user selects `Name` while viewing a folder (whose default is `Manual`)
- **THEN** the sort icon button renders in its active visual state

#### Scenario: Trigger appears inactive when the view's default sort is selected

- **WHEN** the active sort matches the current view's default (e.g. `Manual` in a folder, or `Date added`/`desc` in All)
- **THEN** the sort icon button renders in its normal (inactive) visual state

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
