## ADDED Requirements

### Requirement: Search input scoped to active view
The system SHALL display a search input in the gallery toolbar — `AppLayout.tsx`'s `flex justify-between` toolbar row above `ImageGrid`, positioned at the leftmost edge of the row with the "Image" upload button at the right — that filters images by title within the currently active view (folder / All / Unsorted / Trash). The search behavior differs by view: folder views filter the already-loaded image list client-side, while All/Unsorted/Trash views query the backend. Because the input must render in `AppLayout` while the views it filters are rendered by `ImageGrid`, the search term and its debounced value live as state in `AppLayout` and are passed down to `ImageGrid` as props (`searchTerm`, `debouncedSearchTerm`).

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

### Requirement: Search term resets when switching views
The system SHALL clear the search input and any active title filter whenever the user navigates to a different view.

#### Scenario: Switching views clears an active search term
- **WHEN** the user has typed a search term in one view and then navigates to a different view (a different folder, or All/Unsorted/Trash)
- **THEN** the search input is cleared
- **AND** the new view displays its unfiltered image list

### Requirement: Backend search is debounced
The system SHALL debounce search requests sent to the backend so that a request is only issued after the user has paused typing, rather than on every keystroke.

#### Scenario: Rapid typing produces a single request
- **WHEN** the user types several characters in quick succession while viewing All, Unsorted, or Trash
- **THEN** only one request containing the final search term is sent, after the debounce delay elapses since the last keystroke

### Requirement: Grid retains previous results during debounced search
The system SHALL keep the previously displayed image results visible while a new debounced search request is in flight, rather than clearing the grid to a loading state.

#### Scenario: Previous results remain visible while new results load
- **WHEN** the user changes the search term while viewing All, Unsorted, or Trash and a new request is in flight
- **THEN** the grid continues to display the previous results until the new results arrive
