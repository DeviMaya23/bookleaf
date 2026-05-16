## Context

The app has a standalone `LogoutButton` with no user identity context. Profile data (name, email, avatar) is available on the frontend via `useKindeAuth().getUserProfile()`, which reads from the in-memory ID token — no extra network call needed.

## Goals / Non-Goals

**Goals:**
- Replace `LogoutButton` with a `ProfileMenu` component showing avatar + name and a Sign out dropdown.

**Non-Goals:**
- Extending the `/me` endpoint — profile fields from Kinde are sufficient for display and available directly on the FE.
- A Settings page or any navigation beyond Sign out.
- Storing profile fields in the DB.

## Decisions

### FE: `ProfileMenu` reads profile data from `useKindeAuth().getUserProfile()` directly

`getUserProfile()` reads from the cached ID token — synchronous-equivalent, always available once authenticated, no additional fetch. This keeps the component self-contained with no dependency on a `/me` round-trip.

### FE: `ProfileMenu` placed in `FolderSidebar` footer, not in `AppLayout` header

The mockup shows the profile trigger at the bottom of the sidebar. `FolderSidebar` already owns the left panel; placing `ProfileMenu` there keeps it self-contained. `AppLayout` does not need changes.

## Risks / Trade-offs

- **Stale profile data**: Claims reflect the token at issue time. For display purposes this is acceptable; a forced re-login refreshes them.
- **Missing picture**: If the user has no avatar set in Kinde, `picture` will be null. The component falls back to initials derived from `givenName` + `familyName`.
