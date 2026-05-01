## ADDED Requirements

### Requirement: JWT Validation Middleware

The system SHALL provide an Echo middleware that validates Kinde-issued Bearer tokens on protected routes using Kinde's public JWKS endpoint.

- Validates the token signature, expiry, issuer (`KINDE_ISSUER_URL`), and audience (`KINDE_AUDIENCE`)
- JWKS keys SHALL be cached with TTL to avoid fetching on every request
- Returns `401 Unauthorized` if the `Authorization` header is missing, malformed, or contains an invalid token
- On success, extracts the `sub` claim and passes it downstream via Echo context

#### Scenario: Valid token grants access

- **WHEN** a request arrives with a valid Kinde Bearer token in the `Authorization` header
- **THEN** the middleware calls the next handler
- **AND** the Kinde user ID from the `sub` claim is set on the Echo context

#### Scenario: Missing token is rejected

- **WHEN** a request arrives with no `Authorization` header
- **THEN** the middleware returns `401 Unauthorized`
- **AND** the handler is not called

#### Scenario: Invalid token is rejected

- **WHEN** a request arrives with an expired or malformed Bearer token
- **THEN** the middleware returns `401 Unauthorized`
- **AND** the handler is not called

### Requirement: User Auto-Provisioning

The system SHALL automatically create a `User` record in the database the first time a valid JWT is seen for a Kinde user ID not already present in the `users` table.

- Provisioning SHALL use `INSERT ... ON CONFLICT DO NOTHING` to be safe under concurrent requests
- After provisioning (or confirming the user exists), the user ID SHALL be available on the Echo context
- Provisioning failure (DB error) SHALL return `500 Internal Server Error` and not call the handler

#### Scenario: New user is provisioned on first request

- **WHEN** a request arrives with a valid JWT for a Kinde user ID not in the database
- **THEN** a new `User` record is created with `id` set to the Kinde user ID and `vision_enabled` set to `false`
- **AND** the request proceeds to the handler

#### Scenario: Existing user is not duplicated

- **WHEN** a request arrives with a valid JWT for a Kinde user ID already in the database
- **THEN** no new `User` record is created
- **AND** the request proceeds to the handler

### Requirement: Auth Context

The system SHALL expose the authenticated Kinde user ID on the Echo context using a typed constant key so handlers can retrieve it without string casting.

#### Scenario: Handler retrieves user ID from context

- **WHEN** a handler runs after the auth middleware
- **THEN** the Kinde user ID can be retrieved from the Echo context via the typed constant key
- **AND** the value matches the `sub` claim from the JWT
