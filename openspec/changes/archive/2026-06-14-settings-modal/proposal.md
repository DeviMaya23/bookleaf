## Why

Users have no way to view their account details, manage app preferences, or delete their account from the UI. The backend account-deletion endpoint (`DELETE /me`) already exists and is merged, but nothing in the frontend calls it. A Settings modal, reachable from the existing profile menu, closes this gap and gives the account-deletion feature a home.

## What Changes

- Add a "Settings" item to the existing `ProfileMenu` dropdown (alongside "Sign out"), which opens a new `SettingsModal`.
- `SettingsModal` has a left-nav layout with three sections: **Account**, **App**, **Advanced**.
- **Account** section:
  - Read-only profile display (avatar, name, email) sourced from `useKindeAuth()`.
  - "Log out" button (in addition to the existing ProfileMenu sign-out entry).
  - "Delete account" flow: user must type their account email to enable the delete action, then confirms to call `DELETE /me`. On success, logs out and navigates to `/`. On error, stays on the confirm screen and shows a toast.
- **App** section:
  - Displays the active theme ("Default / Warm parchment") as a single row, wired to the real `useTheme()` hook from the `fe-theme` capability.
- **Advanced** section:
  - Displays an "AI features" toggle reflecting the real `vision_enabled` value from `GET /me`. Rendered disabled/non-interactive since no `PATCH /me` exists yet to change it.

## Capabilities

### New Capabilities
- `fe-settings-modal`: The Settings modal shell, its Account/App/Advanced sections, and the delete-account confirmation flow.

### Modified Capabilities
- `user-profile-menu`: Add a "Settings" item to the `ProfileMenu` dropdown that opens the `SettingsModal`.

## Impact

- **Frontend**: New `SettingsModal` component and section subcomponents (Account/App/Advanced) under a settings feature directory; `ProfileMenu.tsx` gains a "Settings" menu item and modal-open state.
- **APIs**: Consumes existing `DELETE /me` (account deletion) and `GET /me` (`vision_enabled`) endpoints. No backend changes.
- **Dependencies**: Reuses existing `useKindeAuth()`, `useTheme()` (`fe-theme`), and the `me` query already used by `useVisionSuggestion`.
