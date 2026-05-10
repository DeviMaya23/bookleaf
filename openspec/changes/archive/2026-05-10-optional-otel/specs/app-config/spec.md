## MODIFIED Requirements

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
