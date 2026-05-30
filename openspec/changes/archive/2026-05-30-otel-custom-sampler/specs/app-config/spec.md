## MODIFIED Requirements

### Requirement: Observability Config Sub-Struct

The `Config` struct SHALL include an `Obs ObsConfig` field. `ObsConfig` SHALL have:

- `OTELEnabled bool` — loaded from `OTEL_ENABLED`; optional, defaults to `false`
- `OTELExporter string` — loaded from `OTEL_EXPORTER`; **conditionally required**: only validated as non-empty when `OTELEnabled` is `true`
- `OTELMetricsExporter string` — loaded from `OTEL_METRICS_EXPORTER`; **conditionally required**: only validated as non-empty when `OTELEnabled` is `true`
- `LogFormat string` — loaded from `LOG_FORMAT`; optional, defaults to `"json"`
- `SampleRatio float64` — loaded from `OTEL_SAMPLE_RATIO`; optional, defaults to `0.1`

When `OTELEnabled` is `false`, `OTEL_EXPORTER`, `OTEL_METRICS_EXPORTER`, and `OTEL_SAMPLE_RATIO` SHALL be loaded with their defaults without error, even if unset.

#### Scenario: All observability vars are set with OTel enabled

- **WHEN** `OTEL_ENABLED=true`, `OTEL_EXPORTER=gcp`, `OTEL_METRICS_EXPORTER=gcp`, `LOG_FORMAT=json`, and `OTEL_SAMPLE_RATIO=0.2` are set
- **THEN** `cfg.Obs.OTELEnabled` is `true`, `cfg.Obs.OTELExporter` is `"gcp"`, `cfg.Obs.OTELMetricsExporter` is `"gcp"`, `cfg.Obs.LogFormat` is `"json"`, and `cfg.Obs.SampleRatio` is `0.2`

#### Scenario: OTEL_SAMPLE_RATIO defaults to 0.1

- **WHEN** `OTEL_SAMPLE_RATIO` is not set
- **THEN** `cfg.Obs.SampleRatio` is `0.1`

#### Scenario: LOG_FORMAT defaults to json

- **WHEN** `LOG_FORMAT` is not set
- **THEN** `cfg.Obs.LogFormat` is `"json"`

#### Scenario: OTEL_ENABLED defaults to false

- **WHEN** `OTEL_ENABLED` is not set
- **THEN** `cfg.Obs.OTELEnabled` is `false`

#### Scenario: OTEL_EXPORTER missing is not an error when OTel disabled

- **WHEN** `OTEL_ENABLED` is not set (or `false`)
- **AND** `OTEL_EXPORTER` is not set
- **THEN** `config.Load()` returns a non-nil `*Config` with a nil error

#### Scenario: OTEL_METRICS_EXPORTER missing is not an error when OTel disabled

- **WHEN** `OTEL_ENABLED` is not set (or `false`)
- **AND** `OTEL_METRICS_EXPORTER` is not set
- **THEN** `config.Load()` returns a non-nil `*Config` with a nil error

#### Scenario: OTEL_EXPORTER missing causes startup failure when OTel enabled

- **WHEN** `OTEL_ENABLED=true`
- **AND** `OTEL_EXPORTER` is not set
- **THEN** `config.Load()` returns a non-nil error naming the missing variable

#### Scenario: OTEL_METRICS_EXPORTER missing causes startup failure when OTel enabled

- **WHEN** `OTEL_ENABLED=true`
- **AND** `OTEL_METRICS_EXPORTER` is not set
- **THEN** `config.Load()` returns a non-nil error naming the missing variable
