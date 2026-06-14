## ADDED Requirements

### Requirement: PATCH /me Endpoint

The system SHALL expose a `PATCH /me` endpoint that updates the authenticated user's `vision_enabled` flag. The endpoint SHALL be in the protected route group and require a valid JWT.

Request body:
```json
{ "vision_enabled": true }
```

Response body (200) — same shape as `GET /me`:
```json
{ "id": "kp_abc123", "vision_enabled": true }
```

- `vision_enabled` is the only writable field. The request body SHALL NOT permit updating any other user column (`id`, `pending_kinde_deletion`, timestamps).
- The field is REQUIRED in the request body; omitting it is a client error.
- The field MUST be a boolean; any other JSON type is a client error.

#### Scenario: Authenticated user enables vision labelling

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "vision_enabled": true }`
- **THEN** the response is `200 OK`
- **AND** the body is `{ "id": "<user id>", "vision_enabled": true }`
- **AND** the user's `vision_enabled` column is persisted as `true`

#### Scenario: Authenticated user disables vision labelling

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "vision_enabled": false }`
- **THEN** the response is `200 OK`
- **AND** the body is `{ "id": "<user id>", "vision_enabled": false }`
- **AND** the user's `vision_enabled` column is persisted as `false`

#### Scenario: Missing vision_enabled field is rejected

- **WHEN** an authenticated request is made to `PATCH /me` with an empty body `{}`
- **THEN** the response is `400 Bad Request`
- **AND** the user's `vision_enabled` column is unchanged

#### Scenario: Non-boolean vision_enabled is rejected

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "vision_enabled": "yes" }`
- **THEN** the response is `400 Bad Request`
- **AND** the user's `vision_enabled` column is unchanged

#### Scenario: Unauthenticated request is rejected

- **WHEN** a request is made to `PATCH /me` without a valid Bearer token
- **THEN** the response is `401 Unauthorized`
- **AND** no user data is changed
