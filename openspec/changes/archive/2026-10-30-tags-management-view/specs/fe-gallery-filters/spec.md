# Spec Delta

## MODIFIED Requirements

### Requirement: Filter selections reset when switching views

The system SHALL clear all selected tags, file types, and folders whenever the user navigates to a different view (a different folder, or All/Unsorted/Trash) — mirroring the existing search-term and sort reset behavior (`fe-gallery-search`, `fe-gallery-sort`). Filter state is per-view and session-local; it is never carried over from one view to another nor persisted across sessions.

**Exception**: when navigating to the All view from the Tags view, the `filterTagIds` initial value MAY be seeded from router state (`{ filterTagIds: string[] }`). This is the only case where an initial filter value is carried into a view; all other navigation still resets filters to empty.

#### Scenario: Switching views clears active filters

- **WHEN** the user has active filters selected in one view and navigates to a different view (not via a tag click from the Tags view)
- **THEN** all filter selections are cleared
- **AND** the new view displays its unfiltered image list (subject to that view's own default request)

#### Scenario: Navigating from Tags view pre-seeds tag filter

- **WHEN** the user clicks a tag in the Tags view
- **THEN** the route changes to `/app`
- **AND** the tag filter is initialised with that tag's ID selected
- **AND** the image grid shows only images tagged with the selected tag
