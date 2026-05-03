## Context

The backend currently expects a pre-built `DATABASE_URL` string and passes it directly into `gorm.Open(postgres.Open(cfg.DB.URL), ...)`. This makes configuration less explicit and pushes DSN correctness to deploy-time string authoring. The requested change introduces split DB env variables while preserving a single canonical `cfg.DB.URL` used by the database initializer.

## Goals / Non-Goals

**Goals:**
- Load required DB connection fields from explicit env vars (`DATABASE_HOST`, `DATABASE_NAME`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_SSLMODE`)
- Build `cfg.DB.URL` centrally inside config loading
- Keep DB initialization in `cmd/server/main.go` simple and compatible with existing GORM usage
- Update `.env.example` and startup requirements to match the new env model

**Non-Goals:**
- Changing DB driver, pool tuning, or migration tooling
- Adding multiple database targets/read replicas
- Changing API behavior

## Decisions

### Decision 1: Keep `cfg.DB.URL` as the runtime contract
`DBConfig` will hold split fields plus `URL`, but runtime consumers continue to use `cfg.DB.URL` for connection initialization.

**Rationale:** avoids broad call-site churn and keeps DB init logic stable while still improving config ergonomics.

**Alternative considered:** changing all DB initialization code to consume split fields directly; rejected as unnecessary churn.

### Decision 2: Build URL in `config.loadFromEnv()`
The DSN will be assembled in one place using the split env vars and encoded query params (at minimum `sslmode`).

**Rationale:** centralizes validation and formatting; tests can assert a deterministic output URL.

**Alternative considered:** assembling DSN in `main.go`; rejected because validation and formatting belong in config.

### Decision 3: Treat all split DB vars as required
Each split DB var is required and missing vars return explicit startup errors naming the variable.

**Rationale:** preserves fail-fast behavior equivalent to the current required `DATABASE_URL`.

**Alternative considered:** defaults for host/port/sslmode; rejected to avoid hidden environment assumptions.

### Decision 4: Support optional DB URL extras
Add optional `DATABASE_OPTIONS` (query string key-value pairs, e.g. `connect_timeout=10&application_name=bookleaf`) appended after `sslmode`.

**Rationale:** covers advanced Postgres DSN needs without introducing many narrowly-scoped env vars.

**Alternative considered:** adding dedicated env var per option; rejected as unscalable.

## Risks / Trade-offs

- **Misconfigured split vars produce malformed DSN** → Mitigation: validate required vars and cover DSN construction with config unit tests.
- **Password special characters can break naive string concatenation** → Mitigation: percent-encode user/password and query params via standard URL building.
- **Environment migration complexity across deploy targets** → Mitigation: update `.env.example` and specs to document required replacements from `DATABASE_URL`.

## Migration Plan

1. Introduce split DB env loading + DSN construction in `internal/config/config.go`.
2. Update config tests for required split vars and URL output.
3. Update `.env.example` by removing `DATABASE_URL` and adding split DB variables (plus optional `DATABASE_OPTIONS`).
4. Keep `cmd/server/main.go` DB init on `cfg.DB.URL` (or adjust minimally if needed) to avoid behavior drift.
5. Rollout: set split DB env vars in each environment before deploying; remove legacy `DATABASE_URL` from secret stores.

Rollback: reintroduce `DATABASE_URL` support path if deployment issues occur.

## Open Questions

- Should we continue supporting legacy `DATABASE_URL` as a temporary fallback for one release, or enforce split vars immediately?
- Do we want to enforce numeric validation for `DATABASE_PORT` in config load, or allow pass-through and fail at connect time?
