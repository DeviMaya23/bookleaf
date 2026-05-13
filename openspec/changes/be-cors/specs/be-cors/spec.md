## ADDED Requirements

### Requirement: CORS origins configured via environment variable
The server SHALL read allowed CORS origins from the `CORS_ALLOWED_ORIGINS` environment variable. The value is a comma-separated list of origins with no spaces (e.g. `http://localhost:5173,https://app.example.com`). The variable is required — the server SHALL fail to start if it is absent or empty.

#### Scenario: Server starts with valid origins configured
- **WHEN** `CORS_ALLOWED_ORIGINS` is set to one or more valid origins
- **THEN** the server starts successfully and CORS middleware is active

#### Scenario: Server fails to start without the variable
- **WHEN** `CORS_ALLOWED_ORIGINS` is absent or empty
- **THEN** the server fails to start with a descriptive configuration error

### Requirement: CORS preflight requests are handled before authentication
The server SHALL respond to `OPTIONS` preflight requests with the appropriate CORS headers without requiring an `Authorization` header. Preflight requests SHALL NOT be passed to the authentication middleware.

#### Scenario: Preflight from allowed origin succeeds
- **WHEN** a browser sends an `OPTIONS` request with `Origin: <allowed-origin>`
- **THEN** the response is `204` with `Access-Control-Allow-Origin: <allowed-origin>`
- **AND** no `Authorization` header is required

#### Scenario: Preflight from disallowed origin is rejected
- **WHEN** a browser sends an `OPTIONS` request with an origin not in `CORS_ALLOWED_ORIGINS`
- **THEN** the response does not include `Access-Control-Allow-Origin`

### Requirement: Cross-origin API requests include required CORS headers
The server SHALL include `Access-Control-Allow-Origin` on responses to cross-origin requests from allowed origins. The server SHALL permit `Authorization` and `Content-Type` request headers.

#### Scenario: Authenticated request from allowed origin succeeds
- **WHEN** a cross-origin request is made from an allowed origin with `Authorization` and `Content-Type` headers
- **THEN** the response includes `Access-Control-Allow-Origin: <allowed-origin>`
- **AND** the response body is returned normally

#### Scenario: Request from disallowed origin is blocked
- **WHEN** a cross-origin request is made from an origin not in `CORS_ALLOWED_ORIGINS`
- **THEN** the response does not include `Access-Control-Allow-Origin`
