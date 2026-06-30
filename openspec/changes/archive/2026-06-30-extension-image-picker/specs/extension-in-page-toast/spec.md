## ADDED Requirements

### Requirement: Toast color scheme respects dark mode preference

The content script SHALL read `getDarkMode()` from storage each time a toast is displayed. If dark mode is enabled, the toast SHALL render using the dark color palette (dark background, light text, dark border accent). If dark mode is disabled or the preference is unset, the toast SHALL render using the light color palette (white background, dark text). The color palettes SHALL be defined as CSS custom properties on the shadow host element and toggled via a `data-theme` attribute, consistent with how the picker overlay and drop zone apply theming.

#### Scenario: Toast displays with light palette when dark mode is off

- **WHEN** `getDarkMode()` resolves to `false` and a toast message is received
- **THEN** the toast uses the light color palette (white background, dark text)

#### Scenario: Toast displays with dark palette when dark mode is on

- **WHEN** `getDarkMode()` resolves to `true` and a toast message is received
- **THEN** the toast uses the dark color palette (dark background, light text)
