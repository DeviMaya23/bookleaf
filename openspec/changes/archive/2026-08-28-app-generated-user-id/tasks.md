## 1. Domain

- [x] 1.1 Update `domain.User` struct: change `ID` from `string` to `uuid.UUID` (`gorm:"type:uuid;primaryKey"`), add `IDPSubject string` (`gorm:"column:idp_subject;not null;uniqueIndex"`)

## 2. Database Migration

- [x] 2.1 Write up migration SQL: add `idp_subject` to `users` (backfill from `id`, NOT NULL, UNIQUE); add `new_id UUID DEFAULT gen_random_uuid()` to `users`; add `new_user_id UUID` staging columns to `folders`, `images`, `tags`, `pending_uploads`, `ai_categorisation_logs`; backfill staging columns via `UPDATE ... FROM users`; drop FK constraints; drop old `user_id` columns; rename `new_user_id` → `user_id`; re-add FK constraints; swap `users.id` (drop TEXT PK, rename `new_id` → `id`, add UUID PK constraint)
- [x] 2.2 Write down migration (no-op body with a comment noting a DB snapshot is the rollback path)

## 3. UserRepository Interface & Implementation

- [x] 3.1 Update `usecase.UserRepository` interface: change all `id string` parameters to `uuid.UUID`; add `GetByIDPSubject(ctx context.Context, idpSubject string) (*domain.User, error)`
- [x] 3.2 Refactor `repository.userRepository.GetOrCreate`: query by `idp_subject`, INSERT with `gen_random_uuid()` for `id` on conflict-do-nothing
- [x] 3.3 Update `repository.userRepository.GetByID` and all other methods to accept `uuid.UUID`
- [x] 3.4 Add `repository.userRepository.GetByIDPSubject` implementation

## 4. Auth Middleware

- [x] 4.1 Add `AuthenticatedUserUUIDFromContext(c echo.Context) (uuid.UUID, bool)` helper to middleware package; update `handle` to store `user.ID.String()` in context (instead of `claims.Subject`)
- [x] 4.2 Update `middleware/auth_test.go`: assert context value is a UUID string, not the Kinde subject

## 5. Handler Call Sites — UUID Parsing

- [x] 5.1 Update all handlers that call `AuthenticatedUserIDFromContext` to instead call `AuthenticatedUserUUIDFromContext` and pass `uuid.UUID` downstream (covers `me.go`, `image.go`, `image_upload.go`, `folder.go`, `tag.go`, `trash.go`, `events.go`, and any others)
- [x] 5.2 Update handler-layer usecase interfaces to accept `uuid.UUID` for userID parameters

## 6. Worker Args & Enqueuer

- [x] 6.1 Update `AccountWipeArgs` to carry `IDPSubject string` (remove `UserID`)
- [x] 6.2 Update `BookletUserDeletionArgs` to carry `IDPSubject string` (remove `UserID`)
- [x] 6.3 Update `riverEnqueuer.EnqueueAccountWipe` and `EnqueueAccountWipeUnique` to accept and pass `idpSubject string`
- [x] 6.4 Update `riverEnqueuer.EnqueueBookletUserDeletion` to accept and pass `idpSubject string`
- [x] 6.5 Update `accountJobEnqueuer` interface in `account_usecase.go` to match

## 7. Account Usecase

- [x] 7.1 Refactor `WipeAccount(ctx, idpSubject string)`: call `GetByIDPSubject` at the start; if found run full wipe using `user.IDPSubject` for Kinde API calls; if not found run Kinde-only cleanup
- [x] 7.2 Update `MarkForDeletion(ctx, userID uuid.UUID)`: accept UUID, fetch user by ID to get `IDPSubject` for job enqueue
- [x] 7.3 Update `ReconcilePendingDeletions`: enqueue `AccountWipeArgs{IDPSubject: user.IDPSubject}` for each pending-deletion user
- [x] 7.4 Update account usecase tests

## 8. User Usecase

- [x] 8.1 Update `userUsecase.GetByID` to accept `uuid.UUID`
- [x] 8.2 Update `userUsecase.UpdatePreferences` to accept `uuid.UUID`
- [x] 8.3 Update user usecase tests

## 9. Internal Handler & Share Usecase

- [x] 9.1 Add `UserResolver` interface to `internal/handler/internal.go` with `GetByIDPSubject(ctx, idpSubject string) (*domain.User, error)`; inject into `InternalHandler`
- [x] 9.2 Update `ListPublicFolders` handler: resolve Kinde subject → UUID via `UserResolver`; return empty list on `ErrUserNotFound`; pass UUID to `GetPublicFoldersByUser`
- [x] 9.3 Update `DeleteAccount` handler: resolve Kinde subject → UUID via `UserResolver`; call `MarkForDeletion` if found; enqueue `AccountWipeArgs{IDPSubject: ...}` directly if not found
- [x] 9.4 Update `InternalShareUsecase.GetPublicFoldersByUser` signature to accept `uuid.UUID`
- [x] 9.5 Update `shareUsecase.GetPublicFoldersByUser` implementation and `folderShareRepository.ListByUserID` to accept `uuid.UUID`
- [x] 9.6 Update `FolderShareRepository` interface in usecase package
- [x] 9.7 Update internal handler tests to cover resolver behaviour (found and not-found paths for both endpoints)

## 10. Frontend

- [x] 10.1 Remove `id: string` from `Me` interface in `frontend/src/features/auth/lib/me.ts`

## 11. Lint & Build

- [x] 11.1 Run `golangci-lint run ./...` from `backend/`, fix all issues
- [x] 11.2 Run `npm run build` and `npm run lint` from `frontend/`, fix all issues
