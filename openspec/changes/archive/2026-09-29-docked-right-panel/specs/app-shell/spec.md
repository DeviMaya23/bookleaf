## MODIFIED Requirements

### Requirement: Two-panel application shell

The system SHALL render a persistent three-column layout on desktop (at or above the `sm` breakpoint, fine-pointer devices) consisting of: a fixed left sidebar (240 px wide), a fluid main content area, and a fixed right panel (320 px expanded, 32 px collapsed). The right panel is always present; its width changes only via the user's explicit collapse/expand toggle (per `fe-right-panel-docked`). While focus mode is active, both the left sidebar and the right panel SHALL NOT be rendered, and the main content area SHALL fill the full viewport width.

Below the `sm` breakpoint, the left sidebar SHALL instead be hidden off-canvas by default (see `fe-mobile-shell`), the right panel SHALL NOT be rendered (coarse-pointer drawer behavior applies instead, per `fe-right-panel`), and the main content area SHALL fill the full viewport width regardless of focus mode state.

#### Scenario: Layout renders on load

- **WHEN** the application root is mounted at or above the `sm` breakpoint on a fine-pointer device
- **THEN** the left sidebar, main content area, and right panel are all visible simultaneously

#### Scenario: Sidebar does not scroll with content

- **WHEN** the main content area is scrolled
- **THEN** the sidebar remains fixed in place and does not move

#### Scenario: Both panels are hidden while focus mode is active

- **WHEN** focus mode is active at or above the `sm` breakpoint
- **THEN** neither the left sidebar nor the right panel is rendered
- **AND** the main content area fills the full viewport width

#### Scenario: Main content fills the viewport below the breakpoint

- **WHEN** the viewport width is below the `sm` breakpoint
- **THEN** the main content area fills the full viewport width, whether or not the sidebar drawer is open
