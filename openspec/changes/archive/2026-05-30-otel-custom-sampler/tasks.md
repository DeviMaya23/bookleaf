## 1. Config

- [x] 1.1 Add `SampleRatio float64` field to `ObsConfig` in `internal/config/config.go`, loaded from `OTEL_SAMPLE_RATIO` with default `0.1`
- [x] 1.2 Add unit tests for `SampleRatio` default and explicit value in `internal/config/config_test.go`

## 2. FilteringProcessor

- [x] 2.1 Implement `filteringProcessor` struct in `internal/observability/tracing.go` wrapping a delegate `sdktrace.SpanProcessor` with a `ratio float64`
- [x] 2.2 Implement `OnEnd`: always forward if `span.Status().Code == codes.Error`; otherwise hash lower 8 bytes of trace ID and forward if `hash / math.MaxUint64 < ratio`
- [x] 2.3 Implement `OnStart`, `Shutdown`, `ForceFlush` as pass-throughs to delegate
- [x] 2.4 Write unit tests for `filteringProcessor`: error span always forwarded, non-error span forwarded within ratio, non-error span dropped outside ratio, consistent decision across same trace ID

## 3. TracerProvider Wiring

- [x] 3.1 Update `NewTracerProvider` signature to `(ctx context.Context, exporter string, sampleRatio float64)`
- [x] 3.2 Replace `sdktrace.WithBatcher(exp)` with `sdktrace.WithSpanProcessor(filteringProcessor wrapping sdktrace.NewBatchSpanProcessor(exp))`
- [x] 3.3 Update `cmd/server/main.go` to pass `cfg.Obs.SampleRatio` to `NewTracerProvider`
