## MODIFIED Requirements

### Requirement: Callback route

The app SHALL expose a `/callback` route that handles the OAuth redirect from Kinde after a successful login. The route SHALL render a loading state while the SDK processes the code exchange. Once the exchange is complete and the user is authenticated, the app SHALL navigate to `/app`.

#### Scenario: Successful OAuth callback completes login

- **WHEN** Kinde redirects the user to `/callback` with a valid authorisation code
- **THEN** the SDK processes the code exchange
- **AND** the user is navigated to `/app` as an authenticated user

#### Scenario: Callback shows loading state during exchange

- **WHEN** the user lands on `/callback` and the exchange is in progress
- **THEN** a loading indicator is displayed and no content-bearing UI is rendered

#### Scenario: Callback error redirects to the landing page with message

- **WHEN** Kinde returns an error on the `/callback` route (e.g. access denied, invalid state)
- **THEN** the user is redirected to `/`
- **AND** an error message is displayed on the landing page informing them that sign-in failed and they should try again

### Requirement: AuthGuard component

The app SHALL provide an `AuthGuard` layout component that wraps all protected routes. `AuthGuard` SHALL redirect unauthenticated users to `/`. While Kinde's auth state is loading, `AuthGuard` SHALL render nothing (or a neutral loading state) to prevent a flash of protected content. Authenticated users SHALL see the rendered child route via `<Outlet />`.

#### Scenario: Unauthenticated user is redirected to the landing page

- **WHEN** an unauthenticated user navigates to a protected route
- **THEN** they are redirected to `/`

#### Scenario: Authenticated user sees protected content

- **WHEN** an authenticated user navigates to a protected route
- **THEN** the route renders normally

#### Scenario: Loading state prevents premature redirect

- **WHEN** `isLoading` is true on `KindeProvider` initialisation
- **THEN** `AuthGuard` renders nothing and does not redirect

## REMOVED Requirements

### Requirement: Login page

**Reason**: The `/login` route is removed. Its responsibilities — a sign-in trigger, redirecting already-authenticated users away, and displaying an error passed via location state — are now provided by the public landing page at `/` (see `fe-landing-page`).

**Migration**: No user-facing migration needed. Any internal links or redirects that pointed at `/login` now point at `/`.
