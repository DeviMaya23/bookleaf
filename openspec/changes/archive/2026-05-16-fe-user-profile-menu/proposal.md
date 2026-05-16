## Why

The app currently shows only a bare logout button with no user identity visible. Surfacing the user's name, avatar, and email in a profile menu gives the authenticated layout a proper identity anchor and consolidates sign-out into a purposeful action rather than a dangling button.

## What Changes

- The `LogoutButton` component is removed and replaced by a `ProfileMenu` component that renders the user's avatar and name. Clicking it opens a dropdown with a **Sign out** action.
- Profile data (name, email, picture) is sourced directly from `useKindeAuth().getUserProfile()` on the frontend — no backend changes required.

## Capabilities

### New Capabilities

- `user-profile-menu`: FE profile component rendered in the app shell — shows avatar + full name as a trigger, opens a dropdown with a Sign out action.

### Modified Capabilities

- `fe-kinde-auth`: Logout button requirement is superseded by the profile menu; the standalone logout button is removed.

## Impact

- **Frontend only**: `LogoutButton` component removed; new `ProfileMenu` component added to `FolderSidebar` footer.
- No backend changes. No breaking changes.
