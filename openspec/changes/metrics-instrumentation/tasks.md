## 1. Dependency

- [ ] 1.1 Add `github.com/go-gorm/opentelemetry` to `backend/go.mod` via `go get github.com/go-gorm/opentelemetry`

## 2. otelgorm Plugin Registration

- [ ] 2.1 In `cmd/server/main.go`, import `otelgorm "github.com/go-gorm/opentelemetry"` and register `db.Use(otelgorm.NewPlugin())` after `gorm.Open`

## 3. HTTP Metrics Middleware

- [ ] 3.1 In `MetricsMiddleware` in `internal/observability/echo_middleware.go`, add an `http.server.request.count` Int64Counter instrument initialised at construction time
- [ ] 3.2 Add an `http.server.request.errors` Int64Counter instrument initialised at construction time
- [ ] 3.3 In the `defer` block, increment `http.server.request.count` with `http.request.method`, `http.route`, and `http.response.status_code` attributes on every request
- [ ] 3.4 In the `defer` block, increment `http.server.request.errors` with `http.request.method`, `http.route`, and `http.status_class` (`"4xx"` or `"5xx"`) only when status code is ≥ 400

## 4. Presigned URL Duration Metrics

- [ ] 4.1 Add a `r2.presigned_url.duration` Float64Histogram field (unit: `ms`) to the `r2Storage` struct; initialise it in `NewR2Storage` using `tel.Meter`
- [ ] 4.2 In `GeneratePresignedPutURL`, record the histogram after the presign call with `r2.operation="presigned_put"` and `r2.status="success"` or `"error"`
- [ ] 4.3 In `GeneratePresignedGetURL`, record the histogram after the presign call with `r2.operation="presigned_get"` and `r2.status="success"` or `"error"`

## 5. Upload Completion Count Metric

- [ ] 5.1 Add an `r2.upload.count` Int64Counter field to the `imageUsecase` struct; initialise it in `NewImageUsecase` using `tel.Meter`
- [ ] 5.2 In `CompleteUpload`, increment the counter with `r2.status="error"` before returning on repository error, and with `r2.status="success"` before returning on success

## 6. Thumbnail Generation Metrics

- [ ] 6.1 Add a `r2.thumbnail.duration` Float64Histogram field (unit: `ms`) and an `r2.thumbnail.count` Int64Counter field to the `imageUsecase` struct; initialise both in `NewImageUsecase` using `tel.Meter`
- [ ] 6.2 In `generateThumbnail`, record `r2.thumbnail.duration` and increment `r2.thumbnail.count` with `r2.status="error"` at each early-return failure point
- [ ] 6.3 In `generateThumbnail`, record `r2.thumbnail.duration` and increment `r2.thumbnail.count` with `r2.status="success"` at the successful completion point

## 7. Unit Tests — Handler

- [ ] 7.1 Add success scenario test for `MetricsMiddleware`: assert `http.server.request.count` is incremented and `http.server.request.errors` is NOT incremented for a `2xx` response
- [ ] 7.2 Add failure scenario test for `MetricsMiddleware`: assert `http.server.request.errors` is incremented with `http.status_class="4xx"` for a `404` response

## 8. Unit Tests — Usecase

- [ ] 8.1 Add success scenario test for `CompleteUpload`: mock repo returns image, assert `r2.upload.count` is incremented with `r2.status="success"`
- [ ] 8.2 Add failure scenario test for `CompleteUpload`: mock repo returns error, assert `r2.upload.count` is incremented with `r2.status="error"`
