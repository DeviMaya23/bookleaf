## 1. Migration

- [ ] 1.1 Write migration `000011_pending_uploads.up.sql`: `DELETE FROM images WHERE is_uploaded = false`; `ALTER TABLE images DROP COLUMN is_uploaded`; `CREATE TABLE pending_uploads` with all fields, FK constraints, and two indexes (`user_id`, `created_at`)
- [ ] 1.2 Write migration `000011_pending_uploads.down.sql`: `DROP TABLE pending_uploads`; `ALTER TABLE images ADD COLUMN is_uploaded BOOLEAN NOT NULL DEFAULT true`

## 2. Domain Model

- [ ] 2.1 Create `internal/domain/pending_upload.go`: define `PendingUpload` struct with all fields and a `BeforeCreate` hook that assigns a new UUID when `ID` is nil
- [ ] 2.2 Remove `IsUploaded bool` field (and its GORM tag) from `Image` struct in `internal/domain/image.go`

## 3. Repository Interfaces

- [ ] 3.1 Create `internal/usecase/pending_upload_repository.go`: define `PendingUploadRepository` interface with `Create`, `GetByID`, `Delete`, and `ListStale` methods
- [ ] 3.2 Remove `ListStaleUploads` from `ImageRepository` interface in `internal/usecase/image_repository.go`

## 4. PendingUpload Repository Implementation

- [ ] 4.1 Create `internal/repository/pending_upload_repository.go`: implement all four `PendingUploadRepository` methods; add compile-time interface check `var _ usecase.PendingUploadRepository = (*pendingUploadRepository)(nil)`

## 5. Image Repository Updates

- [ ] 5.1 Remove `is_uploaded = true` filter from `List` query in `internal/repository/image_repository.go`
- [ ] 5.2 Remove `ListStaleUploads` method from `internal/repository/image_repository.go`

## 6. Usecase: UploadInitResult Shape

- [ ] 6.1 Update `UploadInitResult` in `internal/usecase/image_usecase.go` to carry flat fields (`ID uuid.UUID`, `UploadURL string`, `R2Path string`) instead of `Image *domain.Image`, so the handler is decoupled from the domain type; update the handler in `internal/handler/image.go` to use the flat fields

## 7. Usecase: InitiateUpload

- [ ] 7.1 Update `InitiateUpload` to call `pendingUploadRepo.Create` with a `domain.PendingUpload` (all metadata fields + resolved `FolderID`); remove the `imageRepo.Create` and `imageRepo.SetImageFolder` calls; return the flat `UploadInitResult`

## 8. Usecase: CompleteUpload

- [ ] 8.1 Update `CompleteUpload` to fetch the pending record via `pendingUploadRepo.GetByID(ctx, id, userID)`; pass `pendingUpload.R2Path` to `prepareThumbnail`
- [ ] 8.2 Wrap the commit in a transaction: `imageRepo.Create` (with all fields from `pendingUpload` plus thumbnail metadata), then `imageRepo.SetImageFolder` if `pendingUpload.FolderID != nil`, then `pendingUploadRepo.Delete`

## 9. Usecase: CleanupStaleUploads

- [ ] 9.1 Update `CleanupStaleUploads` to call `pendingUploadRepo.ListStale(ctx, now-threshold)`; for each record: delete R2 object (best-effort), call `pendingUploadRepo.Delete(ctx, record.ID)`; remove all references to `imageRepo.ListStaleUploads`

## 10. Wire Up

- [ ] 10.1 In `cmd/server/main.go`: instantiate `repository.NewPendingUploadRepository(db)` and pass it as a new argument to `usecase.NewImageUsecase`; update `NewImageUsecase` signature and struct accordingly

## 11. Unit Tests — Usecase

- [ ] 11.1 Add `mockPendingUploadRepository` to `image_usecase_test.go` implementing all four interface methods (with configurable return values for success/failure scenarios)
- [ ] 11.2 Update `TestImageUsecase_InitiateUpload`: replace `imageRepo.Create` mock expectations with `pendingUploadRepo.Create` expectations; one success scenario, one failure scenario
- [ ] 11.3 Update `TestImageUsecase_CompleteUpload`: replace `imageRepo.GetByID` setup with `pendingUploadRepo.GetByID`; add `pendingUploadRepo.Delete` expectation; one success scenario, one failure scenario (pending not found)
- [ ] 11.4 Update `TestImageUsecase_CleanupStaleUploads`: replace `imageRepo.ListStaleUploads` mock with `pendingUploadRepo.ListStale`; add `pendingUploadRepo.Delete` expectations; one success scenario, one no-op scenario

## 12. Integration Tests — PendingUpload Repository

- [ ] 12.1 Create `internal/repository/pending_upload_repository_integration_test.go`
- [ ] 12.2 Add test for `Create`: success — row exists in `pending_uploads` with correct fields
- [ ] 12.3 Add test for `GetByID`: success (returns correct row); failure (ID not found returns error)
- [ ] 12.4 Add test for `Delete`: success — row is removed from `pending_uploads`
- [ ] 12.5 Add test for `ListStale`: returns rows older than threshold; excludes rows newer than threshold

## 13. Integration Tests — Image Repository

- [ ] 13.1 Update `image_repository_integration_test.go`: remove any `IsUploaded: true` fields from `newTestImage` or test setup (field no longer exists on `domain.Image`)
- [ ] 13.2 Remove the `ListStaleUploads` integration test if one exists

## 14. Bruno

- [ ] 14.1 Verify `bruno/images/initiate-upload.bru` request shape is unchanged (no edits needed; confirm only)
