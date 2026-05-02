## ADDED Requirements

### Requirement: Observability Config Sub-Struct

The `Config` struct SHALL include an `Obs ObsConfig` field. `ObsConfig` SHALL have:

- `OTELExporter string` — loaded from `OTEL_EXPORTER`; **required** (server fails to start if unset)
- `OTELMetricsExporter string` — loaded from `OTEL_METRICS_EXPORTER`; **required** (server fails to start if unset)
- `LogFormat string` — loaded from `LOG_FORMAT`; optional, defaults to `"console"`

#### Scenario: All observability vars are set

- **WHEN** `OTEL_EXPORTER=jaeger`, `OTEL_METRICS_EXPORTER=prometheus`, and `LOG_FORMAT=json` are set
- **THEN** `cfg.Obs.OTELExporter` is `"jaeger"`, `cfg.Obs.OTELMetricsExporter` is `"prometheus"`, and `cfg.Obs.LogFormat` is `"json"`

#### Scenario: LOG_FORMAT defaults to console

- **WHEN** `LOG_FORMAT` is not set in the environment
- **THEN** `cfg.Obs.LogFormat` is `"console"`

#### Scenario: OTEL_EXPORTER missing causes startup failure

- **WHEN** `OTEL_EXPORTER` is not set
- **THEN** `config.Load()` returns a non-nil error naming the missing variable

#### Scenario: OTEL_METRICS_EXPORTER missing causes startup failure

- **WHEN** `OTEL_METRICS_EXPORTER` is not set
- **THEN** `config.Load()` returns a non-nil error naming the missing variable
