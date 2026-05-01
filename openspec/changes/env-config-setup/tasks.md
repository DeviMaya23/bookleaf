## 1. Create internal/config package

- [ ] 1.1 Create `backend/internal/config/config.go` with `KindeConfig`, `DBConfig`, and `Config` structs
- [ ] 1.2 Implement `loadFromEnv()` — reads env vars, applies defaults, returns `(*Config, error)` without touching godotenv
- [ ] 1.3 Implement `Load()` — calls `godotenv.Load()` (warn on missing file), then delegates to `loadFromEnv()`
- [ ] 1.4 Move `requireEnv` helper into `internal/config` (unexported, used by `loadFromEnv`)

## 2. Unit tests

- [ ] 2.1 Create `backend/internal/config/config_test.go`
- [ ] 2.2 Test: all required vars set → `Load` returns populated config, nil error
- [ ] 2.3 Test: each required var missing individually → error names the missing var
- [ ] 2.4 Test: `PORT` unset → `cfg.Port` is `"8080"`
- [ ] 2.5 Test: `PORT` set → `cfg.Port` reflects the value
- [ ] 2.6 Run `go test ./internal/config/...` and confirm all pass

## 3. Wire config into main.go

- [ ] 3.1 Replace `godotenv.Load()` call in `main.go` with `config.Load()`
- [ ] 3.2 Remove inline `requireEnv` function from `main.go`
- [ ] 3.3 Update all usages in `main.go` to read from `cfg` (`cfg.Kinde.IssuerURL`, `cfg.Kinde.Audience`, `cfg.DB.URL`, `cfg.Port`)

## 4. Verify

- [ ] 4.1 Run `go build ./...` — no compile errors
- [ ] 4.2 Start server locally with a `.env` file and confirm it boots correctly
- [ ] 4.3 Start server with a missing required var and confirm it exits with a clear error message
