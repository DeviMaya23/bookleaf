# fe-gallery-sort

## Purpose

TBD — covers the gallery's sort control: the toolbar UI for choosing a sort field and direction, per-view field options and defaults, the Manual sort framing for folder views, the active-sort indicator, and threading the selected sort through to the image list query.

## Requirements

### Requirement: Sort control in gallery toolbar

The system SHALL display a sort icon button in the gallery toolbar (`AppLayout.tsx`'s toolbar row, alongside the search input — per Option A of the "Filter & Sort" design handoff) that opens a panel for choosing a sort field and direction. The control SHALL be built from this codebase's existing `DropdownMenu`/`DropdownMenuTrigger`/`DropdownMenuContent` and `DropdownMenuRadioGroup`/`DropdownMenuRadioItem` primitives (the same family already used for the upload split-button), not a bespoke floating-panel implementation.

#### Scenario: Sort button opens the sort panel

- **WHEN** the user clicks the sort icon button in the toolbar
- **THEN** a panel opens showing sort field options as a radio group, plus a direction toggle (when applicable)

#### Scenario: Sort panel closes on selection or outside click

- **WHEN** the user selects a sort field, toggles direction, or clicks outside the panel
- **THEN** the panel closes (or remains open per the underlying `DropdownMenu` primitive's standard interaction — selection does not require a separate "apply" step)

### Requirement: Sort field options differ by view type

The sort panel's radio list SHALL show different options depending on the active view:

- **Folder views**: `Manual`, `Date added`, `Name` — with `Manual` listed first
- **All / Unsorted / Trash views**: `Date added`, `Name` — `Manual` is not offered, since these views have no persisted manual order (`position` is folder-scoped)

`file_size` and `dimensions` (shown in the design handoff) are NOT offered in any view, since the backend's sort allow-list (`image-list-sort`) does not support them.

#### Scenario: Folder view shows Manual as the first option

- **WHEN** the user opens the sort panel while viewing a folder
- **THEN** the radio list shows `Manual`, `Date added`, `Name`, in that order, with `Manual` selected by default

#### Scenario: Non-folder views omit Manual

- **WHEN** the user opens the sort panel while viewing All, Unsorted, or Trash
- **THEN** the radio list shows only `Date added` and `Name`

### Requirement: Manual sort reproduces today's default ordering by sending no sort parameters

Selecting `Manual` SHALL cause the image list request to omit `sort` and `direction` entirely — reproducing exactly the request shape used before this change existed. The backend then applies its existing default (`image_folders.position ASC` for folder views). `Manual` is a frontend-only framing of "no explicit sort requested," not a value sent to or recognized by the backend.

#### Scenario: Selecting Manual omits sort parameters from the request

- **WHEN** the user selects `Manual` in a folder view
- **THEN** `GET /images/in-folder/<id>` is called with no `sort` or `direction` query parameters
- **AND** images are displayed in `image_folders.position ASC` order, matching pre-change behavior

### Requirement: Direction toggle is shown only for orderable (non-Manual) sort fields

The sort panel SHALL display a direction toggle (`↑`/`↓`, with field-specific labels — e.g. "Oldest first"/"Newest first" for `Date added`, "A → Z"/"Z → A" for `Name`) whenever the selected sort field is `Date added` or `Name`. The toggle SHALL be omitted entirely — not shown-but-disabled — when `Manual` is selected, since direction has no meaning for a manually-ordered list.

#### Scenario: Direction toggle visible for Date added or Name

- **WHEN** the user selects `Date added` or `Name` as the sort field
- **THEN** a direction toggle appears showing the field-appropriate label for the current direction
- **AND** clicking it flips the direction and updates the label and the active query

#### Scenario: Direction toggle hidden for Manual

- **WHEN** the user selects `Manual` as the sort field
- **THEN** no direction toggle is shown in the panel

### Requirement: Sort selection resets to the view's default when switching views

The system SHALL reset the active sort field and direction to the new view's default whenever the user navigates to a different view (a different folder, or All/Unsorted/Trash) — mirroring the existing search-term reset behavior (`fe-gallery-search`). Sort state is per-view and session-local; it is never carried over from one view to another nor persisted across sessions.

Per-view defaults:

| View | Default field | Default direction |
|---|---|---|
| Folder | `Manual` | n/a |
| All / Unsorted / Trash | `Date added` | `desc` (newest first) |

These defaults reproduce each view's pre-change ordering exactly, so the new control describes existing behavior on first load rather than changing it.

#### Scenario: Switching from a folder to All resets sort to Date added (newest first)

- **WHEN** the user has an explicit sort active in a folder (e.g. `Name`, ascending) and navigates to the "All" view
- **THEN** the sort control shows `Date added` / "Newest first" as selected
- **AND** the image list is requested with no explicit sort change needed beyond the view's own default (i.e. the All view's default request, matching pre-change behavior)

#### Scenario: Switching between folders resets sort to Manual

- **WHEN** the user has an explicit sort active in one folder and navigates to a different folder
- **THEN** the sort control shows `Manual` as selected for the new folder
- **AND** images are requested with no `sort`/`direction` parameters (the new folder's `position`-ordered list)

### Requirement: Sort trigger shows an active indicator when a non-default sort is selected

The sort icon button SHALL visually distinguish itself when the active sort differs from the current view's default (i.e. the user has made an explicit, non-default choice), using this codebase's existing button-variant styling rather than introducing a new badge/indicator pattern. "Active" is computed as: the selected field differs from the view's default field, OR (the field is not `Manual` AND the selected direction differs from that field's default direction).

#### Scenario: Trigger appears active when a non-default sort is selected

- **WHEN** the user selects `Name` while viewing a folder (whose default is `Manual`)
- **THEN** the sort icon button renders in its active visual state

#### Scenario: Trigger appears inactive when the view's default sort is selected

- **WHEN** the active sort matches the current view's default (e.g. `Manual` in a folder, or `Date added`/`desc` in All)
- **THEN** the sort icon button renders in its normal (inactive) visual state

### Requirement: Selected sort is threaded through to the image list query

`getImages` and `getAllImages` (`src/lib/images.ts`) SHALL accept optional `sort` and `direction` parameters and include them as query parameters on `GET /images` when present. `getFolderImages` (`src/lib/images.ts`) SHALL accept optional `sort` and `direction` parameters and include them as query parameters on `GET /images/in-folder/:id` when present (omitted entirely when `Manual` is selected — see above). `ImageGrid`'s query key SHALL include the active `sort`/`direction` so that changing the sort triggers a re-fetch, and `useInfiniteQuery`'s existing `placeholderData: keepPreviousData` SHALL keep the prior results visible during the transition (mirroring how `debouncedSearchTerm` changes are handled today).

#### Scenario: Selecting an explicit sort re-fetches with the new ordering

- **WHEN** the user selects `Name` / ascending while viewing the "All" view
- **THEN** `GET /images` is called with `sort=title&direction=asc`
- **AND** the displayed images update to the new ordering once the response arrives, with the previous results remaining visible during the fetch

#### Scenario: Selecting Manual re-fetches with no sort parameters

- **WHEN** the user selects `Manual` while viewing a folder that previously had an explicit sort active
- **THEN** `GET /images/in-folder/<id>` is called with no `sort`/`direction` parameters
- **AND** the displayed images update to `position`-ordered results
