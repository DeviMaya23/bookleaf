## Why

Bookleaf currently has no public-facing entry point — unauthenticated visitors land on a bare "Sign in" button at `/login`, and the authenticated app lives at `/`. There's no page that introduces what Bookleaf is before asking someone to sign in. This change adds a real landing page and, since the app currently occupies the root path, moves the authenticated app to its own namespace (`/app`) so `/` can become the public landing page.

## What Changes

- **BREAKING**: Authenticated app routes move from `/`, `/unsorted`, `/trash`, `/folders/:folderId` to `/app`, `/app/unsorted`, `/app/trash`, `/app/folders/:folderId`
- **BREAKING**: The `/login` route is removed; its behavior (sign-in trigger) folds into the new `/` landing page
- `AuthGuard` redirects unauthenticated users to `/` instead of `/login`
- `CallbackPage` navigates to `/app` on successful login instead of `/`
- Sidebar system entries (All / Unsorted / Trash) navigate to `/app`, `/app/unsorted`, `/app/trash`
- New public landing page at `/`, built per the design reference (`Bookleaf Landing.html`):
  - Nav: Bookleaf wordmark + "Sign in" button
  - Hero: overline, headline, tagline, "Get started" button, and an app screenshot (4:3, `object-fit: cover`)
  - Features section (3 items)
  - Footer
  - Both "Sign in" and "Get started" call Kinde's `login()` — Kinde's hosted page handles both sign-in and sign-up
  - Non-scrollable (`100dvh`) layout with reduced font sizes relative to the design reference defaults
- Mobile/narrow-viewport responsiveness is explicitly deferred

## Capabilities

### New Capabilities
- `fe-landing-page`: Public landing page at `/` — nav, hero with screenshot, features section, footer, sign-in/get-started CTAs

### Modified Capabilities
- `fe-kinde-auth`: The "Login page" requirement (`/login` route) is removed. `AuthGuard` redirect target changes from `/login` to `/`. `CallbackPage` redirect target changes from `/` to `/app`.
- `fe-sidebar-nav`: System entry routes change from `/`, `/unsorted`, `/trash` to `/app`, `/app/unsorted`, `/app/trash`.

## Impact

- `frontend/src/App.tsx` — route definitions
- `frontend/src/pages/LoginPage.tsx` — removed
- `frontend/src/pages/LandingPage.tsx` (new) — public landing page
- `frontend/src/pages/CallbackPage.tsx` — redirect target
- `frontend/src/features/auth/components/AuthGuard.tsx` — redirect target
- `frontend/src/features/folder-sidebar/components/FolderSidebar.tsx` — nav targets
- `frontend/src/app-shell/useAppView.ts` — path matching
- `frontend/src/index.css` — possible new token for the features section background
- `frontend/src/assets/landing-hero.png` (new asset, provided by user)
- Affected tests: `AppLayout.test.tsx`, `useAppView.test.tsx`, `CallbackPage.test.tsx`, `LoginPage.test.tsx` (removed/replaced), `AuthGuard.test.tsx`, `FolderSidebar.test.tsx`
