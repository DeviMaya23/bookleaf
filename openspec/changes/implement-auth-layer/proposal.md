## Why

The frontend scaffolding is in place but has no authentication. Without an auth layer, all routes are publicly accessible and API calls cannot be attributed to a logged-in user. Wiring up Kinde's React SDK now unlocks protected routes and provides the bearer token mechanism needed for all future backend API calls.

## What Changes

- Install and configure `KindeProvider` from `@kinde-oss/kinde-auth-react` (already in `package.json`)
- Add a `/login` page with a single sign-in button that redirects to Kinde's hosted login flow
- Add a `/callback` route that handles the OAuth redirect back from Kinde after a successful login
- Add an `AuthGuard` component that redirects unauthenticated users to `/login` and authenticated users to the app home
- Add an API client utility that attaches the Kinde access token as a `Bearer` token in the `Authorization` header for all backend requests
- Add a logout button in the top-right of the app that ends the Kinde session and redirects to the post-logout URL

## Capabilities

### New Capabilities
- `fe-kinde-auth`: Frontend authentication via Kinde React SDK — provider setup, login page, callback route, AuthGuard component, token-attached API client, and logout button

### Modified Capabilities

## Impact

- `frontend/src/main.tsx` — wrap app in `KindeProvider`
- `frontend/src/App.tsx` — add `/login` and `/callback` routes, wrap protected routes with `AuthGuard`
- New files: `src/pages/LoginPage.tsx`, `src/pages/CallbackPage.tsx`, `src/components/AuthGuard.tsx`, `src/components/LogoutButton.tsx`, `src/lib/api.ts`
- `frontend/.env` / `.env.example` — `VITE_KINDE_CLIENT_ID`, `VITE_KINDE_ISSUER_URL`, `VITE_KINDE_REDIRECT_URL`, `VITE_KINDE_LOGOUT_REDIRECT_URL` must be populated
