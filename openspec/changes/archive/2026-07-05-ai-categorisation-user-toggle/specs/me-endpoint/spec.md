## MODIFIED Requirements

### Requirement: GET /me Endpoint

The system SHALL expose a `GET /me` endpoint that returns the authenticated user's Kinde ID, `vision_enabled` flag, `folder_icons_enabled` flag, `ai_categorisation_enabled` flag, and `ai_categorisation_count_this_month` integer. The endpoint SHALL be in the protected route group and require a valid JWT.

`MeHandler` SHALL depend on a `CategorisationCountUsecase` interface (with `CountThisMonth`) in addition to `UserUsecase` and `AccountUsecase`. The count is fetched on every `GET /me` call.

Response body (200):
```json
{
  "id": "kp_abc123",
  "vision_enabled": false,
  "folder_icons_enabled": true,
  "ai_categorisation_enabled": false,
  "ai_categorisation_count_this_month": 14
}
```

- `id` — the Kinde user ID from the authenticated user's DB record
- `vision_enabled` — boolean from the `users` table
- `folder_icons_enabled` — boolean from the `users` table
- `ai_categorisation_enabled` — boolean from the `users` table
- `ai_categorisation_count_this_month` — count of `ai_categorisation_logs` rows for this user in the current UTC calendar month

#### Scenario: Authenticated user retrieves their profile

- **WHEN** an authenticated request is made to `GET /me`
- **THEN** the response is `200 OK`
- **AND** the body contains the user's Kinde ID, `vision_enabled`, `folder_icons_enabled`, `ai_categorisation_enabled`, and `ai_categorisation_count_this_month`

#### Scenario: Unauthenticated request is rejected

- **WHEN** a request is made to `GET /me` without a valid Bearer token
- **THEN** the response is `401 Unauthorized`
- **AND** no user data is returned

---

### Requirement: PATCH /me Endpoint

The system SHALL expose a `PATCH /me` endpoint that updates the authenticated user's `vision_enabled`, `folder_icons_enabled`, and/or `ai_categorisation_enabled` flags. The endpoint SHALL be in the protected route group and require a valid JWT. Only fields present in the request body are modified — omitted fields are left unchanged.

Request body (at least one of the three fields required):
```json
{ "vision_enabled": true, "folder_icons_enabled": false, "ai_categorisation_enabled": true }
```

Response body (200) — same shape as `GET /me`:
```json
{
  "id": "kp_abc123",
  "vision_enabled": true,
  "folder_icons_enabled": false,
  "ai_categorisation_enabled": true,
  "ai_categorisation_count_this_month": 14
}
```

- `vision_enabled`, `folder_icons_enabled`, and `ai_categorisation_enabled` are the writable preference fields. The request body SHALL NOT permit updating any other user column (`id`, `pending_kinde_deletion`, timestamps).
- At least one of the three fields is REQUIRED in the request body; an empty body is a client error.
- Each field present MUST be a boolean; any other JSON type for a present field is a client error.
- The response includes `ai_categorisation_count_this_month` (same as `GET /me`) so the client receives the current count after a preference change.

#### Scenario: Authenticated user enables vision labelling

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "vision_enabled": true }`
- **THEN** the response is `200 OK`
- **AND** the body contains `"vision_enabled": true` and all other fields unchanged
- **AND** the user's `vision_enabled` column is persisted as `true`

#### Scenario: Authenticated user disables vision labelling

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "vision_enabled": false }`
- **THEN** the response is `200 OK`
- **AND** the body contains `"vision_enabled": false` and all other fields unchanged
- **AND** the user's `vision_enabled` column is persisted as `false`

#### Scenario: Authenticated user disables folder icons

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "folder_icons_enabled": false }`
- **THEN** the response is `200 OK`
- **AND** the body contains `"folder_icons_enabled": false`
- **AND** the user's `vision_enabled` column is unchanged

#### Scenario: Authenticated user re-enables folder icons

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "folder_icons_enabled": true }`
- **THEN** the response is `200 OK`
- **AND** the user's `folder_icons_enabled` column is persisted as `true`

#### Scenario: Authenticated user enables AI auto-categorisation

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "ai_categorisation_enabled": true }`
- **THEN** the response is `200 OK`
- **AND** the body contains `"ai_categorisation_enabled": true` and all other fields unchanged
- **AND** the user's `ai_categorisation_enabled` column is persisted as `true`

#### Scenario: Authenticated user disables AI auto-categorisation

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "ai_categorisation_enabled": false }`
- **THEN** the response is `200 OK`
- **AND** the body contains `"ai_categorisation_enabled": false`
- **AND** the user's `ai_categorisation_enabled` column is persisted as `false`

#### Scenario: Missing all fields is rejected

- **WHEN** an authenticated request is made to `PATCH /me` with an empty body `{}`
- **THEN** the response is `400 Bad Request`
- **AND** none of the user's preference columns are changed

#### Scenario: Non-boolean vision_enabled is rejected

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "vision_enabled": "yes" }`
- **THEN** the response is `400 Bad Request`
- **AND** the user's `vision_enabled` column is unchanged

#### Scenario: Non-boolean folder_icons_enabled is rejected

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "folder_icons_enabled": "yes" }`
- **THEN** the response is `400 Bad Request`
- **AND** the user's `folder_icons_enabled` column is unchanged

#### Scenario: Non-boolean ai_categorisation_enabled is rejected

- **WHEN** an authenticated request is made to `PATCH /me` with body `{ "ai_categorisation_enabled": "yes" }`
- **THEN** the response is `400 Bad Request`
- **AND** the user's `ai_categorisation_enabled` column is unchanged

#### Scenario: Unauthenticated request is rejected

- **WHEN** a request is made to `PATCH /me` without a valid Bearer token
- **THEN** the response is `401 Unauthorized`
- **AND** no user data is changed
