## Why

Bookleaf has no account deletion mechanism. Regulations like GDPR require users be able to permanently delete their account and all associated personal data on request — this is a compliance gap. Rather than soft-delete/anonymise, deletion should be a genuine, irreversible removal ("scorched earth") of the user's data from the app's DB and R2 storage, and of their identity in Kinde. This change builds the backend mechanism; a frontend trigger can be wired up later.

## What Changes

- Add a protected `DELETE /me` endpoint that triggers full account deletion for the authenticated user.
- Add an account-deletion usecase that, in a single DB transaction:
  - Hard-deletes all of the user's images, folders, tags, and pending uploads
  - Enqueues existing `R2Delete` jobs for every affected image (and thumbnail) + pending upload object
  - Marks the user row as pending Kinde deletion (a tombstone state) rather than deleting it immediately
- Add a Kinde Management API client using M2M client-credentials auth, supporting: fetch/cache M2M access token, revoke a user's session(s) (`delete:user_session`), delete a user (`delete:users`)
- Add a River job/worker that, for tombstoned users, calls Kinde to revoke the session and delete the user, then hard-deletes the user row on success; retries on failure following existing job patterns
- Add a periodic reconciliation job that sweeps for tombstoned users whose Kinde-deletion job never completed and re-enqueues it

## Capabilities

### New Capabilities
- `account-deletion`: `DELETE /me` endpoint, the orchestration usecase that wipes a user's owned app data (images, folders, tags, pending uploads) and queues R2 cleanup, the Kinde Management API client, the async job that performs Kinde session/account deletion with retries, and the reconciliation sweep for stuck tombstones.

### Modified Capabilities
- `user-domain`: adds a `pending_kinde_deletion` tombstone field (+ migration) to the `User` model, marking accounts whose app data has been wiped but whose Kinde identity deletion is still pending.

## Impact

- **New external dependency**: Kinde Management API via M2M client-credentials — requires new config (M2M client ID/secret, management API base URL) and a new client package (e.g. `internal/platform/kinde`).
- New DB migration adding `pending_kinde_deletion` to `users`.
- New River job kind + periodic reconciliation job, following the `async-job-queue` capability's existing worker/periodic-job patterns.
- Reuses the existing `R2DeleteArgs`/`R2Delete` worker for object cleanup — no changes to that worker.
- Affects `internal/repository/{image,folder,tag,pending_upload,user}_repository.go` (new bulk-delete-by-user methods), `internal/usecase/user_usecase.go` (new account-deletion orchestration), `internal/handler/me.go` (new `DELETE` handler), `cmd/server/main.go` (wiring, config, job registration).
- Frontend trigger/confirmation UI is explicitly out of scope for this change.
