## 1. Config

- [x] 1.1 Add `BookletConfig` struct (`BaseURL string`, `InternalSecret string`) to `internal/platform/config/config.go`; read from env vars `BOOKLET_BASE_URL` and `BOOKLET_INTERNAL_SECRET`

## 2. Booklet HTTP Client

- [x] 2.1 Define `BookletClient` interface in `internal/usecase/` with method `DeleteUser(ctx context.Context, userID string) error`
- [x] 2.2 Create `internal/booklet/booklet.go` — `Client` struct wrapping `*http.Client`, `baseURL`, and `internalSecret`; `NewClient(baseURL, secret string) *Client` constructor; `DeleteUser` calls `DELETE <baseURL>/internal/users/<userID>` with `X-Booklet-Internal-Secret` header; returns error on non-2xx

## 3. Job Args

- [x] 3.1 Add `BookletUserDeletionArgs` to `internal/usecase/job_args.go` (kind: `booklet_user_deletion`, `MaxAttempts: 10`, field: `UserID string`)
- [x] 3.2 Add `DeleteAccountArgs` to `internal/usecase/job_args.go` (kind: `delete_account`, `MaxAttempts: 5`, field: `UserID string`)

## 4. Usecase

- [x] 4.1 Add `bookletClient BookletClient` field to `accountUsecase`; update `NewAccountUsecase` to accept it as the last parameter (nil-safe — no-op when nil)
- [x] 4.2 In `DeleteAccount`, after enqueuing `AccountKindeDeletionArgs`, enqueue `BookletUserDeletionArgs` if `bookletClient` is non-nil; log a warn and continue on enqueue failure (same pattern as R2 and Kinde enqueues)
- [x] 4.3 Add `ProcessBookletUserDeletion(ctx context.Context, userID string) error` to `accountUsecase` — calls `bookletClient.DeleteUser(ctx, userID)`
- [x] 4.4 Add `ScheduleAccountDeletion(ctx context.Context, userID string) error` to `accountUsecase` — looks up user by ID; if `pending_kinde_deletion` is true, return nil (no-op); if user exists and not pending, enqueue `DeleteAccountArgs` and return nil; if user does not exist, enqueue `AccountKindeDeletionArgs` (Kinde-only) and return nil

## 5. Workers

- [x] 5.1 Create `internal/worker/booklet_user_deletion.go` — `BookletUserDeletionWorker` calls `accountUsecase.ProcessBookletUserDeletion(ctx, job.Args.UserID)`
- [x] 5.2 Create `internal/worker/delete_account.go` — `DeleteAccountWorker` calls `accountUsecase.DeleteAccount(ctx, job.Args.UserID)`

## 6. Internal Handler

- [x] 6.1 Add `InternalAccountUsecase` interface to `internal/handler/internal.go` with methods `ScheduleAccountDeletion(ctx context.Context, userID string) error`
- [x] 6.2 Add `accountUsecase InternalAccountUsecase` field to `InternalHandler`; update `NewInternalHandler` to accept it
- [x] 6.3 Implement `(h *InternalHandler) DeleteAccount(c echo.Context) error` — parse `:id` param, call `ScheduleAccountDeletion`; return 202 on nil (always — no 404 case)

## 7. Wiring

- [x] 7.1 In `cmd/server/main.go`, instantiate `booklet.NewClient(cfg.Booklet.BaseURL, cfg.Booklet.InternalSecret)` only when `cfg.Booklet.BaseURL` is non-empty (else pass nil); pass to `NewAccountUsecase`
- [x] 7.2 Register `worker.NewBookletUserDeletionWorker` and `worker.NewDeleteAccountWorker` with River
- [x] 7.3 Register `DELETE /internal/accounts/:id` on the existing `internalGroup`; update `NewInternalHandler` call to pass `accountUsecase`

## 8. Unit Tests

- [x] 8.1 `account_usecase_test.go` — `TestAccountUsecase_DeleteAccount_EnqueuesBookletJobWhenClientConfigured`: assert `BookletUserDeletionArgs` is enqueued when `bookletClient` is set
- [x] 8.2 `account_usecase_test.go` — `TestAccountUsecase_DeleteAccount_SkipsBookletJobWhenClientNil`: assert `DeleteAccount` returns nil and no `BookletUserDeletionArgs` is enqueued when `bookletClient` is nil
- [x] 8.3 `account_usecase_test.go` — `TestAccountUsecase_ProcessBookletUserDeletion_SuccessReturnsNil`: assert nil is returned when the client call succeeds
- [x] 8.4 `account_usecase_test.go` — `TestAccountUsecase_ProcessBookletUserDeletion_ClientErrorPropagates`: assert the client error is returned
- [x] 8.5 `account_usecase_test.go` — `TestAccountUsecase_ScheduleAccountDeletion_EnqueuesJobForActiveUser`: assert `DeleteAccountArgs` is enqueued and nil is returned for a user with `pending_kinde_deletion = false`
- [x] 8.6 `account_usecase_test.go` — `TestAccountUsecase_ScheduleAccountDeletion_SkipsEnqueueForPendingUser`: assert no job is enqueued and nil is returned when `pending_kinde_deletion = true`
- [x] 8.7 `account_usecase_test.go` — `TestAccountUsecase_ScheduleAccountDeletion_EnqueuesKindeDeletionForUnprovisionedUser`: assert `AccountKindeDeletionArgs` is enqueued and nil is returned when user does not exist
- [x] 8.8 `handler/internal_test.go` — `TestInternalHandler_DeleteAccount_ReturnsAcceptedAndEnqueuesForActiveUser`: assert 202 and `ScheduleAccountDeletion` is called
- [x] 8.9 `handler/internal_test.go` — `TestInternalHandler_DeleteAccount_ReturnsAcceptedForPendingUser`: assert 202 when `ScheduleAccountDeletion` returns nil (already pending)
- [x] 8.10 `handler/internal_test.go` — `TestInternalHandler_DeleteAccount_ReturnsAcceptedForUnprovisionedUser`: assert 202 when `ScheduleAccountDeletion` returns nil for an unprovisioned user

## 9. Bruno

- [x] 9.1 Create `bruno/internal/delete-account.bru` for `DELETE /internal/accounts/:id`

## 10. Lint

- [x] 10.1 Run `golangci-lint run ./...` from `backend/` and fix any issues
