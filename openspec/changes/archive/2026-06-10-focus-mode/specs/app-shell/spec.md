## MODIFIED Requirements

### Requirement: Two-panel application shell
The system SHALL render a persistent two-panel layout consisting of a fixed left sidebar (240 px wide) and a fluid right content area that fills the remaining viewport width, except while focus mode is active, in which case the left sidebar SHALL NOT be rendered and the main content area SHALL fill the full viewport width.

#### Scenario: Layout renders on load
- **WHEN** the application root is mounted
- **THEN** the sidebar and main content area are both visible on screen simultaneously

#### Scenario: Sidebar does not scroll with content
- **WHEN** the main content area is scrolled
- **THEN** the sidebar remains fixed in place and does not move

#### Scenario: Sidebar is hidden while focus mode is active
- **WHEN** focus mode is active
- **THEN** the left sidebar is not rendered
- **AND** the main content area fills the full viewport width
