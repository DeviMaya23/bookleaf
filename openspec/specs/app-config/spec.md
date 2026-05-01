## ADDED Requirements

### Requirement: Typed Config Struct
The `internal/config` package SHALL define a `Config` struct grouping related env vars by domain:
- `Kinde KindeConfig` — Kinde-related vars (`IssuerURL`, `Audience`)
- `DB DBConfig` — database vars (`URL`)
- `Port string` — HTTP server port (top-level, not domain-specific)

#### Scenario: Config fields are accessible by domain
- **WHEN** `config.Load()` returns successfully
- **THEN** callers can access `cfg.Kinde.IssuerURL`, `cfg.Kinde.Audience`, `cfg.DB.URL`, and `cfg.Port`

### Requirement: Required Env Var Validation
`config.Load()` SHALL return an error if any required env var (`KINDE_ISSUER_URL`, `KINDE_AUDIENCE`, `DATABASE_URL`) is unset or empty. The error message SHALL name the missing variable.

#### Scenario: Missing required var
- **WHEN** one or more required env vars are not set
- **THEN** `config.Load()` returns a non-nil error naming the missing variable
- **AND** the returned `*Config` is nil

#### Scenario: All required vars present
- **WHEN** all required env vars are set to non-empty values
- **THEN** `config.Load()` returns a populated `*Config` and a nil error

### Requirement: Optional Env Var Defaults
The `internal/config` package SHALL provide an unexported `envWithDefault(name, fallback string) string` helper for optional env vars. It returns the env var value if set and non-empty, otherwise the fallback string. `loadFromEnv()` SHALL use this helper for all optional vars.
- `PORT` defaults to `"8080"`

#### Scenario: PORT not set
- **WHEN** `PORT` is not set in the environment
- **THEN** `cfg.Port` is `"8080"`

#### Scenario: PORT is set
- **WHEN** `PORT` is set to a non-empty value
- **THEN** `cfg.Port` reflects that value

#### Scenario: envWithDefault returns fallback when var is empty
- **WHEN** the named env var is set to an empty string
- **THEN** `envWithDefault` returns the fallback value

### Requirement: Local Dev dotenv Loading
`config.Load()` SHALL attempt to load a `.env` file via godotenv before reading env vars. A missing `.env` file SHALL NOT be treated as an error — it is expected in non-local environments. Vars already set in the environment SHALL NOT be overwritten by the `.env` file.

#### Scenario: .env file present
- **WHEN** a `.env` file exists in the working directory
- **THEN** its values are loaded into the environment before validation

#### Scenario: .env file absent
- **WHEN** no `.env` file exists
- **THEN** `config.Load()` continues without error and reads from the existing environment

#### Scenario: Injected env vars not overwritten
- **WHEN** a var is already set in the environment and also present in `.env`
- **THEN** the pre-existing environment value is used

### Requirement: Config Package Unit Tests
The `internal/config` package SHALL have unit tests covering required var validation, optional var defaults, and error message content. Tests SHALL NOT rely on `.env` files on disk.

#### Scenario: Unit test for missing required var
- **WHEN** a required var is absent from the test environment
- **THEN** the test asserts `Load` (or the internal loader) returns an error containing the variable name

#### Scenario: Unit test for defaults
- **WHEN** `PORT` is not set in the test environment
- **THEN** the test asserts `cfg.Port` equals `"8080"`
