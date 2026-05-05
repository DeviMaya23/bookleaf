## Why

Traces and structured logs are in place across HTTP, DB, and R2 layers, but no OpenTelemetry metric instruments have been wired up. Without metrics, we have no basis for dashboards, alerting, or SLO tracking on request throughput, error rates, query latency, or storage operation performance.

## What Changes

- Add a request count counter and a 4xx/5xx error counter to the existing Echo `MetricsMiddleware` in `internal/observability/echo_middleware.go`
- Verify whether `otelgorm` v0.1.16 emits query duration histograms automatically; if not, add a custom GORM plugin that emits query duration and error count metrics
- Add OTel Meter instruments to `internal/storage/r2.go` for presigned URL generation duration and upload success/failure count
- Add OTel Meter instruments to `internal/usecase/image_usecase.go` for thumbnail generation duration and thumbnail success/failure count

## Capabilities

### New Capabilities

- `http-server-metrics`: Per-endpoint request count counter and error rate counters (4xx and 5xx separately) in Echo middleware
- `r2-storage-metrics`: OTel metric instruments for presigned URL generation duration, upload success/failure count, thumbnail generation duration, and thumbnail success/failure count

### Modified Capabilities

- `observability-metrics`: Extend metric instrument definitions to include the new HTTP and R2 instruments introduced by this change

## Impact

- `backend/internal/observability/echo_middleware.go` — two new metric instruments added to existing middleware
- `backend/internal/storage/r2.go` — histogram and counter instruments added to upload and presigned URL operations
- `backend/internal/usecase/image_usecase.go` — histogram and counter instruments added to thumbnail operations
- `backend/cmd/server/main.go` — may need updates if new instruments require initialisation at startup
- `backend/go.mod` — no new dependencies expected; otelgorm already present
