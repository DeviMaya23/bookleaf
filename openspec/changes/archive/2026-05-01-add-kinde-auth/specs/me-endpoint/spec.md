## ADDED Requirements

### Requirement: GET /me Endpoint

The system SHALL expose a `GET /me` endpoint that returns the authenticated user's Kinde ID and `vision_enabled` flag. The endpoint SHALL be in the protected route group and require a valid JWT.

Response body (200):
```json
{ "id": "kp_abc123", "vision_enabled": false }
```

- `id` — the Kinde user ID from the authenticated user's DB record
- `vision_enabled` — boolean from the `users` table

#### Scenario: Authenticated user retrieves their profile

- **WHEN** an authenticated request is made to `GET /me`
- **THEN** the response is `200 OK`
- **AND** the body contains the user's Kinde ID and `vision_enabled` value

#### Scenario: Unauthenticated request is rejected

- **WHEN** a request is made to `GET /me` without a valid Bearer token
- **THEN** the response is `401 Unauthorized`
- **AND** no user data is returned
