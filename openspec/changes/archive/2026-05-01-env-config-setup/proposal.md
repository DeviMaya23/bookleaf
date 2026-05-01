## Why

Environment variables are currently read via scattered `os.Getenv` calls in `main.go`, with no central validation or structure. This makes it easy to miss required vars at startup and hard to understand the full config surface of the app.

## What Changes

- Introduce an `internal/config` package with a typed `Config` struct and all env-related helpers (`requireEnv`, etc.) — nothing env-related lives in `main.go`
- All required env vars (`KINDE_ISSUER_URL`, `KINDE_AUDIENCE`, `DATABASE_URL`) validated at startup in one place
- Optional vars (`PORT`) with defaults handled in the config loader
- `godotenv` used to load `.env` for local dev; prod relies on injected environment variables
- `main.go` receives a single `*config.Config` and passes values to dependencies — no more scattered `os.Getenv`
- Unit tests for the config package covering required var validation and defaults

## Capabilities

### New Capabilities
- `app-config`: Typed config struct, loader, and all env-related helpers in `internal/config`; validates required env vars at startup, loads `.env` for local dev via godotenv; unit tested

### Modified Capabilities
- `server-bootstrap`: Startup behavior changes — config is now loaded and validated before any other initialization, with a single structured failure point if required vars are missing

## Impact

- `backend/internal/config/` — new package
- `backend/cmd/server/main.go` — refactored to use `config.Load()`; `requireEnv` and any other env helpers removed from `main.go` entirely
- `backend/internal/config/config_test.go` — unit tests for the config package
- `godotenv` dependency already present (no new deps)
- No API or behavioral changes visible to callers
