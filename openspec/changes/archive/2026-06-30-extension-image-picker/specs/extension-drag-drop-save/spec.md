## ADDED Requirements

### Requirement: Drop zone color scheme respects dark mode preference

The content script SHALL read `getDarkMode()` from storage each time the drop zone is rendered. If dark mode is enabled, the drop zone SHALL render using the dark color palette. If dark mode is disabled or unset, it SHALL use the light color palette. Theming SHALL be applied via the same CSS custom property and `data-theme` mechanism used by the toast and picker overlay.

#### Scenario: Drop zone displays with light palette when dark mode is off

- **WHEN** `getDarkMode()` resolves to `false` and a drag gesture triggers the drop zone
- **THEN** the drop zone renders with the light color palette

#### Scenario: Drop zone displays with dark palette when dark mode is on

- **WHEN** `getDarkMode()` resolves to `true` and a drag gesture triggers the drop zone
- **THEN** the drop zone renders with the dark color palette
