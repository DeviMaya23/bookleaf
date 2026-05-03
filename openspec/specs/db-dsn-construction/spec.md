## ADDED Requirements

### Requirement: Database URL Construction From Split Environment Variables
The configuration loader SHALL construct a PostgreSQL connection URL from split required environment variables instead of reading a single `DATABASE_URL` value.

Required variables:
- `DATABASE_HOST`
- `DATABASE_NAME`
- `DATABASE_PORT`
- `DATABASE_USER`
- `DATABASE_PASSWORD`
- `DATABASE_SSLMODE`

The constructed URL SHALL be exposed as `cfg.DB.URL`.

#### Scenario: Build URL from required split vars
- **WHEN** all required split DB env vars are set
- **THEN** `config.Load()` returns a non-nil config with `cfg.DB.URL` populated as a valid PostgreSQL URL
- **AND** the URL includes host, port, database name, user info, and `sslmode` query parameter

### Requirement: Optional Database URL Query Options
The configuration loader SHALL support optional `DATABASE_OPTIONS` for additional DSN query parameters.

When set, `DATABASE_OPTIONS` SHALL be appended to the generated URL query string in addition to `sslmode`.

#### Scenario: Additional options are appended
- **WHEN** `DATABASE_OPTIONS` is set (for example `connect_timeout=10&application_name=bookleaf`)
- **THEN** `cfg.DB.URL` includes those query parameters and `sslmode`

#### Scenario: No additional options provided
- **WHEN** `DATABASE_OPTIONS` is not set
- **THEN** `cfg.DB.URL` is still constructed successfully using required split vars and `sslmode` only
