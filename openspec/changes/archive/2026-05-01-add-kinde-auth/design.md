## Context

The backend is a bare Echo server with no auth. All routes are currently unprotected. Kinde is the chosen identity provider — it handles login, registration, and token issuance. The backend's only responsibility is validating the JWTs Kinde issues and provisioning a local User record on first contact.

## Goals / Non-Goals

**Goals:**
- Validate Kinde JWTs on protected routes using Kinde's public JWKS endpoint
- Auto-provision a User record in PostgreSQL on first authenticated request
- Make the authenticated user ID available to all handlers via Echo context
- Expose `GET /me` as the first protected endpoint

**Non-Goals:**
- Login/logout/registration flows — owned by Kinde and the future frontend
- Role-based access control
- Token refresh logic — handled client-side by Kinde SDK
- Rate limiting

## Decisions

### JWKS-based JWT validation (no client secret)

Validate tokens by fetching Kinde's public keys from `{KINDE_ISSUER_URL}/.well-known/jwks` rather than using a shared secret.

**Why:** RS256 asymmetric signing is Kinde's default. JWKS validation requires no secret on the backend — only the issuer URL and audience. Keys are cached locally with TTL to avoid fetching on every request.

**Library:** `github.com/MicahParks/jwkset` or a leaky-bucket JWKS cache via `github.com/golang-jwt/jwt/v5` + a JWKS fetcher. Both are well-maintained; `golang-jwt/jwt` is the more widely used foundation.

### Middleware placement: Echo group, not global

Auth middleware is applied to a protected route group, not globally. The `/health` endpoint stays unauthenticated.

```
e.GET("/health", ...)           ← no middleware
protected := e.Group("")
protected.Use(AuthMiddleware)
protected.GET("/me", ...)
```

**Why:** Explicit opt-in is clearer than global with carve-outs. New routes must consciously join the protected group.

### User auto-provisioning inside middleware

On a valid JWT for an unknown user, the middleware calls the user usecase to create the record before passing control to the handler.

**Why:** Keeps handlers clean — every handler that runs can assume the user exists in the DB. The alternative (provisioning in each handler) would be repetitive and error-prone.

**Idempotency:** Use `INSERT ... ON CONFLICT DO NOTHING` so concurrent first-requests don't cause duplicate key errors.

### Auth context key: typed constant

Store the authenticated user ID on Echo context using a typed constant key (not a raw string) to avoid collisions.

```go
type contextKey string
const userIDKey contextKey = "userID"
```

### `/me` response shape

```json
{ "id": "kp_abc123", "vision_enabled": false }
```

Minimal — returns only what's stored in the DB. No proxying to Kinde's user API.

## Risks / Trade-offs

- [JWKS fetch failure at startup] If Kinde is unreachable, the first token validation will fail. → Mitigation: JWKS client retries with backoff; 503 returned to client rather than panic.
- [Clock skew on JWT expiry] Slight clock drift between server and Kinde could cause valid tokens to appear expired. → Mitigation: `golang-jwt` allows a small leeway (e.g. 5s), enabled by default.
- [Auto-provisioning on every request for new users] The DB write on first request adds latency for that one call. → Acceptable at this scale; no mitigation needed.

## Migration Plan

No data migration needed. The `users` table already exists. New routes are additive. Deploying this change makes `/me` (and future protected routes) require a valid JWT — no previously-open routes are being locked down.
