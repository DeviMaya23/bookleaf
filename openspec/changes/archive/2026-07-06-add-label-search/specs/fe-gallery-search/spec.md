## ADDED Requirements

### Requirement: Label search toggle

The system SHALL display a "Search in AI labels" toggle inside the gallery filter dropdown, visible only when the authenticated user has `vision_enabled = true`. The toggle SHALL be rendered as a `DropdownMenuCheckboxItem` in a dedicated section within the dropdown.

The toggle's checked state SHALL be managed as `searchLabels: boolean` in `useGalleryControls`. The hook SHALL accept a `visionEnabled: boolean` parameter; `AppLayout` SHALL pass `me.vision_enabled ?? false` into it after fetching the current user via `getMe`. `AppLayout` already calls `useGalleryControls` — it SHALL add a `getMe` query (same pattern as `useVisionSuggestion`) and pass the resolved flag.

`searchLabels` SHALL default to `false` and SHALL reset to `false` whenever the active view changes (same `useEffect` that resets other filter state).

`filterCount` (the badge on the Filters button) SHALL NOT include the `searchLabels` toggle — it is a search-scope modifier, not a data filter.

A new `'labelSearch'` entry SHALL be added to the `FilterSection` union type and returned by `filterSectionsForViewType` for the `'all'` and `'unsorted'` view types only, when `visionEnabled` is true. Because `filterSectionsForViewType` is currently a pure function of `viewType`, it SHALL be updated to also accept `visionEnabled: boolean`.

#### Scenario: Toggle is visible for vision-enabled users on All/Unsorted views

- **WHEN** the filter dropdown is opened while the user has `vision_enabled = true` and the active view is All or Unsorted
- **THEN** a "Search in AI labels" checkbox item is rendered in the dropdown

#### Scenario: Toggle is hidden for users without vision_enabled

- **WHEN** the filter dropdown is opened and the user has `vision_enabled = false`
- **THEN** no label search toggle appears in the dropdown

#### Scenario: Toggle resets when switching views

- **WHEN** the user has the label search toggle checked and navigates to a different view
- **THEN** `searchLabels` resets to `false`

#### Scenario: Toggle does not increment the filter badge count

- **WHEN** the label search toggle is the only active "filter" (no tags, mime types, or folders selected)
- **THEN** the Filters button shows no count badge

## MODIFIED Requirements

### Requirement: Search input scoped to active view

The system SHALL display a search input in the gallery toolbar — `AppLayout.tsx`'s `flex justify-between` toolbar row above `ImageGrid`, positioned at the leftmost edge of the row with the "Image" upload button at the right — that filters images by title within the currently active view (folder / All / Unsorted / Trash). The search behavior differs by view: folder views filter the already-loaded image list client-side, while All/Unsorted/Trash views query the backend. Because the input must render in `AppLayout` while the views it filters are rendered by `ImageGrid`, the search term and its debounced value live as state in `AppLayout` and are passed down to `ImageGrid` as props (`searchTerm`, `debouncedSearchTerm`). The `searchLabels` boolean from `useGalleryControls` is also passed to `ImageGrid` and forwarded to `useGalleryImages`, which includes it in the query key and passes it to the relevant API functions.

#### Scenario: Search input is positioned at the leftmost edge of the toolbar row

- **WHEN** the gallery toolbar is rendered
- **THEN** the search input appears in the same row as the "Image" upload button, at the leftmost edge of the row, with the upload button at the right

#### Scenario: Searching within a folder view filters client-side

- **WHEN** the user types a term into the search input while viewing a folder
- **THEN** the displayed images are filtered to those whose title contains the term, case-insensitively
- **AND** no additional network request is made

#### Scenario: Searching within All/Unsorted/Trash queries the backend

- **WHEN** the user types a term into the search input while viewing All, Unsorted, or Trash
- **THEN** after the debounce delay elapses, the system requests the image list with a `name` query parameter set to the term
- **AND** the displayed images are replaced with the filtered, paginated results

#### Scenario: Label search toggle sends search_labels=true with name

- **WHEN** the user has the label search toggle checked and types a term into the search input while viewing All or Unsorted
- **THEN** after the debounce delay elapses, the system requests the image list with both `name=<term>` and `search_labels=true` query parameters
- **AND** the displayed images include both title matches and AI label matches for that term
