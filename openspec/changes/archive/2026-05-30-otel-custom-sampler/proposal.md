## Why

Running `AlwaysSample()` in production exports every span to GCP Cloud Trace, which increases cost and noise as traffic grows. A ratio-based sampler reduces volume while still guaranteeing that error traces are always captured for debugging.

## What Changes

- Add `SampleRatio float64` to `ObsConfig`, loaded from `OTEL_SAMPLE_RATIO` (optional, defaults to `0.1`)
- Replace `AlwaysSample()` in `NewTracerProvider` with a `FilteringProcessor` that keeps `AlwaysSample` as the SDK sampler but drops spans at export time based on ratio — unless the span has an error status, in which case it is always forwarded
- `NewTracerProvider` receives the ratio from `ObsConfig` and wires the processor

## Capabilities

### New Capabilities

- `otel-filtering-processor`: A custom `SpanProcessor` in `internal/observability/tracing.go` that wraps the batch exporter. Records all spans (via `AlwaysSample`), then at `OnEnd` decides whether to forward to the exporter: always forward if `span.Status().Code == codes.Error`, otherwise forward if `hash(traceID) < ratio`.

### Modified Capabilities

- `observability-tracing`: `NewTracerProvider` signature changes to accept a `sampleRatio float64` parameter; sampling behaviour changes from always-on to ratio + error override
- `app-config`: `ObsConfig` gains a `SampleRatio float64` field backed by `OTEL_SAMPLE_RATIO`

## Impact

- `backend/internal/observability/tracing.go` — new `FilteringProcessor` type, updated `NewTracerProvider` signature
- `backend/internal/config/config.go` — new `SampleRatio` field in `ObsConfig`
- `backend/cmd/server/main.go` — pass `cfg.Obs.SampleRatio` to `NewTracerProvider`
- No API changes, no database changes, no frontend changes
