## ADDED Requirements

### Requirement: KindeProvider initialised

The app SHALL wrap the entire React tree with `KindeProvider` from `@kinde-oss/kinde-auth-react` in `main.tsx`, configured via `VITE_KINDE_CLIENT_ID`, `VITE_KINDE_ISSUER_URL`, `VITE_KINDE_REDIRECT_URL`, and `VITE_KINDE_LOGOUT_REDIRECT_URL` environment variables. `KindeProvider` MUST be the outermost wrapper, above `<BrowserRouter>`.

#### Scenario: Provider initialises on app load

- **WHEN** the app loads with all required Kinde env vars set
- **THEN** `KindeProvider` initialises without errors and auth state is available throughout the component tree

#### Scenario: Missing env vars are surfaced

- **WHEN** a required Kinde env var is missing or empty
- **THEN** the console emits a warning identifying which variable is unset

### Requirement: Login page

The app SHALL expose a `/login` route rendering a page with a single sign-in button. Clicking the button SHALL call `login()` from `useKindeAuth()`, redirecting the user to Kinde's hosted login page. The login page SHALL NOT be accessible to already-authenticated users — they SHALL be redirected to `/`. The login page SHALL display an error message when one is passed via React Router location state (e.g. after a failed callback).

#### Scenario: Unauthenticated user sees login page

- **WHEN** an unauthenticated user navigates to `/login`
- **THEN** the login page renders with a single sign-in button

#### Scenario: Sign-in button initiates Kinde login

- **WHEN** the user clicks the sign-in button
- **THEN** the browser redirects to Kinde's hosted login flow

#### Scenario: Authenticated user is redirected away from login

- **WHEN** an already-authenticated user navigates to `/login`
- **THEN** they are redirected to `/`

#### Scenario: Error message is shown when passed via location state

- **WHEN** the user is redirected to `/login` with an error message in React Router location state
- **THEN** the error message is displayed on the login page above the sign-in button

### Requirement: Callback route

The app SHALL expose a `/callback` route that handles the OAuth redirect from Kinde after a successful login. The route SHALL render a loading state while the SDK processes the code exchange. Once the exchange is complete and the user is authenticated, the app SHALL navigate to `/`.

#### Scenario: Successful OAuth callback completes login

- **WHEN** Kinde redirects the user to `/callback` with a valid authorisation code
- **THEN** the SDK processes the code exchange
- **AND** the user is navigated to `/` as an authenticated user

#### Scenario: Callback shows loading state during exchange

- **WHEN** the user lands on `/callback` and the exchange is in progress
- **THEN** a loading indicator is displayed and no content-bearing UI is rendered

#### Scenario: Callback error redirects to login with message

- **WHEN** Kinde returns an error on the `/callback` route (e.g. access denied, invalid state)
- **THEN** the user is redirected to `/login`
- **AND** an error message is displayed on the login page informing them that sign-in failed and they should try again

### Requirement: AuthGuard component

The app SHALL provide an `AuthGuard` layout component that wraps all protected routes. `AuthGuard` SHALL redirect unauthenticated users to `/login`. While Kinde's auth state is loading, `AuthGuard` SHALL render nothing (or a neutral loading state) to prevent a flash of protected content. Authenticated users SHALL see the rendered child route via `<Outlet />`.

#### Scenario: Unauthenticated user is redirected to login

- **WHEN** an unauthenticated user navigates to a protected route
- **THEN** they are redirected to `/login`

#### Scenario: Authenticated user sees protected content

- **WHEN** an authenticated user navigates to a protected route
- **THEN** the route renders normally

#### Scenario: Loading state prevents premature redirect

- **WHEN** `isLoading` is true on `KindeProvider` initialisation
- **THEN** `AuthGuard` renders nothing and does not redirect

### Requirement: Logout button

The app SHALL render a logout button in the top-right corner of the authenticated layout. Clicking it SHALL call `logout()` from `useKindeAuth()`, ending the Kinde session and redirecting the browser to `VITE_KINDE_LOGOUT_REDIRECT_URL`. The button SHALL only be visible to authenticated users.

#### Scenario: Logout button ends session

- **WHEN** an authenticated user clicks the logout button
- **THEN** `logout()` is called and the browser redirects to the configured post-logout URL

#### Scenario: Logout button is not visible to unauthenticated users

- **WHEN** an unauthenticated user views any page
- **THEN** the logout button is not rendered

### Requirement: Token-attached API client

The app SHALL export an `apiFetch(path, options?)` function from `src/lib/api.ts`. Before sending any request, it SHALL retrieve the current Kinde access token via `getToken()` and attach it as `Authorization: Bearer <token>` in the request headers. The base URL for all requests SHALL be read from `VITE_API_BASE_URL`. The function SHALL be async and return a typed `Response`.

#### Scenario: Authenticated request includes bearer token

- **WHEN** `apiFetch` is called while the user is authenticated
- **THEN** the outgoing request includes `Authorization: Bearer <token>` with the current Kinde access token

#### Scenario: Request uses configured base URL

- **WHEN** `apiFetch('/images')` is called
- **THEN** the request is sent to `${VITE_API_BASE_URL}/images`
