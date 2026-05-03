## Why

`DATABASE_URL` is currently provided as one opaque string, which is harder to validate, document, and configure consistently across local/dev/prod. Splitting DB configuration into explicit environment variables improves clarity, enables stronger startup validation, and keeps URL construction logic centralized.

## What Changes

- Replace single required `DATABASE_URL` input with split required DB env vars:
  - `DATABASE_HOST`
  - `DATABASE_NAME`
  - `DATABASE_PORT`
  - `DATABASE_USER`
  - `DATABASE_PASSWORD`
  - `DATABASE_SSLMODE`
- Build `cfg.DB.URL` from those variables in `internal/config/config.go` so downstream DB initialization can keep using one canonical DSN string.
- Update `.env.example` to document the new DB env vars and remove `DATABASE_URL`.
- Update server bootstrap behavior/docs to require the split DB env vars at startup.
- Keep DSN construction compatible with current GORM postgres driver usage in `cmd/server/main.go`.

## Capabilities

### New Capabilities
- `db-dsn-construction`: deterministic construction of Postgres DSN/URL from split environment variables in config loading.

### Modified Capabilities
- `app-config`: change required DB configuration from `DATABASE_URL` to split DB env vars and expose constructed `cfg.DB.URL`.
- `server-bootstrap`: update startup requirements so server fails fast when required split DB env vars are missing.

## Impact

- **Affected code:** `backend/internal/config/config.go`, `backend/internal/config/config_test.go`, `backend/cmd/server/main.go` (if DB init path needs minor adjustments), `.env.example`
- **APIs:** No HTTP/API contract changes
- **Dependencies:** No new runtime dependency required; uses existing Go stdlib URL/query handling for safe DSN assembly
- **Operational impact:** Environment configuration changes required in all deployment environments
