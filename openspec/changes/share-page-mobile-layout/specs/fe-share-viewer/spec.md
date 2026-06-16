## ADDED Requirements

### Requirement: SharePage responsive layout

At viewport widths below the `sm:` breakpoint (640px), `SharePage` SHALL render as a single unified scrollable column. `SharedFolderPanel` SHALL appear above the masonry gallery in this column. At `sm:` and above, the existing side-by-side layout (gallery left, panel right) SHALL be preserved unchanged.

The layout transition SHALL be achieved using Tailwind responsive prefixes only — no JavaScript resize listeners or conditional rendering.

#### Scenario: Panel stacks above gallery on phone-width viewports

- **WHEN** `SharePage` is rendered at a viewport width below 640px and the share data has loaded successfully
- **THEN** `SharedFolderPanel` is displayed above the masonry gallery in a single scrollable column
- **AND** no horizontal overflow or clipping occurs

#### Scenario: Desktop side-by-side layout is preserved

- **WHEN** `SharePage` is rendered at a viewport width of 640px or above
- **THEN** `SharedFolderPanel` appears as a fixed-width right-side panel alongside the masonry gallery
- **AND** the gallery and panel each scroll independently

### Requirement: SharedFolderPanel mobile mode

At viewport widths below the `sm:` breakpoint (640px), `SharedFolderPanel` SHALL render as full-width inline flow content: no fixed height, no internal independent scroll, no sticky export footer. A `border-b` SHALL separate the panel from the gallery below it. At `sm:` and above, the panel SHALL restore its current form: fixed `w-[280px]`, `h-full`, `border-l`, with the export button in a sticky footer.

#### Scenario: Panel renders full-width with inline export on mobile

- **WHEN** `SharedFolderPanel` is rendered at a viewport width below 640px
- **THEN** it spans the full available width
- **AND** the export button is inline with the rest of the panel content, not pinned to the bottom of the viewport

#### Scenario: Panel restores fixed-width side-column layout at sm: and above

- **WHEN** `SharedFolderPanel` is rendered at a viewport width of 640px or above
- **THEN** it renders at 280px wide with a left border
- **AND** the export button is pinned to the bottom of the panel via a sticky footer
