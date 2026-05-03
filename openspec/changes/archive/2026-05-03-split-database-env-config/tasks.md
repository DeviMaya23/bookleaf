## 1. Config Model and Environment Inputs

- [x] 1.1 Update `backend/internal/config/config.go` `DBConfig` to include split DB fields (`Host`, `Name`, `Port`, `User`, `Password`, `SSLMode`) while retaining `URL`
- [x] 1.2 Replace required `DATABASE_URL` loading with required split env var loading (`DATABASE_HOST`, `DATABASE_NAME`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_SSLMODE`)
- [x] 1.3 Add optional `DATABASE_OPTIONS` handling and append it to the constructed DB URL query string
- [x] 1.4 Build `cfg.DB.URL` from split fields in `loadFromEnv()` using safe URL construction and encoding

## 2. Server Bootstrap and Runtime Wiring

- [x] 2.1 Ensure `backend/cmd/server/main.go` DB initialization remains wired to the constructed `cfg.DB.URL`
- [x] 2.2 Update any startup error paths/messages that still reference legacy `DATABASE_URL`

## 3. Developer Environment Documentation

- [x] 3.1 Update `.env.example` to remove `DATABASE_URL` and add `DATABASE_HOST`, `DATABASE_NAME`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_SSLMODE`
- [x] 3.2 Document optional `DATABASE_OPTIONS` usage in `.env.example` with an example value

## 4. Tests and Validation

- [x] 4.1 Update `backend/internal/config/config_test.go` required-var tests to assert split DB env validation failures name the missing variable
- [x] 4.2 Add config unit tests asserting deterministic `cfg.DB.URL` construction from split env vars
- [x] 4.3 Add config unit tests asserting `DATABASE_OPTIONS` is included in constructed URL when set
