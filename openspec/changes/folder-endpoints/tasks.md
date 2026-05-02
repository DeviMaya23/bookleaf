## 1. Domain & Migration

- [ ] 1.1 Remove `DeletedAt gorm.DeletedAt` field and its GORM tag from `internal/domain/folder.go`
- [ ] 1.2 Create `backend/migration/000005_remove_folders_soft_delete.up.sql` — drop `deleted_at` column and `idx_folders_deleted_at` index from `folders`; drop and recreate `fk_images_folder` on `images.folder_id` as `ON DELETE RESTRICT`
- [ ] 1.3 Create `backend/migration/000005_remove_folders_soft_delete.down.sql` — restore `deleted_at` column and index; restore `fk_images_folder` as `ON DELETE SET NULL`
- [ ] 1.4 Run `make migrate-up` locally and verify no errors

## 2. Repository

- [ ] 2.1 Create `internal/usecase/folder_repository.go` — define `FolderRepository` interface with `Create`, `List`, `GetByID`, `Update`, `DeleteWithCascade` methods
- [ ] 2.2 Create `internal/repository/folder_repository.go` — implement `FolderRepository`; all queries scoped by `userID`
- [ ] 2.3 Implement `Create` — insert folder row, return created folder
- [ ] 2.4 Implement `List` — select all folders for a user ordered by `created_at`
- [ ] 2.5 Implement `GetByID` — select single folder by `id` and `user_id`; return error on not found
- [ ] 2.6 Implement `Update` — update `name` and `parent_id` by `id` and `user_id`; return updated folder
- [ ] 2.7 Implement `DeleteWithCascade` — in a single transaction: `UPDATE folders SET parent_id = NULL WHERE parent_id = $id`, `UPDATE images SET folder_id = NULL WHERE folder_id = $id`, `DELETE FROM folders WHERE id = $id AND user_id = $userID`

## 3. Usecase

- [ ] 3.1 Create `internal/usecase/folder_usecase.go` — define `FolderUsecase` interface and `folderUsecase` struct
- [ ] 3.2 Implement `Create` usecase method — validate name is non-empty, delegate to repository
- [ ] 3.3 Implement `List` usecase method — delegate to repository
- [ ] 3.4 Implement `GetByID` usecase method — delegate to repository, propagate not-found
- [ ] 3.5 Implement `Update` usecase method — validate name is non-empty, delegate to repository, propagate not-found
- [ ] 3.6 Implement `Delete` usecase method — delegate to `DeleteWithCascade` repository method, propagate not-found

## 4. Handler

- [ ] 4.1 Create `internal/handler/folder.go` — define `FolderHandler` struct and constructor
- [ ] 4.2 Implement `CreateFolder` handler — parse body, extract userID from context, call usecase, return 201
- [ ] 4.3 Implement `ListFolders` handler — extract userID from context, call usecase, return 200
- [ ] 4.4 Implement `GetFolder` handler — parse `:id` param, extract userID, call usecase, return 200 or 404
- [ ] 4.5 Implement `UpdateFolder` handler — parse `:id` and body, extract userID, call usecase, return 200 or 404
- [ ] 4.6 Implement `DeleteFolder` handler — parse `:id`, extract userID, call usecase, return 204 or 404

## 5. Routing

- [ ] 5.1 Wire `FolderRepository`, `FolderUsecase`, and `FolderHandler` in `cmd/server/main.go`
- [ ] 5.2 Register `POST /folders`, `GET /folders`, `GET /folders/:id`, `PUT /folders/:id`, `DELETE /folders/:id` on the protected route group

## 6. Tests

- [ ] 6.1 Create `internal/usecase/folder_usecase_test.go` — success + failure scenario for each usecase method (mock `FolderRepository`)
- [ ] 6.2 Create `internal/handler/folder_test.go` — success + failure scenario for each handler method (mock `FolderUsecase`)
- [ ] 6.3 Create `internal/repository/folder_repository_integration_test.go` — integration tests using Testcontainers; success + failure scenario for each repository method against a real PostgreSQL container
