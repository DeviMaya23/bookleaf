## Why

The app has no authentication layer — all routes are currently open. Implementing Kinde JWT auth is the prerequisite for every feature that requires a known user (image upload, folder management, AI organising).

## What Changes

- Add JWT validation middleware that verifies Kinde Bearer tokens on protected routes
- Add user auto-provisioning: first request from a valid but unknown user creates their DB record
- Add auth context: authenticated user ID is set on Echo context for downstream handlers
- Add `GET /me` endpoint returning the authenticated user's Kinde ID and `vision_enabled` flag

## Capabilities

### New Capabilities

- `kinde-auth`: JWT middleware, user auto-provisioning, and auth context
- `me-endpoint`: `GET /me` endpoint for the authenticated user

### Modified Capabilities

- `server-bootstrap`: server now loads Kinde env vars and registers auth middleware + `/me` route

## Impact

- New Go dependency: Kinde JWT validation library (JWKS-based)
- `backend/internal/middleware/` — new auth middleware package
- `backend/internal/handler/` — new me handler
- `backend/internal/usecase/` — new user usecase (provision + fetch)
- `backend/internal/repository/` — new user repository
- `backend/cmd/server/main.go` — wire middleware and route
- Requires `KINDE_ISSUER_URL` and `KINDE_AUDIENCE` env vars at runtime
