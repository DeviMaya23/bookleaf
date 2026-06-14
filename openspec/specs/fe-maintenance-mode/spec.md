## Purpose

Allow the frontend to detect when the backend is in maintenance mode via response headers, support an operator bypass token stored in `localStorage`, and show a dedicated `MaintenancePage` in place of the app shell while maintenance is active.

## Requirements

### Requirement: apiFetch attaches the stored bypass token
`apiFetch` SHALL, if `localStorage` key `bookleaf-maintenance-bypass` holds a non-empty value, include it as the `X-Bookleaf-Bypass` request header on every request it makes.

#### Scenario: Bypass token present in localStorage
- **WHEN** `localStorage.getItem('bookleaf-maintenance-bypass')` returns a non-empty string
- **THEN** the request sent by `apiFetch` includes header `X-Bookleaf-Bypass` set to that value

#### Scenario: No bypass token stored
- **WHEN** `localStorage.getItem('bookleaf-maintenance-bypass')` returns `null` or an empty string
- **THEN** the request sent by `apiFetch` does not include an `X-Bookleaf-Bypass` header

### Requirement: apiFetch updates shared maintenance state from response header
`apiFetch` SHALL, after receiving a response, set the shared maintenance-active state to `true` if the response includes header `X-Bookleaf-Maintenance: true`, and set it to `false` for any response that does not include that header.

#### Scenario: Response signals maintenance is active
- **WHEN** a response to an `apiFetch` call includes header `X-Bookleaf-Maintenance: true`
- **THEN** the shared maintenance-active state becomes `true`

#### Scenario: Response does not signal maintenance
- **WHEN** a response to an `apiFetch` call does not include the `X-Bookleaf-Maintenance` header
- **THEN** the shared maintenance-active state becomes `false`

### Requirement: /app shell shows MaintenancePage while maintenance is active
`AppLayout` SHALL render `MaintenancePage` instead of the normal application shell whenever the shared maintenance-active state is `true`. This applies to all `/app/*` routes that render `AppLayout` (`index`, `unsorted`, `trash`, `folders/:folderId`). When the shared state transitions back to `false`, `AppLayout` SHALL render the normal shell again without requiring a page reload.

#### Scenario: Maintenance page shown while active
- **WHEN** the shared maintenance-active state is `true` and the user is on any `/app/*` route
- **THEN** `MaintenancePage` is rendered instead of the normal application shell

#### Scenario: Normal shell resumes after maintenance ends
- **WHEN** the shared maintenance-active state transitions from `true` to `false` (following a subsequent `apiFetch` response without the maintenance header)
- **THEN** `AppLayout` renders the normal application shell again without a page reload
