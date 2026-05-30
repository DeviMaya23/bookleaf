## Context

The backend uses OTel with `AlwaysSample()` and a batch exporter to GCP Cloud Trace. All spans are currently recorded and exported. The `NewTracerProvider` function in `internal/observability/tracing.go` owns sampler configuration. Config is loaded via `ObsConfig` in `internal/config/config.go`.

## Goals / Non-Goals

**Goals:**
- Reduce span export volume to approximately `OTEL_SAMPLE_RATIO` of requests (default 10%)
- Always export spans from traces that contain at least one error
- Keep sampler config centralised in `ObsConfig` with a sensible default

**Non-Goals:**
- Tail-based sampling across distributed services
- Per-route or per-user sampling rules
- Changing the local Tempo path (Tempo keeps `AlwaysSample` behaviour via ratio=1.0 or separate config)

## Decisions

### Decision: SpanProcessor filter over Sampler

OTel's `Sampler` interface runs at span start — errors are unknown at that point. A custom `SpanProcessor` wrapping the batch exporter runs at `OnEnd`, where `span.Status()` is available.

**Alternative considered:** Head-based sampler only (ratio, no error override). Rejected because error traces would be silently dropped at the configured ratio, undermining the main debugging use case.

### Decision: AlwaysSample + FilteringProcessor, not ParentBased

`ParentBased(AlwaysSample)` respects the upstream sampling flag from Cloud Run's load balancer, which sends `sampled=0` on most requests. This caused spans to be recorded but not exported. Using `AlwaysSample()` unconditionally bypasses the upstream hint — the app owns its own sampling decision.

**Alternative considered:** Adding the GCP propagator to correctly inherit Cloud Run's sampling decision. Rejected for now because Cloud Run samples at an unknown rate and we want deterministic control.

### Decision: Hash trace ID for ratio check

The `FilteringProcessor` hashes the lower 8 bytes of the trace ID against the ratio threshold — the same algorithm used by OTel's `TraceIDRatioBased`. This ensures all spans within the same trace get the same keep/drop decision (consistent traces), without needing to coordinate across spans.

### Decision: Ratio passed as parameter to NewTracerProvider

`NewTracerProvider(ctx, exporter, sampleRatio float64)` receives the ratio from `main.go` via `cfg.Obs.SampleRatio`. This keeps env var reading centralised in `config.go` and makes `NewTracerProvider` testable without env var setup.

## Risks / Trade-offs

- **Partial traces on error**: If a child span errors but the root span does not, only the error child is forwarded — the root and sibling spans are dropped, producing a headless trace in Cloud Trace. → Acceptable: error recording on spans (`span.SetStatus(codes.Error, ...)`) is already applied at every layer, so the root HTTP span will also be in error when anything below it fails.
- **Memory overhead**: `AlwaysSample` records all spans in memory before the processor drops them at export. For low-traffic prod this is negligible. → Revisit if span volume grows significantly.
- **Ratio is approximate**: TraceID hashing distributes evenly in aggregate but individual windows may vary. → No mitigation needed; approximate sampling is the expected behaviour.

## Migration Plan

1. Deploy with `OTEL_SAMPLE_RATIO` unset → defaults to `0.1`
2. Verify error traces appear in Cloud Trace after a failed request
3. Adjust ratio via env var as needed — no redeploy required beyond setting the var
4. Rollback: set `OTEL_ENABLED=false` or pin `OTEL_SAMPLE_RATIO=1.0` to restore previous behaviour
