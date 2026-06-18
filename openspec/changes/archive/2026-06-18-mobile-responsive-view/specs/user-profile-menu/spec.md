## MODIFIED Requirements

### Requirement: ProfileMenu component

The app SHALL render a `ProfileMenu` component in the footer of the `FolderSidebar`. The component SHALL display the user's avatar (from `picture`) and full name (from `givenName` + `familyName`) as a trigger button. When `picture` is unavailable the component SHALL render an avatar fallback using the user's initials. At or above the `sm` breakpoint, clicking the trigger SHALL open a dropdown menu containing **Settings** and **Sign out** items; clicking **Settings** SHALL open the `SettingsModal`. Below the `sm` breakpoint, clicking the trigger SHALL open a dropdown menu containing only **Sign out** — the **Settings** item SHALL NOT be rendered. Clicking **Sign out** SHALL call `logout()` from `useKindeAuth()`, regardless of breakpoint.

#### Scenario: Profile trigger renders with avatar

- **WHEN** the authenticated user has a `picture` URL in their Kinde profile
- **THEN** the trigger renders an avatar image alongside the user's full name

#### Scenario: Profile trigger renders initials fallback

- **WHEN** the authenticated user has no `picture` URL
- **THEN** the trigger renders an avatar with the user's initials derived from `givenName` and `familyName`

#### Scenario: Dropdown opens on trigger click at or above the breakpoint

- **WHEN** the user clicks the `ProfileMenu` trigger at or above the `sm` breakpoint
- **THEN** a dropdown appears containing Settings and Sign out items

#### Scenario: Settings opens the settings modal

- **WHEN** the user clicks Settings in the dropdown
- **THEN** the `SettingsModal` opens

#### Scenario: Sign out calls logout

- **WHEN** the user clicks Sign out in the dropdown
- **THEN** `logout()` is called and the session ends

#### Scenario: Dropdown shows only Sign out below the breakpoint

- **WHEN** the user clicks the `ProfileMenu` trigger below the `sm` breakpoint
- **THEN** a dropdown appears containing only the Sign out item
- **AND** the Settings item is not rendered
