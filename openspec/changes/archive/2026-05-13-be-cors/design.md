## Context

The Echo server has no CORS middleware today. The `echomiddleware` package (already imported for `Recover()`) provides `CORSWithConfig`, so no new dependency is needed. Config loading follows a consistent pattern: required values go through `requireEnv`, optional ones through `envWithDefault`. CORS origins are security-sensitive, so they should be required — misconfiguration should be loud.

## Goals / Non-Goals

**Goals:**
- Allow browsers to make cross-origin requests to the API from configured origins
- Keep allowed origins out of source code — driven entirely by environment variable
- Follow existing config conventions (`requireEnv`, typed config struct)

**Non-Goals:**
- Per-route CORS configuration
- Credentials / cookie support (not needed — auth is Bearer token via header)
- Wildcard support (`*`) — intentionally excluded to prevent accidental open access in prod

## Decisions

**1. `CORS_ALLOWED_ORIGINS` is a required env var, comma-separated**
`requireEnv` is used so the server fails fast if the var is missing. A comma-separated string (e.g. `http://localhost:5173,https://app.example.com`) is the simplest multi-origin format and easy to set in any environment.
- Alternative: separate `CORS_ORIGIN_1`, `CORS_ORIGIN_2` vars — unnecessary complexity.
- Alternative: `envWithDefault` with `*` fallback — rejected; wildcard in prod is a security risk.

**2. Middleware registered globally on `e`, not on the `protected` group**
CORS preflight requests (`OPTIONS`) are sent before auth headers are attached, so the CORS middleware must run before the auth middleware. Registering on `e` ensures preflights are handled correctly and don't return 401.
- Alternative: register on `protected` group — preflights would hit auth middleware first and return 401, breaking the browser handshake.

**3. Use `echomiddleware.CORSWithConfig` with explicit `AllowOrigins`**
Echo's built-in CORS middleware handles the `OPTIONS` preflight, sets `Access-Control-Allow-Origin`, and passes through `AllowHeaders` / `AllowMethods`. No custom middleware needed.

**4. `AllowHeaders` includes `Authorization` and `Content-Type`**
The frontend sends `Authorization: Bearer <token>` and `Content-Type: application/json` on every authenticated request. Both must be in `AllowHeaders` or the browser will block the preflight.

## Risks / Trade-offs

- [Misconfigured origins in prod] → `requireEnv` makes missing config a startup failure, not a silent bug. Operators must set the var deliberately.
- [Multiple origins parsed from a single string] → `strings.Split` on comma is simple but requires no spaces around commas. Document this clearly in `.env.example`.
