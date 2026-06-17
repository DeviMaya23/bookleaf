## ADDED Requirements

### Requirement: GET /me Endpoint

The system SHALL expose a `GET /me` endpoint that returns the authenticated user's Kinde ID, `vision_enabled` flag, and `folder_icons_enabled` flag. The endpoint SHALL be in the protected route group and require a valid JWT.

Response body (200):
```json
{ "id": "kp_abc123", "vision_enabled": false, "folder_icons_enabled": true }
```

- `id` — the Kinde user ID from the authenticated user's DB record
- `vision_enabled` — boolean from the `users` table
- `folder_icons_enabled` — boolean from the `users` table

#### Scenario: Authenticated user retrieves their profile

- **WHEN** an authenticated request is made to `GET /me`
- **THEN** the response is `200 OK`
- **AND** the body contains the user's Kinde ID, `vision_enabled`, and `folder_icons_enabled` values

#### Scenario: Unauthenticated request is rejected

- **WHEN** a request is made to `GET /me` without a valid Bearer token
- **THEN** the response is `401 Unauthorized`
- **AND** no user data is returned

### Requirement: PATCH /me Endpoint

The system SHALL expose a `PATCH /me` endpoint that updates the authenticated user's `vision_enabled` and/or `folder_icons_enabled` flags. The endpoint SHALL be in the protected route group and require a valid JWT. Only fields present in the request body are modified — omitted fields are left unchanged.

Request body (at least one of the two fields required):
```json
{ "vision_enabled": true, "folder_icons_enabled": false }
```

Response body (200) — same shape as `GET /me`:
```json
{ "id": "kp_abc123", "vision_enabled": true, "folder_icons_enabled": false }
```

- `vision_enabled` and `folder_icons_enabled` are the only writable fields. The request body SHALL NOT permit updating any other user column (`id`, `pending_kinde_deletion`, timestamps).
- At least one of `vision_enabled` or `folder_icons_enabled` is REQUIRED in the request body; an empty body is a client error.
- Each field present MUST be a boolean; any other JSON type for a present field is a client error.

#### Scenario: Authenticated user enables vision labelling

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "vision_enabled": true }`
- **THEN** the response is `200 OK`
- **AND** the body is `{ "id": "<user id>", "vision_enabled": true, "folder_icons_enabled": <unchanged> }`
- **AND** the user's `vision_enabled` column is persisted as `true`

#### Scenario: Authenticated user disables vision labelling

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "vision_enabled": false }`
- **THEN** the response is `200 OK`
- **AND** the body is `{ "id": "<user id>", "vision_enabled": false, "folder_icons_enabled": <unchanged> }`
- **AND** the user's `vision_enabled` column is persisted as `false`

#### Scenario: Authenticated user disables folder icons

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "folder_icons_enabled": false }`
- **THEN** the response is `200 OK`
- **AND** the body is `{ "id": "<user id>", "vision_enabled": <unchanged>, "folder_icons_enabled": false }`
- **AND** the user's `folder_icons_enabled` column is persisted as `false`
- **AND** the user's `vision_enabled` column is unchanged

#### Scenario: Authenticated user re-enables folder icons

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "folder_icons_enabled": true }`
- **THEN** the response is `200 OK`
- **AND** the user's `folder_icons_enabled` column is persisted as `true`

#### Scenario: Missing both fields is rejected

- **WHEN** an authenticated request is made to `PATCH /me` with an empty body `{}`
- **THEN** the response is `400 Bad Request`
- **AND** neither the user's `vision_enabled` nor `folder_icons_enabled` column is changed

#### Scenario: Non-boolean vision_enabled is rejected

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "vision_enabled": "yes" }`
- **THEN** the response is `400 Bad Request`
- **AND** the user's `vision_enabled` column is unchanged

#### Scenario: Non-boolean folder_icons_enabled is rejected

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "folder_icons_enabled": "yes" }`
- **THEN** the response is `400 Bad Request`
- **AND** the user's `folder_icons_enabled` column is unchanged

#### Scenario: Unauthenticated request is rejected

- **WHEN** a request is made to `PATCH /me` without a valid Bearer token
- **THEN** the response is `401 Unauthorized`
- **AND** no user data is changed
