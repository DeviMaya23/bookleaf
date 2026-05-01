## Context

`main.go` currently reads env vars inline with `os.Getenv` and a local `requireEnv` helper. There are three required vars (`KINDE_ISSUER_URL`, `KINDE_AUDIENCE`, `DATABASE_URL`) and one optional with a default (`PORT`). `godotenv` is already a dependency and loaded at the top of `main`. The goal is to centralise all of this into `internal/config` without changing any runtime behaviour.

## Goals / Non-Goals

**Goals:**
- Single `config.Load()` call in `main.go` that returns a validated `*Config` or an error
- All env helpers (`requireEnv`, optional-with-default) live in `internal/config`, none in `main.go`
- Unit tests exercise: missing required var, all vars present, PORT default
- `.env` loading via godotenv remains dev-only (warn, don't fatal, on missing file)

**Non-Goals:**
- Config hot-reloading
- Support for config files, flags, or third-party config libraries (viper, etc.)
- Validation beyond "is it set and non-empty"

## Decisions

### 1. Grouped `Config` struct by domain

```go
type Config struct {
    Kinde KindeConfig
    DB    DBConfig
    Port  string
}

type KindeConfig struct {
    IssuerURL string
    Audience  string
}

type DBConfig struct {
    URL string
}
```

Grouping by domain (`Kinde`, `DB`) makes the config surface self-documenting and scales naturally as each group grows (e.g. adding `Kinde.ClientID` or `DB.MaxConns` later requires no struct reshuffling). `Port` stays top-level as it belongs to neither domain.

### 2. `Load()` returns `(*Config, error)` — caller decides how to fatal

`Load` returns an error rather than calling `log.Fatal` internally. This keeps the package testable without process exit side-effects and lets `main.go` control the fatal path. The pattern:

```go
cfg, err := config.Load()
if err != nil {
    e.Logger.Fatal(err)
}
```

### 3. godotenv loaded inside `Load()`, not in `main.go`

Moving the `godotenv.Load()` call into `config.Load()` keeps all env concerns in one place. Missing `.env` file logs a warning (via returned error that caller can log at warn level) but does not block startup — same behaviour as today.

Since `godotenv.Load()` only sets vars that aren't already set, injected prod environment variables are never overwritten.

### 4. Unit tests use `t.Setenv` — no test `.env` files

`t.Setenv` is idiomatic (auto-restores after test), doesn't require disk fixtures, and avoids godotenv interaction in tests. Tests call an internal `loadFromEnv()` function that skips the godotenv step, keeping the file-loading path separate from the validation logic.

## Risks / Trade-offs

- **godotenv silently no-ops if var already set** — this is the desired prod behaviour, but worth documenting so future devs don't wonder why their `.env` overrides aren't working in CI. → Note in package-level comment.
- **Sub-struct grouping adds a level of indirection** — callers write `cfg.Kinde.IssuerURL` instead of `cfg.KindeIssuerURL`. Acceptable given the clarity benefit; new domains (e.g. S3, Redis) just add a new sub-struct.
