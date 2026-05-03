## MODIFIED Requirements

### Requirement: Kinde Environment Variables
The server SHALL load all required configuration via `config.Load()` at startup before any other initialisation. Required env vars are `KINDE_ISSUER_URL`, `KINDE_AUDIENCE`, `DATABASE_HOST`, `DATABASE_NAME`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, and `DATABASE_SSLMODE`. If any are missing, the server SHALL fail to start with a clear error message naming the missing variable.

#### Scenario: Server starts with all required env vars present
- **WHEN** `KINDE_ISSUER_URL`, `KINDE_AUDIENCE`, `DATABASE_HOST`, `DATABASE_NAME`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, and `DATABASE_SSLMODE` are set in the environment
- **THEN** the server starts successfully

#### Scenario: Server fails without required env vars
- **WHEN** any of `KINDE_ISSUER_URL`, `KINDE_AUDIENCE`, `DATABASE_HOST`, `DATABASE_NAME`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, or `DATABASE_SSLMODE` is missing from the environment
- **THEN** the server exits with a non-zero status code and logs which variable is missing
