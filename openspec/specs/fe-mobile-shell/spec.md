## Purpose

Mobile-responsive shell behavior below the `sm` breakpoint: a fixed mobile top bar with hamburger access to the sidebar drawer, the off-canvas sidebar drawer interaction, and a floating action button for uploads.

## Requirements

### Requirement: Mobile top bar

Below the `sm` breakpoint, the system SHALL render a fixed top bar consisting of a hamburger button on the left and the centered "Bookleaf" wordmark, except while focus mode is active, in which case the top bar SHALL NOT be rendered (consistent with the sidebar it opens also not rendering during focus mode, per `app-shell`). At `sm` and above, this top bar SHALL NOT be rendered.

#### Scenario: Top bar renders below the breakpoint
- **WHEN** the viewport width is below the `sm` breakpoint and focus mode is not active
- **THEN** a fixed top bar is visible with a hamburger button and the centered "Bookleaf" wordmark

#### Scenario: Top bar does not render at or above the breakpoint
- **WHEN** the viewport width is at or above the `sm` breakpoint
- **THEN** the mobile top bar is not rendered

#### Scenario: Top bar does not render while focus mode is active
- **WHEN** focus mode is active, regardless of viewport width
- **THEN** the mobile top bar is not rendered

### Requirement: Sidebar drawer interaction below the breakpoint

Below the `sm` breakpoint, the left sidebar (`FolderSidebar`) SHALL be hidden off-canvas by default. Tapping the hamburger button in the mobile top bar SHALL slide the sidebar into view and display a backdrop behind it. Tapping the backdrop, or selecting a navigation entry within the sidebar, SHALL close the drawer. At `sm` and above, the sidebar SHALL remain always visible as today, and the drawer/backdrop behavior SHALL NOT apply. While focus mode is active, neither the drawer nor its backdrop SHALL be reachable, since the mobile top bar (the only entry point to open the drawer) is not rendered in that state.

#### Scenario: Sidebar is hidden by default below the breakpoint
- **WHEN** the viewport width is below the `sm` breakpoint and the drawer has not been opened
- **THEN** the sidebar is not visible on screen

#### Scenario: Hamburger button opens the drawer
- **WHEN** the user taps the hamburger button in the mobile top bar
- **THEN** the sidebar slides into view and a backdrop is displayed behind it

#### Scenario: Tapping the backdrop closes the drawer
- **WHEN** the drawer is open and the user taps the backdrop
- **THEN** the sidebar slides back off-canvas and the backdrop is removed

#### Scenario: Selecting a navigation entry closes the drawer
- **WHEN** the drawer is open and the user selects a folder, "All", "Unsorted", or "Trash" entry in the sidebar
- **THEN** the drawer closes after navigating to the selected view

#### Scenario: Drawer behavior does not apply at or above the breakpoint
- **WHEN** the viewport width is at or above the `sm` breakpoint
- **THEN** the sidebar is always visible and no backdrop is rendered, regardless of drawer open/close state

### Requirement: Floating upload button below the breakpoint

Below the `sm` breakpoint, the system SHALL render a floating action button (FAB), fixed to the bottom-right of the viewport, that opens the same upload modal as the existing "+Image" toolbar action. At `sm` and above, the FAB SHALL NOT be rendered.

#### Scenario: FAB renders below the breakpoint
- **WHEN** the viewport width is below the `sm` breakpoint
- **THEN** a floating circular upload button is visible at the bottom-right of the viewport

#### Scenario: FAB opens the upload modal
- **WHEN** the user taps the FAB
- **THEN** the upload modal opens

#### Scenario: FAB does not render at or above the breakpoint
- **WHEN** the viewport width is at or above the `sm` breakpoint
- **THEN** the floating upload button is not rendered
