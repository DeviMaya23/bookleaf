## ADDED Requirements

### Requirement: Two-panel application shell
The system SHALL render a persistent two-panel layout consisting of a fixed left sidebar (240 px wide) and a fluid right content area that fills the remaining viewport width.

#### Scenario: Layout renders on load
- **WHEN** the application root is mounted
- **THEN** the sidebar and main content area are both visible on screen simultaneously

#### Scenario: Sidebar does not scroll with content
- **WHEN** the main content area is scrolled
- **THEN** the sidebar remains fixed in place and does not move

### Requirement: Folder list in sidebar
The system SHALL display a placeholder folder list in the sidebar containing at minimum: "Unfiled", "Nature", and "Travel" entries, followed by a "+ New folder" affordance.

#### Scenario: Folder list is visible
- **WHEN** the application shell is rendered
- **THEN** each placeholder folder name is displayed in the sidebar in order

#### Scenario: New folder affordance is visible
- **WHEN** the application shell is rendered
- **THEN** a "+ New folder" label or button is visible at the bottom of the folder list

### Requirement: Empty image grid in main area
The system SHALL render an empty image grid placeholder in the main content area. No real image data is required — the grid shell (container and spacing) SHALL be present.

#### Scenario: Image grid area is present
- **WHEN** the application shell is rendered
- **THEN** the main content area contains an image grid container element

#### Scenario: No API calls are made
- **WHEN** the application shell is rendered
- **THEN** no network requests are issued to any backend endpoint
