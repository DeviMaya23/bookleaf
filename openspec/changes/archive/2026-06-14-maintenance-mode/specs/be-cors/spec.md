## MODIFIED Requirements

### Requirement: Cross-origin API requests include required CORS headers
The server SHALL include `Access-Control-Allow-Origin` on responses to cross-origin requests from allowed origins. The server SHALL permit `Authorization`, `Content-Type`, and `X-Bookleaf-Bypass` request headers. The server SHALL expose the `X-Bookleaf-Maintenance` response header via `Access-Control-Expose-Headers` so that frontend JavaScript can read it on cross-origin responses.

#### Scenario: Authenticated request from allowed origin succeeds
- **WHEN** a cross-origin request is made from an allowed origin with `Authorization` and `Content-Type` headers
- **THEN** the response includes `Access-Control-Allow-Origin: <allowed-origin>`
- **AND** the response body is returned normally

#### Scenario: Request from disallowed origin is blocked
- **WHEN** a cross-origin request is made from an origin not in `CORS_ALLOWED_ORIGINS`
- **THEN** the response does not include `Access-Control-Allow-Origin`

#### Scenario: Request with bypass header from allowed origin succeeds
- **WHEN** a cross-origin request is made from an allowed origin with an `X-Bookleaf-Bypass` header
- **THEN** the response includes `Access-Control-Allow-Origin: <allowed-origin>`
- **AND** the response body is returned normally

#### Scenario: Maintenance header is readable by frontend JavaScript
- **WHEN** a cross-origin request is made from an allowed origin
- **THEN** the response includes `Access-Control-Expose-Headers` containing `X-Bookleaf-Maintenance`
