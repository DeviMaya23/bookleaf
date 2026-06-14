## Purpose

TBD - created by archiving change settings-modal. Update Purpose after archive.

## Requirements

### Requirement: SettingsModal shell and navigation

The system SHALL provide a `SettingsModal` component, opened via the `ProfileMenu`'s "Settings" item. The modal SHALL display a left-hand navigation listing three sections — **Account**, **App**, and **Advanced** — and SHALL render the content of the currently selected section. The **Account** section SHALL be selected by default when the modal opens. The modal SHALL be dismissible via a close control.

#### Scenario: Modal opens to Account section by default

- **WHEN** the user opens the SettingsModal
- **THEN** the Account section is displayed and highlighted as active in the left nav

#### Scenario: Switching sections

- **WHEN** the user clicks "App" or "Advanced" in the left nav
- **THEN** the corresponding section's content replaces the displayed content
- **AND** that nav item is highlighted as active

#### Scenario: Closing the modal

- **WHEN** the user activates the close control
- **THEN** the SettingsModal closes

### Requirement: Account section profile display

The Account section SHALL display the authenticated user's avatar, full name, and email as read-only information, sourced from the user's Kinde profile (`useKindeAuth()`). When no avatar `picture` is available, an initials-based fallback SHALL be shown, consistent with `ProfileMenu`.

#### Scenario: Profile information is displayed read-only

- **WHEN** the Account section is displayed
- **THEN** the user's avatar (or initials fallback), full name, and email are shown
- **AND** none of these fields are editable

### Requirement: Account section log out

The Account section SHALL include a "Log out" action that calls `logout()` from `useKindeAuth()`.

#### Scenario: Logging out from the Account section

- **WHEN** the user clicks "Log out" in the Account section
- **THEN** `logout()` is called and the session ends

### Requirement: Delete account confirmation

The Account section SHALL include a "Delete account" action that, when initiated, requires the user to type their account email address into a confirmation field before the deletion can be confirmed. The confirm action SHALL remain disabled until the typed value matches the authenticated user's email, compared case-insensitively with surrounding whitespace trimmed.

#### Scenario: Delete action disabled until email matches

- **WHEN** the user has initiated account deletion and typed an email that does not match their account email
- **THEN** the confirm delete action remains disabled

#### Scenario: Delete action enabled when email matches

- **WHEN** the user types their account email, ignoring case and surrounding whitespace, into the confirmation field
- **THEN** the confirm delete action becomes enabled

### Requirement: Account deletion execution and outcomes

Confirming account deletion SHALL call `DELETE /me`. On success, the system SHALL call `logout()` and navigate to `/`. On failure, the system SHALL display a toast notification describing the error and SHALL remain on the delete confirmation view with the typed email preserved.

#### Scenario: Successful account deletion

- **WHEN** the user confirms account deletion and `DELETE /me` succeeds
- **THEN** the user is logged out
- **AND** the user is navigated to `/`

#### Scenario: Failed account deletion

- **WHEN** the user confirms account deletion and `DELETE /me` fails
- **THEN** a toast notification describing the error is shown
- **AND** the user remains on the delete confirmation view with their typed email preserved

### Requirement: App section theme display

The App section SHALL display the application's active theme using `useTheme()`, shown as the selected option. Only the "Default" (warm) theme SHALL be shown.

#### Scenario: Active theme is displayed

- **WHEN** the App section is displayed
- **THEN** the "Default" theme option is shown as selected, reflecting the value returned by `useTheme()`

### Requirement: Advanced section AI features toggle

The Advanced section SHALL display an "AI features" toggle reflecting the authenticated user's `vision_enabled` value from `GET /me`. The toggle SHALL be rendered in a disabled (non-interactive) state.

#### Scenario: Toggle reflects enabled state

- **WHEN** `vision_enabled` is `true` for the authenticated user
- **THEN** the AI features toggle is displayed as on and disabled

#### Scenario: Toggle reflects disabled state

- **WHEN** `vision_enabled` is `false` for the authenticated user
- **THEN** the AI features toggle is displayed as off and disabled
