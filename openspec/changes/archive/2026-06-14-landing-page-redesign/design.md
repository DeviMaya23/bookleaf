## Context

Today, `App.tsx` has two top-level routes outside `AuthGuard` (`/login`, `/callback`) and a group of `AuthGuard`-protected routes at `/`, `/unsorted`, `/trash`, `/folders/:folderId`. `/login` is a bare "Sign in" button (`LoginPage.tsx`) that redirects authenticated users to `/` and displays an error passed via location state after a failed callback. `AuthGuard` redirects unauthenticated users to `/login`; `CallbackPage` redirects successful logins to `/`. `FolderSidebar` and `useAppView` reference `/`, `/unsorted`, `/trash`, `/folders/:id` directly.

This change repurposes `/` as a public landing page and moves the authenticated app to `/app`, per the proposal.

## Goals / Non-Goals

**Goals:**
- `/` renders the new public landing page for unauthenticated visitors
- Authenticated users are routed to `/app` (and bounced there if they land on `/`)
- All existing auth-flow guarantees (sign-in trigger, error display after failed callback, authenticated-redirect) are preserved, just relocated to the landing page
- Landing page matches the design reference's structure (nav / hero / features / footer) in a non-scrollable viewport, using existing warm-theme tokens wherever they already match

**Non-Goals:**
- Mobile/narrow-viewport layout (deferred)
- Final copy — placeholders from the Claude Design handover are used as-is, to be hand-edited later
- Backend or data-model changes (none needed)

## Decisions

### 1. Nest protected routes under `/app` as a parent route

```tsx
<Routes>
  <Route path="/" element={<LandingPage />} />
  <Route path="/callback" element={<CallbackPage />} />
  <Route path="/app" element={<AuthGuard />}>
    <Route index element={<AppLayout />} />
    <Route path="unsorted" element={<AppLayout />} />
    <Route path="trash" element={<AppLayout />} />
    <Route path="folders/:folderId" element={<AppLayout />} />
  </Route>
  <Route path="*" element={<Navigate to="/" replace />} />
</Routes>
```

Alternative considered: flatten each path (`/app`, `/app/unsorted`, ...) as siblings, each individually wrapped. Rejected — nesting under one `/app` parent keeps a single `AuthGuard` wrapper (matching the current structure) and is the idiomatic React Router pattern for a route group with a shared guard.

### 2. `/login` is removed; `LandingPage` absorbs its responsibilities

`LoginPage.tsx` and its test are deleted. `LandingPage.tsx` takes over all three behaviors `fe-kinde-auth`'s "Login page" requirement previously covered:
- Sign-in trigger (both "Sign in" in the nav and "Get started" in the hero call `login()`)
- Redirect authenticated users away (to `/app` instead of `/`)
- Display an error message passed via `location.state.error` (after a failed `/callback`)

Alternative considered: keep `/login` as a redirect alias to `/`. Rejected — nothing references `/login` externally (Kinde's redirect URLs point at `/callback`), so it would be dead weight.

### 3. Legacy bookmarked paths redirect to the landing page, not the app

Because `/`, `/unsorted`, `/trash`, `/folders/:id` no longer match any route, the catch-all sends them to `/` — including for already-authenticated users (who'd then see the landing page briefly before... actually they would just see the landing page, since `/` doesn't auto-forward unauthenticated vs authenticated except via `LandingPage`'s own check, which *does* redirect authenticated users to `/app`). Net effect: an authenticated user with `/unsorted` bookmarked hits `*` → `/` → (authenticated) → `/app`. One extra hop, but lands in the right place.

This is acceptable for a low-traffic personal project; no legacy redirect shims (`/unsorted` → `/app/unsorted`) are being added. Can revisit if it becomes annoying.

### 4. Error banner on the landing page

`CallbackPage`'s error path changes from `navigate('/login', { state: { error } })` to `navigate('/', { state: { error } })`. `LandingPage` reads `location.state.error` and renders it as a small inline banner. Since an error means the user is *not* authenticated, this doesn't conflict with the authenticated-redirect check. Exact placement/styling (e.g. a thin banner above the nav) is a visual detail decided during implementation.

### 5. Feature strip: fixed at 3 items, array + map

`FEATURES` is a local array of `{ num, title, body }` (placeholder copy taken verbatim from the Claude Design handover), rendered via `.map()` into a 3-column grid. The count is fixed at 3 for layout reasons (fits the non-scroll viewport at the design's padding/sizing); using a map instead of inline JSX costs nothing and means changing the count later is a one-line + copy change, not a structural one.

### 6. Features-section background color: scoped, not a new global token

The design's `FEAT_BG` value (`#F0EBE3`) doesn't match any existing `--color-*` token. Rather than adding a new global warm-theme token for a color used in exactly one section of one page, it's defined as a locally-scoped value within `LandingPage` (e.g. a CSS variable or class scoped to the component). Alternative considered: add `--color-landing-feature-bg` to `index.css`'s theme tokens — rejected to avoid growing the shared design-system surface for a single-use color; can be promoted to a shared token later if it turns out to be reused.

### 7. Hero screenshot

`frontend/src/assets/landing-hero.png` (provided by the user, 4:3) is rendered with `aspect-[4/3] object-cover` inside the hero's right-hand panel, with the existing border/shadow treatment from the design reference. The current unrelated `hero.png` (unused Vite-era placeholder) is deleted as part of this change.

### 8. Non-scroll layout structure

`LandingPage` uses a `h-dvh overflow-hidden` wrapper with `grid-template-rows: auto 1fr auto auto` (nav / hero / features / footer), mirroring `Bookleaf Landing.html`. Font sizes are reduced from the design's defaults (e.g. 60px headline) — exact values are tuned visually during implementation once the real screenshot is in place, since the screenshot's footprint affects how much room the hero text has.

## Risks / Trade-offs

- **[Risk]** Bookmarked/shared links to old routes (`/`, `/unsorted`, `/trash`, `/folders/:id`) now redirect through the landing page instead of going straight to the app → **Mitigation**: acceptable for a personal project with few users; add legacy redirect shims later if it becomes a real annoyance.
- **[Risk]** Non-scrollable layout combined with a fixed 4:3 screenshot may not look balanced at all common desktop widths → **Mitigation**: verify visually at a few breakpoints during implementation; mobile is explicitly out of scope for this change.
- **[Risk]** Removing `/login` changes scenarios in the `fe-kinde-auth` spec → **Mitigation**: scenarios are migrated (not dropped) — sign-in trigger, authenticated-redirect, and error-display all still exist, relocated to `fe-landing-page`.

## Migration Plan

1. Update `App.tsx` to the new route tree (Decision 1).
2. Update `AuthGuard` (`/login` → `/`) and `CallbackPage` (`/` → `/app`, error target `/login` → `/`).
3. Update `FolderSidebar` nav calls and `useAppView` path checks for the `/app` prefix.
4. Delete `LoginPage.tsx` + its test; add `LandingPage.tsx` + its test.
5. Update affected tests: `AppLayout.test.tsx`, `useAppView.test.tsx`, `CallbackPage.test.tsx`, `AuthGuard.test.tsx`, `FolderSidebar.test.tsx`.
6. Add `landing-hero.png`; remove unused `hero.png`.
7. Manual verification: unauthenticated `/` shows landing; "Sign in"/"Get started" trigger Kinde; successful login lands on `/app`; authenticated visit to `/` bounces to `/app`; unauthenticated visit to a protected path bounces to `/`.

No backend changes or data migrations — rollback is a straight revert of the branch.

## Open Questions

- Should legacy path redirects (`/unsorted` → `/app/unsorted`, etc.) be added now or deferred? Leaning deferred (Decision 3).
- Exact visual placement of the error banner on the landing page — left to implementation.
- Final copy (headline, tagline, feature blurbs) — placeholders from the design handover used until the user edits them manually.
