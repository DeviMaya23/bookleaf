## MODIFIED Requirements

### Requirement: Two-panel application shell
The system SHALL render a persistent two-panel layout consisting of a fixed left sidebar (240 px wide) and a fluid right content area that fills the remaining viewport width, except while focus mode is active, in which case the left sidebar SHALL NOT be rendered and the main content area SHALL fill the full viewport width. Below the `sm` breakpoint, the left sidebar SHALL instead be hidden off-canvas by default (see `fe-mobile-shell`) and the main content area SHALL fill the full viewport width regardless of focus mode state.

#### Scenario: Layout renders on load
- **WHEN** the application root is mounted at or above the `sm` breakpoint
- **THEN** the sidebar and main content area are both visible on screen simultaneously

#### Scenario: Sidebar does not scroll with content
- **WHEN** the main content area is scrolled
- **THEN** the sidebar remains fixed in place and does not move

#### Scenario: Sidebar is hidden while focus mode is active
- **WHEN** focus mode is active at or above the `sm` breakpoint
- **THEN** the left sidebar is not rendered
- **AND** the main content area fills the full viewport width

#### Scenario: Main content fills the viewport below the breakpoint
- **WHEN** the viewport width is below the `sm` breakpoint
- **THEN** the main content area fills the full viewport width, whether or not the sidebar drawer is open
