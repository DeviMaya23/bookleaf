## ADDED Requirements

### Requirement: Maintenance mode gates protected routes
The server SHALL, when `MAINTENANCE_MODE` is enabled, respond to every request in the protected route group with `503 Service Unavailable`, a JSON body `{"error":"maintenance"}`, and a response header `X-Bookleaf-Maintenance: true`, without executing the route handler. Routes outside the protected group (`/health`, `/metrics`) SHALL be unaffected.

#### Scenario: Protected route is gated while maintenance mode is enabled
- **WHEN** `MAINTENANCE_MODE` is enabled and a request is made to a protected route (e.g. `GET /me`)
- **THEN** the response is `503 Service Unavailable`
- **AND** the response body is `{"error":"maintenance"}`
- **AND** the response includes header `X-Bookleaf-Maintenance: true`

#### Scenario: Health endpoint is unaffected by maintenance mode
- **WHEN** `MAINTENANCE_MODE` is enabled and `GET /health` is called
- **THEN** the response is `200 OK` as normal

#### Scenario: Maintenance mode disabled leaves protected routes unaffected
- **WHEN** `MAINTENANCE_MODE` is disabled or unset
- **THEN** requests to protected routes are handled normally, with no `X-Bookleaf-Maintenance` header

### Requirement: Maintenance gate runs before authentication
The maintenance middleware SHALL be applied to the protected route group before the authentication middleware, so that gated requests do not require a valid `Authorization` header and do not trigger JWT validation or user provisioning.

#### Scenario: Gated request without Authorization header returns 503, not 401
- **WHEN** `MAINTENANCE_MODE` is enabled and a request to a protected route is made with no `Authorization` header
- **THEN** the response is `503 Service Unavailable` with `X-Bookleaf-Maintenance: true`
- **AND** the response is not `401 Unauthorized`

### Requirement: Bypass header skips the maintenance gate
The server SHALL skip the maintenance gate for any request to the protected group whose `X-Bookleaf-Bypass` header exactly matches a non-empty `MAINTENANCE_BYPASS_TOKEN` configuration value, regardless of `MAINTENANCE_MODE`. If `MAINTENANCE_BYPASS_TOKEN` is unset or empty, no value of `X-Bookleaf-Bypass` SHALL be treated as a bypass.

#### Scenario: Matching bypass header skips the gate
- **WHEN** `MAINTENANCE_MODE` is enabled and a request includes header `X-Bookleaf-Bypass` equal to the configured `MAINTENANCE_BYPASS_TOKEN`
- **THEN** the request proceeds to normal routing (including authentication) instead of receiving the maintenance response

#### Scenario: Non-matching bypass header does not skip the gate
- **WHEN** `MAINTENANCE_MODE` is enabled and a request includes header `X-Bookleaf-Bypass` that does not match the configured `MAINTENANCE_BYPASS_TOKEN`
- **THEN** the response is `503 Service Unavailable` with `X-Bookleaf-Maintenance: true`

#### Scenario: Bypass header has no effect when no token is configured
- **WHEN** `MAINTENANCE_BYPASS_TOKEN` is unset or empty and `MAINTENANCE_MODE` is enabled
- **THEN** any request, including one with an `X-Bookleaf-Bypass` header, receives the maintenance response

### Requirement: Maintenance configuration via environment variables
`config.Load()` SHALL populate `cfg.Maintenance.Enabled` from the `MAINTENANCE_MODE` environment variable, parsed as a boolean (`strconv.ParseBool`), defaulting to `false` if unset or invalid. `config.Load()` SHALL populate `cfg.Maintenance.BypassToken` from the `MAINTENANCE_BYPASS_TOKEN` environment variable, defaulting to an empty string if unset.

#### Scenario: MAINTENANCE_MODE unset defaults to disabled
- **WHEN** `MAINTENANCE_MODE` is not set in the environment
- **THEN** `cfg.Maintenance.Enabled` is `false`

#### Scenario: MAINTENANCE_MODE set to true enables maintenance mode
- **WHEN** `MAINTENANCE_MODE` is set to `"true"`
- **THEN** `cfg.Maintenance.Enabled` is `true`

#### Scenario: MAINTENANCE_MODE set to an invalid value defaults to disabled
- **WHEN** `MAINTENANCE_MODE` is set to a value that is not a valid boolean (e.g. `"notabool"`)
- **THEN** `cfg.Maintenance.Enabled` is `false`

#### Scenario: MAINTENANCE_BYPASS_TOKEN unset defaults to empty
- **WHEN** `MAINTENANCE_BYPASS_TOKEN` is not set in the environment
- **THEN** `cfg.Maintenance.BypassToken` is `""`
