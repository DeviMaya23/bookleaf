## MODIFIED Requirements

### Requirement: Typed Config Struct
The `internal/config` package SHALL define a `Config` struct grouping related env vars by domain:
- `Kinde KindeConfig` — Kinde-related vars (`IssuerURL`, `Audience`)
- `DB DBConfig` — database vars (`Host`, `Name`, `Port`, `User`, `Password`, `SSLMode`, `URL`)
- `Port string` — HTTP server port (top-level, not domain-specific)

#### Scenario: Config fields are accessible by domain
- **WHEN** `config.Load()` returns successfully
- **THEN** callers can access `cfg.Kinde.IssuerURL`, `cfg.Kinde.Audience`, `cfg.DB.URL`, and `cfg.Port`
- **AND** callers can access split DB fields from `cfg.DB` (`Host`, `Name`, `Port`, `User`, `Password`, `SSLMode`)

### Requirement: Required Env Var Validation
`config.Load()` SHALL return an error if any required env var (`KINDE_ISSUER_URL`, `KINDE_AUDIENCE`, `DATABASE_HOST`, `DATABASE_NAME`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_SSLMODE`) is unset or empty. The error message SHALL name the missing variable.

#### Scenario: Missing required var
- **WHEN** one or more required env vars are not set
- **THEN** `config.Load()` returns a non-nil error naming the missing variable
- **AND** the returned `*Config` is nil

#### Scenario: All required vars present
- **WHEN** all required env vars are set to non-empty values
- **THEN** `config.Load()` returns a populated `*Config` and a nil error
