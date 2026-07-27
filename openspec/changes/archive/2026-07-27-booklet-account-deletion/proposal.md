## Why

Bookleaf and Booklet share the same Kinde ecosystem. When an account is deleted, both apps must purge their own data — but neither app currently knows about the other. This change makes Bookleaf the orchestrator of cross-app account deletion, ensuring both sides clean up regardless of which app the user initiated deletion from.

## What Changes

- **New `internal/booklet` HTTP client** — calls Booklet's `DELETE /internal/users/:id` endpoint
- **New `BookletUserDeletionArgs` job** — notifies Booklet to purge its own data as part of Bookleaf's deletion flow (max 10 attempts, exponential backoff)
- **New `DeleteAccountArgs` job** — async wrapper that calls `DeleteAccount`; used when Booklet triggers deletion on Bookleaf's side
- **New `DELETE /internal/accounts/:id` internal endpoint on Bookleaf** — accepts a signal from Booklet that a Bookverse account should be deleted; returns `202 Accepted` if the account is found (and enqueues `DeleteAccountArgs`), `404 Not Found` if unknown; idempotency-gated via `pending_kinde_deletion` flag
- **`DeleteAccount` extended** — after existing cleanup, also enqueues `BookletUserDeletionArgs`

## Capabilities

### New Capabilities

- `booklet-deletion-sync`: Booklet HTTP client, the two new River jobs (`BookletUserDeletionArgs`, `DeleteAccountArgs`), and the new internal endpoint that receives deletion signals from Booklet

### Modified Capabilities

- `account-deletion`: `DeleteAccount` now also enqueues a `BookletUserDeletionArgs` job after committing the data wipe transaction

## Impact

- `backend/internal/usecase/account_usecase.go` — `DeleteAccount` enqueues the new Booklet job
- `backend/internal/usecase/job_args.go` — two new job arg types
- `backend/internal/booklet/booklet.go` — new package, Booklet HTTP client
- `backend/internal/handler/internal.go` — new `DELETE /internal/accounts/:id` handler (alongside existing internal routes from `internal-folder-api`)
- `backend/cmd/server/main.go` — wire up Booklet client and new handler
- `backend/internal/platform/config/config.go` — new Booklet config (base URL, internal secret)
- No frontend or extension changes
