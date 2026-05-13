## 1. Config

- [x] 1.1 Add `CORSAllowedOrigins []string` field to the `Config` struct in `internal/config/config.go`
- [x] 1.2 Load `CORS_ALLOWED_ORIGINS` via `requireEnv` and split on comma into `CORSAllowedOrigins`

## 2. Middleware

- [x] 2.1 Register `echomiddleware.CORSWithConfig` on `e` in `cmd/server/main.go`, before auth middleware, using `cfg.CORSAllowedOrigins` as `AllowOrigins` and `Authorization`/`Content-Type` in `AllowHeaders`

## 3. Documentation

- [x] 3.1 Add `CORS_ALLOWED_ORIGINS=http://localhost:5173` to `.env.example` with a comment explaining the comma-separated format
