## ADDED Requirements

### Requirement: FilteringProcessor Type

The system SHALL define a `filteringProcessor` struct in `internal/observability/tracing.go` that implements `sdktrace.SpanProcessor`. It SHALL wrap a delegate `sdktrace.SpanProcessor` (the batch exporter processor) and hold a `ratio float64`.

`OnEnd` SHALL apply the following logic:
1. If `span.Status().Code == codes.Error` → forward to delegate unconditionally
2. Otherwise → hash the lower 8 bytes of `span.SpanContext().TraceID()` as a `uint64` and forward only if `hash / math.MaxUint64 < ratio`

`OnStart`, `Shutdown`, and `ForceFlush` SHALL delegate directly to the wrapped processor without modification.

#### Scenario: Error span is always forwarded

- **WHEN** `OnEnd` is called with a span whose status code is `codes.Error`
- **THEN** the span is forwarded to the delegate processor regardless of ratio

#### Scenario: Non-error span is forwarded when hash falls within ratio

- **WHEN** `OnEnd` is called with a non-error span
- **AND** the trace ID hash divided by `math.MaxUint64` is less than the configured ratio
- **THEN** the span is forwarded to the delegate processor

#### Scenario: Non-error span is dropped when hash falls outside ratio

- **WHEN** `OnEnd` is called with a non-error span
- **AND** the trace ID hash divided by `math.MaxUint64` is greater than or equal to the configured ratio
- **THEN** the span is not forwarded to the delegate processor

#### Scenario: Spans from the same trace get the same sampling decision

- **WHEN** multiple spans share the same trace ID
- **THEN** all non-error spans from that trace are either all forwarded or all dropped
