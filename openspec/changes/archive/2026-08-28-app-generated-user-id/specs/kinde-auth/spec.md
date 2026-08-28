## MODIFIED Requirements

### Requirement: Auth Context

The system SHALL expose the authenticated user's **internal UUID** on the Echo context using a typed constant key so handlers can retrieve it without string casting.

The value stored in context SHALL be `user.ID.String()` — the app-generated UUID returned by `GetOrProvision` — not the `sub` claim from the JWT. The context value type remains `string`.

#### Scenario: Handler retrieves internal UUID from context

- **WHEN** a handler runs after the auth middleware
- **THEN** the value retrieved from the Echo context via the typed constant key is the user's internal UUID string
- **AND** the value does NOT equal the Kinde subject from the JWT `sub` claim (unless they happen to be equal, which they never are after this change)

#### Scenario: Valid token grants access and stores UUID

- **WHEN** a request arrives with a valid Kinde Bearer token in the `Authorization` header
- **THEN** the middleware calls `GetOrProvision` with the `sub` claim
- **AND** stores `user.ID.String()` (the internal UUID) in the Echo context
- **AND** calls the next handler
