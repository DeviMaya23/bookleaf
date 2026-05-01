## 1. Create internal/config package

- [x] 1.1 Create `backend/internal/config/config.go` with `KindeConfig`, `DBConfig`, and `Config` structs
- [x] 1.2 Implement `loadFromEnv()` — reads env vars, applies defaults, returns `(*Config, error)` without touching godotenv
- [x] 1.3 Implement `Load()` — calls `godotenv.Load()` (warn on missing file), then delegates to `loadFromEnv()`
- [x] 1.4 Move `requireEnv` helper into `internal/config` (unexported, used by `loadFromEnv`)
- [x] 1.5 Add `envWithDefault(name, fallback string) string` helper and use it for `PORT` in `loadFromEnv()`

## 2. Unit tests

- [x] 2.1 Create `backend/internal/config/config_test.go`
- [x] 2.2 Test: all required vars set → `Load` returns populated config, nil error
- [x] 2.3 Test: each required var missing individually → error names the missing var
- [x] 2.4 Test: `PORT` unset → `cfg.Port` is `"8080"`
- [x] 2.5 Test: `PORT` set → `cfg.Port` reflects the value
- [x] 2.6 Test: `PORT` set to empty string → `cfg.Port` is `"8080"` (fallback applies)
- [x] 2.7 Run `go test ./internal/config/...` and confirm all pass

## 3. Wire config into main.go

- [x] 3.1 Replace `godotenv.Load()` call in `main.go` with `config.Load()`
- [x] 3.2 Remove inline `requireEnv` function from `main.go`
- [x] 3.3 Update all usages in `main.go` to read from `cfg` (`cfg.Kinde.IssuerURL`, `cfg.Kinde.Audience`, `cfg.DB.URL`, `cfg.Port`)

## 4. Verify

- [x] 4.1 Run `go build ./...` — no compile errors
- [x] 4.2 Start server locally with a `.env` file and confirm it boots correctly
- [x] 4.3 Start server with a missing required var and confirm it exits with a clear error message
