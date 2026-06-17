## MODIFIED Requirements

### Requirement: Sort control in gallery toolbar

The system SHALL display a sort icon button in the gallery toolbar (`AppLayout.tsx`'s toolbar row, alongside the search input — per Option A of the "Filter & Sort" design handoff) that opens a panel for choosing a sort field and direction. The control SHALL be built from this codebase's existing `DropdownMenu`/`DropdownMenuTrigger`/`DropdownMenuContent` and `DropdownMenuRadioGroup`/`DropdownMenuRadioItem` primitives (the same family already used for the upload split-button), not a bespoke floating-panel implementation. Below the `sm` breakpoint, the sort icon button SHALL NOT be rendered.

#### Scenario: Sort button opens the sort panel

- **WHEN** the user clicks the sort icon button in the toolbar at or above the `sm` breakpoint
- **THEN** a panel opens showing sort field options as a radio group, plus a direction toggle (when applicable)

#### Scenario: Sort panel closes on selection or outside click

- **WHEN** the user selects a sort field, toggles direction, or clicks outside the panel
- **THEN** the panel closes (or remains open per the underlying `DropdownMenu` primitive's standard interaction — selection does not require a separate "apply" step)

#### Scenario: Sort button is hidden below the breakpoint

- **WHEN** the viewport width is below the `sm` breakpoint
- **THEN** the sort icon button is not rendered in the gallery toolbar
