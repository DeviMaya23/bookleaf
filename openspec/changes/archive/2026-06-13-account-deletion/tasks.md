## 1. Config & Migration

- [x] 1.1 Add `KINDE_M2M_CLIENT_ID`, `KINDE_M2M_CLIENT_SECRET`, `KINDE_M2M_TOKEN_URL`, and the Management API audience to `KindeConfig` and `loadFromEnv` (via `requireEnv`)
- [x] 1.2 Add migration `000014_add_pending_kinde_deletion_to_users` (up: add `pending_kinde_deletion BOOLEAN NOT NULL DEFAULT false`; down: drop column)
- [x] 1.3 Add `PendingKindeDeletion bool` field (`gorm:"column:pending_kinde_deletion;default:false"`) to the `User` struct in `internal/domain/user.go`

## 2. Kinde Management API Client

- [x] 2.1 Create `internal/platform/kinde` package: M2M client-credentials token fetch with in-memory caching (refetch on expiry)
- [x] 2.2 Implement `DeleteUser(ctx, kindeUserID) error` calling `DELETE /api/v1/user?id=<kindeUserID>&is_delete_profile=true`; confirm the "user not found" response shape against the dev Kinde tenant and treat it as success

## 3. Repository Layer (integration tests only)

- [x] 3.1 Image repo: add a method to list all of a user's images unscoped (including trashed), and a method to hard-delete all of a user's images
- [x] 3.2 Folder repo: add a method to clear `parent_id` on all of a user's folders, and a method to hard-delete all of a user's folders
- [x] 3.3 Tag repo: add a method to hard-delete all of a user's tags
- [x] 3.4 Pending upload repo: add a method to list all of a user's pending uploads, and a method to hard-delete all of a user's pending uploads
- [x] 3.5 User repo: add methods to set `pending_kinde_deletion = true`, hard-delete a user row, and list users with `pending_kinde_deletion = true`
- [x] 3.6 Write integration tests for all new repository methods (ownership isolation, cascade behaviour, FK ordering for the folder/image/tag/pending-upload wipe)

## 4. Account Deletion Usecase

- [x] 4.1 Define `AccountKindeDeletionArgs{UserID string}` (kind `account_kinde_deletion`, `MaxAttempts: 5`) in `job_args.go`
- [x] 4.2 Implement a new `accountUsecase.DeleteAccount(ctx, userID)`: runs the data-wipe transaction (clear folder parents → delete images incl. trashed → delete folders → delete tags → delete pending uploads → set `pending_kinde_deletion = true`), then enqueues `R2DeleteArgs` for each collected image/pending-upload path and one `AccountKindeDeletionArgs` job
- [x] 4.3 Implement `accountUsecase.ProcessAccountKindeDeletion(ctx, userID)`: calls the Kinde client's `DeleteUser`; on success or "user not found", hard-deletes the `users` row; otherwise returns an error to trigger a retry
- [x] 4.4 Implement `accountUsecase.ReconcilePendingKindeDeletions(ctx)`: lists users with `pending_kinde_deletion = true` and enqueues `AccountKindeDeletionArgs` for each
- [x] 4.5 Unit tests: `DeleteAccount` assembles the correct `R2DeleteArgs` jobs from collected paths; `ProcessAccountKindeDeletion` branches correctly on success / not-found / other-error; `ReconcilePendingKindeDeletions` enqueues one job per pending user

## 5. Handler

- [x] 5.1 Add a `DeleteMe` handler method to `MeHandler` (`DELETE /me`) calling `accountUsecase.DeleteAccount`, returning `204 No Content` on success
- [x] 5.2 Unit tests for `DeleteMe` (success path, missing auth context)

## 6. Auth Middleware Lockout

- [x] 6.1 Update the auth middleware to check `PendingKindeDeletion` on the provisioned user and return `401 Unauthorized` if true
- [x] 6.2 Unit test for the middleware lockout scenario

## 7. Worker & Periodic Job

- [x] 7.1 Implement `AccountKindeDeletionWorker` (`river.Worker[AccountKindeDeletionArgs]`) calling `accountUsecase.ProcessAccountKindeDeletion`
- [x] 7.2 Register the worker and a 24-hour periodic job calling `accountUsecase.ReconcilePendingKindeDeletions` in `main.go`, alongside the existing periodic jobs

## 8. Wiring & Routes

- [x] 8.1 Wire the Kinde client and `accountUsecase` (with image/folder/tag/pending-upload/user repos, storage enqueuer, Kinde client) into `initApp`
- [x] 8.2 Register `DELETE /me` on the protected route group

## 9. API Client (Bruno)

- [x] 9.1 Add a Bruno request for `DELETE /me`

## 10. Lint

- [x] 10.1 Run `golangci-lint` and fix any issues
