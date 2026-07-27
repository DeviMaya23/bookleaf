## Context

Bookleaf and Booklet share a Kinde identity ecosystem. Bookleaf already orchestrates its own account deletion via `DeleteAccount` (DB wipe transaction → async R2 and Kinde jobs). This change extends that orchestration to cover Booklet's data as well, and introduces an internal endpoint so Booklet can signal Bookleaf to run the same flow when deletion originates from Booklet's side.

The existing internal API group (`/internal`, gated by `X-Bookleaf-Internal-Secret` shared secret middleware) provides the right foundation for the new endpoint. The `internal/vision` package establishes the pattern for external HTTP client adapters.

## Goals / Non-Goals

**Goals:**
- Both apps purge their own data when deletion is initiated from either side
- Bookleaf remains the sole orchestrator — it calls Booklet as part of its own deletion flow
- All cross-app calls are durable (River jobs with retries)
- Circular calls terminate cleanly without a new DB column

**Non-Goals:**
- Booklet-side implementation (out of scope)
- Strong consistency — eventual cleanup of Booklet records is acceptable
- Handling the case where a user exists in Booklet but not in Bookleaf

## Decisions

### Booklet HTTP client lives in `internal/booklet/`

Follows the `internal/vision/` pattern exactly. The concrete client (`booklet.Client`) lives in `internal/booklet/booklet.go`. The interface (`BookletClient`) is defined in the `usecase` package so the dependency flows inward. Wired in `cmd/server/main.go`.

The client sends `X-Booklet-Internal-Secret` as a header (new config value `BookletInternalSecret`), alongside `BookletBaseURL`. Calls `DELETE /internal/users/:id` on Booklet.

### Two new jobs at different levels of abstraction

**`BookletUserDeletionArgs`** (kind: `booklet_user_deletion`, max 10 attempts) — the outbound notification. Enqueued by `DeleteAccount` after the DB wipe transaction. Calls `bookletClient.DeleteUser`. This job is a side-effect produced by the deletion pipeline.

**`DeleteAccountArgs`** (kind: `delete_account`, max attempts: 5) — the async entry point. Enqueued by the new internal endpoint. Calls `accountUsecase.DeleteAccount` directly. Retries are necessary: if `DeleteAccount` fails before the DB transaction commits (e.g. transient DB error), no downstream jobs are enqueued yet — River must retry the whole job or the deletion is lost. Once the transaction commits, downstream jobs (`R2DeleteArgs`, `AccountKindeDeletionArgs`, `BookletUserDeletionArgs`) carry their own retry budgets and a second `DeleteAccountArgs` attempt is a safe no-op.

```
Booklet-triggered path:
  DELETE /internal/accounts/:id
    → enqueue DeleteAccountArgs
      → DeleteAccount (DB wipe, R2 jobs, Kinde job)
        → enqueue BookletUserDeletionArgs
          → DELETE /internal/users/:id on Booklet  [idempotent 2xx]
            → Booklet calls back DELETE /internal/accounts/:id
              → pending_kinde_deletion = true → 202, no re-enqueue ✓

Bookleaf-triggered path:
  DELETE /me
    → DeleteAccount (DB wipe, R2 jobs, Kinde job)
      → enqueue BookletUserDeletionArgs
        → DELETE /internal/users/:id on Booklet
          → Booklet calls back DELETE /internal/accounts/:id
            → pending_kinde_deletion = true → 202, no re-enqueue ✓
```

### `accountUsecase` gains a `BookletClient` dependency

`accountUsecase` receives a `bookletClient BookletClient` field. The `BookletClient` interface exposes `DeleteUser(ctx context.Context, userID string) error`. If `bookletClient` is nil (e.g. in local dev without Booklet configured), the enqueue step is skipped — same pattern as `visionService` in `imageUploadUsecase`.

### Idempotency via existing `pending_kinde_deletion` — no new column

The `DELETE /internal/accounts/:id` handler checks the user row:

- User not found → enqueue `AccountKindeDeletionArgs` (Kinde-only, no DB wipe needed), return `202 Accepted`
- `pending_kinde_deletion = true` → `202 Accepted`, no enqueue (already in pipeline)
- User found, not pending → enqueue `DeleteAccountArgs`, return `202 Accepted`

`DeleteAccount` sets `pending_kinde_deletion` synchronously inside the DB wipe transaction, before any jobs are enqueued. By the time Booklet calls back, the flag is already set — the back-call hits the second case and terminates the cycle.

### The new endpoint joins the existing `/internal` group

`DELETE /internal/accounts/:id` is registered on the same Echo group as the existing internal folder/share routes, protected by the same `X-Bookleaf-Internal-Secret` middleware. No new auth mechanism or route group needed.

## Risks / Trade-offs

**Circular call race** → `pending_kinde_deletion` is set in the DB transaction that runs synchronously in `DeleteAccount`, before `BookletUserDeletionArgs` is enqueued. There is no window where the flag is unset when Booklet calls back. Low risk.

**Booklet persistently unavailable (> 10 attempts)** → orphaned records remain in Booklet's DB. The Kinde identity is still deleted, so the user has no access. Orphaned records are a data hygiene problem, not a security problem. Acceptable trade-off given the "best effort" framing.

**`DeleteAccountArgs` called for already-hard-deleted user** → user row is gone after the Kinde job completes. The internal endpoint falls into the "user not found" branch and enqueues `AccountKindeDeletionArgs`, which calls `kinde.DeleteUser` (already deleted — treated as success) then `userRepo.HardDelete` (no-op). Harmless.

**`DeleteAccount` not fully idempotent** → if called twice concurrently (race between user-triggered and Booklet-triggered paths), the DB operations are safe no-ops (DELETE WHERE on missing rows), but duplicate R2/Kinde/Booklet jobs would be enqueued. Downstream workers are idempotent so this is harmless, just wasteful. The `pending_kinde_deletion` gate on the internal endpoint makes this race unlikely in practice.
