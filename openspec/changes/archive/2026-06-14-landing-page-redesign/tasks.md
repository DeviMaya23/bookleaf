## 1. Routing & auth plumbing

- [x] 1.1 Update `App.tsx`: nest protected routes under `/app` (`index`, `unsorted`, `trash`, `folders/:folderId`) wrapped by `AuthGuard`; render `LandingPage` at `/`; catch-all redirects to `/`
- [x] 1.2 Update `AuthGuard.tsx`: redirect unauthenticated users to `/` instead of `/login`
- [x] 1.3 Update `CallbackPage.tsx`: navigate to `/app` on successful login; on callback error, navigate to `/` (with `state.error`) instead of `/login`
- [x] 1.4 Update `useAppView.ts`: match `/app/unsorted` and `/app/trash` instead of `/unsorted` and `/trash`
- [x] 1.5 Update `FolderSidebar.tsx`: change nav calls to `/app`, `/app/unsorted`, `/app/trash`, `/app/folders/:id`

## 2. Landing page

- [x] 2.1 Create `frontend/src/pages/LandingPage.tsx` with the non-scrollable (`h-dvh overflow-hidden`, grid rows nav/hero/features/footer) layout from `Bookleaf Landing.html`, using existing warm-theme tokens for background/foreground/border/muted colors
- [x] 2.2 Implement nav: "Bookleaf" wordmark + "Sign in" button calling `login()`
- [x] 2.3 Implement hero: overline, headline, tagline, "Get started" button calling `login()`, and a screenshot panel (`aspect-[4/3] object-cover`, border/shadow per design reference) referencing `landing-hero.png`
- [x] 2.4 Implement features section: local `FEATURES` array of 3 `{ num, title, body }` items (placeholder copy from the Claude Design handover) rendered in a 3-column grid, with a locally-scoped background color for the section (not a new global theme token)
- [x] 2.5 Implement footer with "Bookleaf" wordmark
- [x] 2.6 Add authenticated-redirect: if `isAuthenticated`, `<Navigate to="/app" replace />`
- [x] 2.7 Add error banner: display `location.state.error` when present
- [x] 2.8 Reduce font sizes from the design reference defaults so the layout fits common desktop viewports without scrolling; verify visually

## 3. Assets & cleanup

- [x] 3.1 Add `frontend/src/assets/landing-hero.png` (4:3 app screenshot, provided by user)
- [x] 3.2 Remove unused `frontend/src/assets/hero.png`
- [x] 3.3 Delete `frontend/src/pages/LoginPage.tsx` and `LoginPage.test.tsx`

## 4. Tests

- [x] 4.1 Create `LandingPage.test.tsx`: unauthenticated user sees nav/hero/features/footer; "Sign in" and "Get started" both call `login()`; authenticated user is redirected to `/app`; error message from location state is displayed
- [x] 4.2 Update `AuthGuard.test.tsx`: unauthenticated redirect target is `/`
- [x] 4.3 Update `CallbackPage.test.tsx`: success navigates to `/app`; error navigates to `/` with `state.error`
- [x] 4.4 Update `useAppView.test.tsx`: route definitions and assertions use `/app`, `/app/unsorted`, `/app/trash`, `/app/folders/:id`
- [x] 4.5 Update `FolderSidebar.test.tsx`: nav assertions use `/app`-prefixed routes (no nav assertions exist in this file; no-op)
- [x] 4.6 Update `AppLayout.test.tsx`: route definitions and `renderApp()` calls use `/app`-prefixed paths

## 5. Final checks

- [x] 5.1 Run `npm run build` and fix any issues
- [x] 5.2 Run `npm run lint` and fix any issues
