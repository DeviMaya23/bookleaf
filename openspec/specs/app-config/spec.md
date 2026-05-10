## ADDED Requirements

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

### Requirement: Observability Config Sub-Struct

The `Config` struct SHALL include an `Obs ObsConfig` field. `ObsConfig` SHALL have:

- `OTELEnabled bool` — loaded from `OTEL_ENABLED`; optional, defaults to `false`
- `OTELExporter string` — loaded from `OTEL_EXPORTER`; **conditionally required**: only validated as non-empty when `OTELEnabled` is `true`
- `OTELMetricsExporter string` — loaded from `OTEL_METRICS_EXPORTER`; **conditionally required**: only validated as non-empty when `OTELEnabled` is `true`
- `LogFormat string` — loaded from `LOG_FORMAT`; optional, defaults to `"json"`

When `OTELEnabled` is `false`, `OTEL_EXPORTER` and `OTEL_METRICS_EXPORTER` SHALL be loaded as empty strings without error, even if unset.

#### Scenario: All observability vars are set with OTel enabled

- **WHEN** `OTEL_ENABLED=true`, `OTEL_EXPORTER=tempo`, `OTEL_METRICS_EXPORTER=prometheus`, and `LOG_FORMAT=json` are set
- **THEN** `cfg.Obs.OTELEnabled` is `true`, `cfg.Obs.OTELExporter` is `"tempo"`, `cfg.Obs.OTELMetricsExporter` is `"prometheus"`, and `cfg.Obs.LogFormat` is `"json"`

#### Scenario: LOG_FORMAT defaults to json

- **WHEN** `LOG_FORMAT` is not set in the environment
- **THEN** `cfg.Obs.LogFormat` is `"json"`

#### Scenario: OTEL_ENABLED defaults to false

- **WHEN** `OTEL_ENABLED` is not set in the environment
- **THEN** `cfg.Obs.OTELEnabled` is `false`

#### Scenario: OTEL_EXPORTER missing is not an error when OTel disabled

- **WHEN** `OTEL_ENABLED` is not set (or `false`)
- **AND** `OTEL_EXPORTER` is not set
- **THEN** `config.Load()` returns a non-nil `*Config` with a nil error
- **AND** `cfg.Obs.OTELExporter` is `""`

#### Scenario: OTEL_METRICS_EXPORTER missing is not an error when OTel disabled

- **WHEN** `OTEL_ENABLED` is not set (or `false`)
- **AND** `OTEL_METRICS_EXPORTER` is not set
- **THEN** `config.Load()` returns a non-nil `*Config` with a nil error
- **AND** `cfg.Obs.OTELMetricsExporter` is `""`

#### Scenario: OTEL_EXPORTER missing causes startup failure when OTel enabled

- **WHEN** `OTEL_ENABLED=true`
- **AND** `OTEL_EXPORTER` is not set
- **THEN** `config.Load()` returns a non-nil error naming the missing variable

#### Scenario: OTEL_METRICS_EXPORTER missing causes startup failure when OTel enabled

- **WHEN** `OTEL_ENABLED=true`
- **AND** `OTEL_METRICS_EXPORTER` is not set
- **THEN** `config.Load()` returns a non-nil error naming the missing variable

### Requirement: Vision Config Sub-Struct

The `Config` struct SHALL include a `Vision VisionConfig` field. `VisionConfig` SHALL have:

- `APIKey string` — loaded from `GOOGLE_VISION_API_KEY`; **optional** (empty string if unset)

`config.Load()` SHALL NOT return an error if `GOOGLE_VISION_API_KEY` is absent. When `APIKey` is empty, the application starts normally and Vision features are skipped at runtime.

#### Scenario: GOOGLE_VISION_API_KEY is set

- **WHEN** `GOOGLE_VISION_API_KEY=abc123` is present in the environment
- **THEN** `cfg.Vision.APIKey` is `"abc123"`

#### Scenario: GOOGLE_VISION_API_KEY is absent

- **WHEN** `GOOGLE_VISION_API_KEY` is not set
- **THEN** `config.Load()` returns a non-nil `*Config` with `cfg.Vision.APIKey` equal to `""`
- **AND** `config.Load()` returns a nil error
