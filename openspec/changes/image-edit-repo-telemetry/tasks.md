## 1. Dependency

- [ ] 1.1 Add `github.com/go-gorm/opentelemetry` to `backend/go.mod` via `go get`

## 2. Zap Bridge for GORM Logger

- [ ] 2.1 Create `backend/internal/repository/gorm_logger.go` — define `zapGORMLogger` struct implementing `logger.Interface`
- [ ] 2.2 Implement level mapping: `Info` → `zap.Debug`, `Warn` → `zap.Warn`, `Error` → `zap.Error`
- [ ] 2.3 Implement `Trace` method: call `LoggerFromContext(ctx, base)` to inject `trace_id`; emit `zap.Warn` for slow queries and `zap.Error` for errors with `elapsed_ms`, `rows_affected`, `sql` fields; default slow-query threshold 200ms

## 3. GORM Plugin Registration

- [ ] 3.1 In `cmd/server/main.go`, after `gorm.Open`, register `otelgorm.NewPlugin(otelgorm.WithLogger(zapLogger))` via `db.Use(...)`

## 4. UpdateImage Repository Method

- [ ] 4.1 Add `Update(ctx context.Context, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error)` to the `ImageRepository` interface in `internal/usecase/`
- [ ] 4.2 Implement `Update` in `internal/repository/image_repository.go` using `db.Model(&Image{}).Where("id = ? AND user_id = ?", id, userID).Updates(fields)`; treat `RowsAffected == 0` as `gorm.ErrRecordNotFound`; return the updated record

## 5. UpdateImage Usecase Method

- [ ] 5.1 Define `UpdateImageParams` struct in `internal/usecase/` with `Title *string` and `FolderID **uuid.UUID`
- [ ] 5.2 Add `UpdateImage(ctx, id uuid.UUID, userID string, params UpdateImageParams) (*domain.Image, error)` to the `ImageUsecase` interface
- [ ] 5.3 Implement `UpdateImage` in `internal/usecase/image_usecase.go`: fetch existing image, build `fields` map from non-nil params only, delegate to `repo.Update`
- [ ] 5.4 Add conditional `image.mutated / moved_to_folder` log: emit only when `FolderID` is present in params AND the new value differs from the current image's `folder_id`; use `LoggerFromContext(ctx, tel.Logger)`

## 6. UpdateImage Handler

- [ ] 6.1 Define `updateImageRequest` struct in `internal/handler/image.go` with presence tracking for `title` and `folder_id` (use `*string` and a custom JSON decoder or `json.RawMessage` wrapper to distinguish absent from null `folder_id`)
- [ ] 6.2 Implement `UpdateImage` handler method: bind request, validate non-empty title if provided, map to `UpdateImageParams`, call `uc.UpdateImage`, return 200 with updated image or appropriate error status
- [ ] 6.3 Register `PATCH /images/:id` route in `cmd/server/main.go` on the protected group

## 7. Unit Tests — Handler

- [ ] 7.1 Add success scenario test for `UpdateImage` handler: mock usecase returns updated image, assert `200 OK` and response body
- [ ] 7.2 Add failure scenario test for `UpdateImage` handler: mock usecase returns `gorm.ErrRecordNotFound`, assert `404 Not Found`

## 8. Unit Tests — Usecase

- [ ] 8.1 Add success scenario test for `UpdateImage` usecase: mock repo returns updated image, assert returned image reflects the update
- [ ] 8.2 Add failure scenario test for `UpdateImage` usecase: mock repo returns error, assert error is propagated

## 9. Integration Test — Repository

- [ ] 9.1 Add integration test for `Update` repository method: assert selective field update does not overwrite unrelated fields (e.g. `thumbnail_path` unchanged after title-only update)
- [ ] 9.2 Add integration test for `Update` with non-existent ID: assert `gorm.ErrRecordNotFound` is returned
