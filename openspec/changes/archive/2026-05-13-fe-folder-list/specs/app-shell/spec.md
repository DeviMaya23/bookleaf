## MODIFIED Requirements

### Requirement: Folder list in sidebar
The system SHALL fetch the folder list from `GET /folders` and display it in the sidebar. An "Unsorted" entry SHALL be pinned permanently at the top of the list, visually separated from the API-sourced folders by a horizontal divider. The "+ New folder" affordance SHALL remain visible below the folder list.

#### Scenario: Folder list is populated from API
- **WHEN** the application shell is rendered and `GET /folders` returns a non-empty list
- **THEN** each folder from the API is displayed in the sidebar below the "Unsorted" entry and divider

#### Scenario: Unsorted entry is always first
- **WHEN** the application shell is rendered
- **THEN** "Unsorted" appears at the top of the sidebar folder list regardless of API response order

#### Scenario: Divider separates Unsorted from API folders
- **WHEN** the application shell is rendered
- **THEN** a horizontal divider is rendered between the "Unsorted" entry and the API folder list

#### Scenario: New folder affordance is visible
- **WHEN** the application shell is rendered
- **THEN** a "+ New folder" button is visible below the folder list

#### Scenario: Empty API folder list
- **WHEN** the application shell is rendered and `GET /folders` returns an empty list
- **THEN** only the "Unsorted" entry is shown, with the divider and "+ New folder" button still present

## MODIFIED Requirements

### Requirement: Empty image grid in main area
The system SHALL render an empty image grid placeholder in the main content area. No real image data is required — the grid shell (container and spacing) SHALL be present.

#### Scenario: Image grid area is present
- **WHEN** the application shell is rendered
- **THEN** the main content area contains an image grid container element
