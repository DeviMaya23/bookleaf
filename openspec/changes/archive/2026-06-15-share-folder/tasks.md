## 1. Migration and domain

- [x] 1.1 Add migration `000015_create_folder_shares` (up/down) creating the `folder_shares` table per `design.md` (`id`, `folder_id` unique FK with `ON DELETE CASCADE`, `token` unique, `created_at`)
- [x] 1.2 Add `domain.FolderShare{ID, FolderID, Token, CreatedAt, Folder Folder}` (with `BeforeCreate` UUID generation, mirroring `domain.Folder`)

## 2. FolderShareRepository

- [x] 2.1 Define `usecase.FolderShareRepository` interface (`Create`, `GetByFolderID`, `GetByToken` with `Folder` preload, `DeleteByFolderID`) per `specs/folder-sharing/spec.md`
- [x] 2.2 Implement `repository/folder_share_repository.go` (GORM), `NewFolderShareRepository(db) usecase.FolderShareRepository`
- [x] 2.3 Integration tests for `folder_share_repository.go`: `GetByFolderID` not-found, `GetByToken` preloads folder / not-found, `DeleteByFolderID` idempotent, unique constraint on `folder_id`, cascade delete when folder is deleted

## 3. ShareUsecase

- [x] 3.1 Define `usecase.ShareFolderRepository` and `usecase.ShareImageRepository` narrow interfaces per `specs/folder-sharing/spec.md`
- [x] 3.2 Implement `shareUsecase` struct + `NewShareUsecase(folderShareRepo, shareFolderRepo, shareImageRepo, store, tel)`
- [x] 3.3 Implement `CreateShare`: ownership check, idempotent get-or-create, token generation (`crypto/rand` + `base64.RawURLEncoding`), unique-violation fallback to `GetByFolderID`
- [x] 3.4 Implement `GetShare`: ownership check + `GetByFolderID`
- [x] 3.5 Implement `DeleteShare`: ownership check + `DeleteByFolderID`
- [x] 3.6 Implement `GetSharedFolder`: resolve token, list folder images via `ShareImageRepository.ListByFolder(ctx, share.Folder.UserID, share.FolderID, nil, nil)`, build `SharedFolder`/`SharedImage` with presigned thumbnail/full-res URLs (reuse `presignedGetTTL`)
- [x] 3.7 Unit tests for `shareUsecase` per `specs/folder-sharing/spec.md`: `CreateShare` (new/idempotent/concurrent-fallback/not-owned), `GetShare` (found/not-shared/not-owned), `DeleteShare` (revoke/no-op/not-owned), `GetSharedFolder` (assembly with presigned URLs, nil thumbnail, unknown token, empty folder) — fakes for repositories, value-return spy for `StorageService`

## 4. ShareHandler, routes, and Bruno

- [x] 4.1 Implement `ShareHandler` with `CreateShare`, `GetShare`, `DeleteShare` (parse UUID, extract `userID`, map `gorm.ErrRecordNotFound` → 404) and `GetSharedFolder` (public, maps not-found → 404)
- [x] 4.2 Register routes in `cmd/server/main.go`: `protected.POST/GET/DELETE("/folders/:id/share", ...)` and `e.GET("/share/:token", ...)` (outside `protected`)
- [x] 4.3 Handler unit tests per `specs/folder-sharing/spec.md`: 201 vs 200 on create, 404/400/401 mappings for owner endpoints, public endpoint 200 body shape and 404 for unknown token — value-return spy `ShareUsecase`
- [x] 4.4 Add Bruno requests: `bruno/folders/create-share.bru` (POST), `bruno/folders/get-share.bru` (GET), `bruno/folders/delete-share.bru` (DELETE), `bruno/share/get-shared-folder.bru` (public GET)

## 5. Verification

- [x] 5.1 Run `golangci-lint run` in `backend/` and fix any issues
