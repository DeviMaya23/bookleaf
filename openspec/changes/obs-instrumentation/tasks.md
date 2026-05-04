## 1. Telemetry Foundation

- [ ] 1.1 Create `internal/observability/telemetry.go` — define `Telemetry` struct with `Logger *zap.Logger`, `Tracer trace.Tracer`, `Meter metric.Meter` fields; implement `NewTelemetry(logger, tracer, meter)` and `NewNoopTelemetry()` (using `zap.NewNop()`, `noop.NewTracerProvider().Tracer("")`, `noop.NewMeterProvider().Meter("")`)
- [ ] 1.2 Add `LoggingMiddleware(tel *Telemetry) echo.MiddlewareFunc` to `internal/observability/echo_middleware.go` — after handler returns, call `LoggerFromContext(ctx, tel.Logger)` and emit one INFO log with fields: `user_id` (from `middleware.AuthenticatedUserIDFromContext`), `http.request.method`, `http.route` (matched route pattern), `http.response.status_code`, `duration_ms`, `trace_id`
- [ ] 1.3 Write unit tests for `LoggingMiddleware` — success scenario (200 response, all fields present) and failure scenario (500 response, correct `http.response.status_code` logged)

## 2. Auth Middleware

- [ ] 2.1 Add `logger *zap.Logger` field to `authMiddleware` struct in `internal/middleware/auth.go`; update `NewAuthMiddleware` and `newAuthMiddlewareWithStorage` to accept and store `*zap.Logger`
- [ ] 2.2 In `authMiddleware.handle`, after each rejection branch, emit a WARN log using `observability.LoggerFromContext(ctx, m.logger)` with fields: `event: "auth.token_rejected"`, `reason` (e.g. `"missing_header"`, `"invalid_token"`)
- [ ] 2.3 Update `middleware/auth_test.go` — pass `zap.NewNop()` to `newAuthMiddlewareWithStorage` in all test cases

## 3. R2 Storage Instrumentation

- [ ] 3.1 Add `tel *observability.Telemetry` field to `r2Storage`; update `NewR2Storage(cfg config.R2Config, tel *observability.Telemetry) StorageService`
- [ ] 3.2 In `GeneratePresignedPutURL`: add child span `storage.GeneratePresignedPutURL` with `defer span.End()`; on success emit INFO `event: "r2.presigned_put.success"`, `image_id`; on error record span error and emit ERROR `event: "r2.presigned_put.failed"`, `image_id`, `error`
- [ ] 3.3 In `GeneratePresignedGetURL`: same pattern — span `storage.GeneratePresignedGetURL`; INFO `event: "r2.presigned_get.success"` or ERROR `event: "r2.presigned_get.failed"`, `image_id`, `error`
- [ ] 3.4 In `PutObject`: add child span `storage.PutObject` with `defer span.End()`; record error on span if it fails
- [ ] 3.5 In `GetObject`: add child span `storage.GetObject` with `defer span.End()`; record error on span if it fails

## 4. User Usecase Instrumentation

- [ ] 4.1 Add `tel *observability.Telemetry` to `userUsecase` struct; update `NewUserUsecase(userRepo UserRepository, tel *observability.Telemetry) UserUsecase`
- [ ] 4.2 Add child span `usecase.GetOrCreateUser` (or per-method name) with `defer span.End()` to each public method; propagate updated `ctx` to repo calls; record errors on span
- [ ] 4.3 After a new user record is successfully persisted, emit INFO log: `event: "user.created"`, `user_id`

## 5. Image Usecase Instrumentation

- [ ] 5.1 Add `tel *observability.Telemetry` to `imageUsecase` struct; update `NewImageUsecase(imageRepo ImageRepository, store StorageService, thumbnails ThumbnailService, tel *observability.Telemetry) ImageUsecase`
- [ ] 5.2 Add child span to each public method (`usecase.InitiateUpload`, `usecase.CompleteUpload`, `usecase.ListImages`, `usecase.GetImage`, `usecase.SoftDelete`, `usecase.ListTrashed`, `usecase.Restore`, `usecase.UpdateImage`, `usecase.MoveToFolder`) with `defer span.End()`; propagate `ctx`; record errors on span
- [ ] 5.3 In `InitiateUpload`, after image record is created: emit INFO `event: "r2.upload.started"`, `image_id`, `user_id`, `mime_type`, `file_size`, `r2_key`
- [ ] 5.4 In `CompleteUpload`, on success: emit INFO `event: "r2.upload.completed"`, `image_id`, `user_id`, `duration_ms`
- [ ] 5.5 In thumbnail processing path: emit INFO `event: "thumbnail.job.started"` (`image_id`, `user_id`) before job runs; INFO `event: "thumbnail.job.completed"` (`image_id`, `user_id`, `duration_ms`) on success; ERROR `event: "thumbnail.job.failed"` (`image_id`, `user_id`, `error`) on failure
- [ ] 5.6 In `UpdateImage`/edit path: emit INFO `event: "image.mutated"`, `image_id`, `user_id`, `operation: "edited"`
- [ ] 5.7 In `MoveToFolder`: emit INFO `event: "image.mutated"`, `image_id`, `user_id`, `operation: "moved_to_folder"`, `folder_id`
- [ ] 5.8 In `SoftDelete`: emit INFO `event: "image.mutated"`, `image_id`, `user_id`, `operation: "trashed"`

## 6. Folder Usecase Instrumentation

- [ ] 6.1 Add `tel *observability.Telemetry` to `folderUsecase` struct; update `NewFolderUsecase(folderRepo FolderRepository, tel *observability.Telemetry) FolderUsecase`
- [ ] 6.2 Add child span to each public method (`usecase.CreateFolder`, `usecase.ListFolders`, `usecase.GetFolder`, `usecase.UpdateFolder`, `usecase.DeleteFolder`) with `defer span.End()`; propagate `ctx`; record errors on span
- [ ] 6.3 In `DeleteFolder`, before or after deletion: emit INFO `event: "folder.mutated"`, `folder_id`, `user_id`, `operation: "deleted"`, `image_count` (number of images associated with the folder at deletion time)

## 7. Image Handler Instrumentation

- [ ] 7.1 Add `tel *observability.Telemetry` to `ImageHandler`; update `NewImageHandler(imageUsecase ImageUsecase, store StorageService, tel *observability.Telemetry) *ImageHandler`
- [ ] 7.2 Add child span to each handler method (`handler.InitiateUpload`, `handler.CompleteUpload`, `handler.ListImages`, `handler.GetImage`, `handler.SoftDelete`, `handler.ListTrashed`, `handler.Restore`) with `defer span.End()`; propagate `ctx` to usecase calls; record errors on span
- [ ] 7.3 Where a usecase returns an authorisation error (user accessing another user's resource), emit WARN log before returning: `event: "auth.unauthorized_access"`, `user_id`, `resource_id` (the requested image ID)

## 8. Folder Handler Instrumentation

- [ ] 8.1 Add `tel *observability.Telemetry` to `FolderHandler`; update `NewFolderHandler(folderUsecase FolderUsecase, tel *observability.Telemetry) *FolderHandler`
- [ ] 8.2 Add child span to each handler method (`handler.CreateFolder`, `handler.ListFolders`, `handler.GetFolder`, `handler.UpdateFolder`, `handler.DeleteFolder`) with `defer span.End()`; propagate `ctx`; record errors on span
- [ ] 8.3 Where a usecase returns an authorisation error, emit WARN log: `event: "auth.unauthorized_access"`, `user_id`, `resource_id` (the requested folder ID)

## 9. Me Handler Instrumentation

- [ ] 9.1 Add `tel *observability.Telemetry` to `MeHandler`; update `NewMeHandler(userUsecase UserUsecase, tel *observability.Telemetry) *MeHandler`
- [ ] 9.2 Add child span `handler.GetMe` with `defer span.End()` to `GetMe`; propagate `ctx`; record error on span if usecase fails

## 10. Main Wiring

- [ ] 10.1 In `cmd/server/main.go`, after `NewTracerProvider` and `NewMeterProvider` return, construct `tel := observability.NewTelemetry(logger, otel.Tracer("bookleaf"), otel.Meter("bookleaf"))`
- [ ] 10.2 Pass `tel` to all updated constructors: `NewR2Storage`, `NewImageUsecase`, `NewFolderUsecase`, `NewUserUsecase`, `NewImageHandler`, `NewFolderHandler`, `NewMeHandler`
- [ ] 10.3 Pass `logger` to `authmiddleware.NewAuthMiddleware`
- [ ] 10.4 Register `observability.LoggingMiddleware(tel)` on the `protected` route group (after `protected.Use(authMiddleware)`)

## 11. Unit Tests

- [ ] 11.1 Update `handler/image_test.go`, `handler/folder_test.go`, `handler/me_test.go` — pass `observability.NewNoopTelemetry()` to updated handler constructors
- [ ] 11.2 Update `usecase/image_usecase_test.go`, `usecase/folder_usecase_test.go`, `usecase/user_usecase_test.go` — pass `observability.NewNoopTelemetry()` to updated usecase constructors
- [ ] 11.3 Update `middleware/auth_test.go` — pass `zap.NewNop()` to `newAuthMiddlewareWithStorage` in all existing test cases
