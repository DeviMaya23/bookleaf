## MODIFIED Requirements

### Requirement: System folder navigation entries
The sidebar SHALL display three pinned system entries at the top of the navigation: **All**, **Unsorted**, and **Trash**. All and Unsorted SHALL use the default text color. Trash SHALL use a muted/de-emphasized color. Selecting a system entry SHALL navigate to the corresponding route (`/app`, `/app/unsorted`, `/app/trash`).

#### Scenario: All three system entries are visible
- **WHEN** the sidebar is rendered
- **THEN** "All", "Unsorted", and "Trash" entries are displayed in that order at the top of the nav

#### Scenario: Active system entry is highlighted
- **WHEN** the current route matches a system entry's route
- **THEN** that entry is shown with an active/selected style

#### Scenario: Trash entry is visually de-emphasized
- **WHEN** the sidebar is rendered
- **THEN** the "Trash" entry uses a muted color distinct from "All" and "Unsorted"
