## ADDED Requirements

### Requirement: Warm theme is the default active theme
The system SHALL apply the "warm" named theme by default. The warm theme defines the application's color palette (`background`, `foreground`, `card`, `popover`, `muted`, `accent`, `secondary`, `primary`, `border`, `input`, `ring`, `sidebar`, and their `-foreground` counterparts, plus `radius`) and typography (DM Sans for UI text, Lora for the application wordmark).

#### Scenario: Warm theme renders on initial load
- **WHEN** the application loads with no stored theme preference
- **THEN** the `<html>` element has `data-theme="warm"`
- **AND** the warm palette CSS variables are in effect

### Requirement: Theme preference provider
The system SHALL provide a `ThemeProvider` exposing a `useTheme()` hook. `useTheme()` SHALL return the current theme and a setter function. `ThemeProvider` SHALL apply the active theme via a `data-theme` attribute on the `<html>` element and SHALL persist the theme preference to `localStorage`.

#### Scenario: Default theme with no stored preference
- **WHEN** no theme preference is stored in `localStorage`
- **THEN** `useTheme()` returns `"warm"`
- **AND** the `<html>` element has `data-theme="warm"`

#### Scenario: Stored preference is restored on load
- **WHEN** `localStorage` contains a previously-saved theme preference
- **THEN** `ThemeProvider` initializes `useTheme()` with that stored value
- **AND** applies it via the `data-theme` attribute on `<html>`

#### Scenario: Setting a theme persists and applies it
- **WHEN** `setTheme` is called with a valid theme value
- **THEN** the `<html>` element's `data-theme` attribute updates to that value
- **AND** the value is written to `localStorage`

#### Scenario: useTheme used outside provider throws
- **WHEN** `useTheme()` is called from a component not wrapped in `ThemeProvider`
- **THEN** an error is thrown

### Requirement: dark: utilities key off the active theme
The system SHALL define the Tailwind `dark:` variant to apply only when the active theme is `"dark"` (i.e. `data-theme="dark"` on `<html>` or an ancestor), rather than based on the OS `prefers-color-scheme` setting.

#### Scenario: dark: utilities inactive under the warm theme
- **WHEN** the active theme is `"warm"`
- **THEN** elements using `dark:` utility classes do not apply their dark-variant styles, regardless of the OS color-scheme preference
