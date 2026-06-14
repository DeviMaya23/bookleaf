## MODIFIED Requirements

### Requirement: dark: utilities key off the active theme

The system SHALL define the Tailwind `dark:` variant to apply only when the active theme is `"sunless"` (i.e. `data-theme="sunless"` on `<html>` or an ancestor), rather than based on the OS `prefers-color-scheme` setting or an unused `"dark"` theme value.

#### Scenario: dark: utilities inactive under the warm or lumen themes

- **WHEN** the active theme is `"warm"` or `"lumen"`
- **THEN** elements using `dark:` utility classes do not apply their dark-variant styles, regardless of the OS color-scheme preference

#### Scenario: dark: utilities active under the sunless theme

- **WHEN** the active theme is `"sunless"`
- **THEN** elements using `dark:` utility classes apply their dark-variant styles

## ADDED Requirements

### Requirement: Lumen theme

The system SHALL provide a `"lumen"` named theme, selectable via `setTheme('lumen')`, defining a bright neutral-white color palette (`background`, `foreground`, `card`, `popover`, `muted`, `accent`, `secondary`, `primary`, `border`, `input`, `ring`, `sidebar`, `destructive`, and their `-foreground` counterparts, plus `radius`).

#### Scenario: Lumen theme applies its palette

- **WHEN** the active theme is `"lumen"`
- **THEN** the relevant element has `data-theme="lumen"`
- **AND** the lumen palette CSS variables are in effect

### Requirement: Sunless theme

The system SHALL provide a `"sunless"` named theme, selectable via `setTheme('sunless')`, defining a near-black color palette (`background`, `foreground`, `card`, `popover`, `muted`, `accent`, `secondary`, `primary`, `border`, `input`, `ring`, `sidebar`, `destructive`, and their `-foreground` counterparts, plus `radius`).

#### Scenario: Sunless theme applies its palette

- **WHEN** the active theme is `"sunless"`
- **THEN** the relevant element has `data-theme="sunless"`
- **AND** the sunless palette CSS variables are in effect

### Requirement: Theme picker in Settings

The Settings App section SHALL render all available themes (`warm`, `lumen`, `sunless`) as a selectable list, each entry showing a label, short description, and color swatches. The entry matching the current active theme SHALL be shown as selected. Selecting a different entry SHALL call `setTheme` with that theme's id.

#### Scenario: Active theme is shown as selected

- **WHEN** the App section is displayed
- **THEN** the entry for the active theme is shown as selected
- **AND** the other entries are not shown as selected

#### Scenario: Selecting a theme updates the active theme

- **WHEN** the user selects a theme entry other than the currently active one
- **THEN** `setTheme` is called with that entry's theme id

### Requirement: Public pages render in the warm theme

The system SHALL render all public/marketing routes (`/`, `/about`, `/privacy`, `/ai-notes`) with `data-theme="warm"`, regardless of the user's stored theme preference. The authenticated app (`/app/*`) SHALL continue to render with the user's stored theme preference.

#### Scenario: Public routes ignore the stored theme preference

- **WHEN** a user with a stored theme preference of `"lumen"` or `"sunless"` navigates to `/`, `/about`, `/privacy`, or `/ai-notes`
- **THEN** that page renders with the warm palette

#### Scenario: Authenticated app uses the stored preference

- **WHEN** a user with a stored theme preference of `"lumen"` or `"sunless"` navigates to `/app`
- **THEN** the app renders with the `lumen` or `sunless` palette accordingly

### Requirement: Initial theme is applied before first paint

The system SHALL set the `data-theme` attribute on `<html>` from the stored theme preference (or default to `"warm"` if none is stored or the stored value is unrecognized) before the page's first paint, preventing a flash of an incorrect theme.

#### Scenario: Stored theme applies before first paint

- **WHEN** the page loads with a theme preference stored in `localStorage`
- **THEN** `<html>` has the corresponding `data-theme` attribute set before first paint

#### Scenario: No stored preference defaults to warm before first paint

- **WHEN** the page loads with no theme preference stored in `localStorage`
- **THEN** `<html>` has `data-theme="warm"` set before first paint
