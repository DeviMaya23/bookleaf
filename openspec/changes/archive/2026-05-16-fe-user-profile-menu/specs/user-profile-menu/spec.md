## ADDED Requirements

### Requirement: ProfileMenu component

The app SHALL render a `ProfileMenu` component in the footer of the `FolderSidebar`. The component SHALL display the user's avatar (from `picture`) and full name (from `givenName` + `familyName`) as a trigger button. When `picture` is unavailable the component SHALL render an avatar fallback using the user's initials. Clicking the trigger SHALL open a dropdown menu containing a single **Sign out** item. Clicking **Sign out** SHALL call `logout()` from `useKindeAuth()`.

#### Scenario: Profile trigger renders with avatar

- **WHEN** the authenticated user has a `picture` URL in their Kinde profile
- **THEN** the trigger renders an avatar image alongside the user's full name

#### Scenario: Profile trigger renders initials fallback

- **WHEN** the authenticated user has no `picture` URL
- **THEN** the trigger renders an avatar with the user's initials derived from `givenName` and `familyName`

#### Scenario: Dropdown opens on trigger click

- **WHEN** the user clicks the `ProfileMenu` trigger
- **THEN** a dropdown appears containing a Sign out item

#### Scenario: Sign out calls logout

- **WHEN** the user clicks Sign out in the dropdown
- **THEN** `logout()` is called and the session ends
