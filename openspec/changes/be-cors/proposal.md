## Why

The frontend (running on `localhost:5173` in dev, a production domain in prod) is blocked by browsers from calling the backend API because no CORS headers are set. Now that the frontend is actively making API calls, CORS must be configured before any cross-origin request can succeed.

## What Changes

- Add `CORS_ALLOWED_ORIGINS` environment variable (required, comma-separated list of allowed origins)
- Register Echo's built-in CORS middleware globally on the server, configured from the env var
- Document `CORS_ALLOWED_ORIGINS` in `.env.example`

## Capabilities

### New Capabilities

- `be-cors`: CORS middleware configuration — allowed origins driven by environment variable

### Modified Capabilities

<!-- none -->

## Impact

- `internal/config/config.go` — new `CORSAllowedOrigins` field and env var loading
- `cmd/server/main.go` — register CORS middleware on the Echo instance
- `.env.example` — document the new required var
- No new dependencies — Echo's `echomiddleware` package already provides CORS support
- Local `.env` files will need `CORS_ALLOWED_ORIGINS=http://localhost:5173` added
