## Context

Currently `users.id` stores the Kinde-issued subject (e.g. `kp_abc123`) as a `TEXT` primary key. This string propagates as a foreign key into `folders`, `images`, `tags`, `pending_uploads`, and `ai_categorisation_logs`, and is extracted verbatim from the JWT `sub` claim and placed in the Echo request context. The app's identity is fully entangled with the IdP's.

The change introduces an app-generated UUID as the stable internal ID, with the Kinde subject demoted to `idp_subject` — a lookup bridge and cross-system communication key.

## Goals / Non-Goals

**Goals:**
- App owns `users.id` (UUID); Kinde subject is stored as `idp_subject` for IdP operations only
- Auth context carries both the internal UUID and the IDP subject, giving handlers the right key for each operation
- Internal API endpoints continue to accept Kinde subject from external callers (booklet), resolving to UUID internally
- Kinde and Booklet API calls use `idp_subject`, not the internal UUID
- Data migration via pure SQL in a single transaction within a maintenance window

**Non-Goals:**
- Changing what external callers (booklet) send — they continue using Kinde subject
- Migrating existing R2 objects (stored paths remain valid as opaque strings)
- Frontend or extension changes beyond dropping the unused `id` field from `GET /me`

## Decisions

### Decision: `users.id` column type becomes UUID

`users.id` changes from `TEXT` to `UUID` at the DB level. The `domain.User` struct changes `ID` from `string` to `uuid.UUID` (consistent with all other entity types in the codebase). GORM handles the mapping natively. All usecases that currently accept `userID string` will accept `userID uuid.UUID` where the value originates from the auth context or a resolved lookup.

**Alternative considered**: Keep `id` as TEXT, store a UUID string. Rejected — the column type should match the value's actual type, and `uuid.UUID` is already the convention everywhere else.

### Decision: Auth context stores both internal UUID and IDP subject

After `GetOrProvision` returns the provisioned user, the middleware stores two values:
- `AuthenticatedUserIDContextKey` → `user.ID.String()` (the internal UUID, used by most handlers)
- `AuthenticatedIDPSubjectContextKey` → `claims.Subject` (the Kinde subject, used by deletion handlers)

A typed helper `AuthenticatedUserUUIDFromContext` parses the UUID string into `uuid.UUID` for handlers that need it. The original `AuthenticatedUserIDFromContext` (string) is retained for compatibility.

**Why store IDP subject separately**: `MarkForDeletion` operates in IDP subject space (see decision below). `DeleteMe` needs the Kinde subject to call it, but the auth context UUID is the value that represents the user everywhere else. Storing both avoids an extra DB round-trip to re-resolve UUID → IDP subject inside the usecase.

### Decision: `GetOrCreate` looks up by `idp_subject`, not `id`

`UserRepository.GetOrCreate(ctx, idpSubject string)` is refactored to:
1. `SELECT * FROM users WHERE idp_subject = $1 LIMIT 1`
2. If not found: `INSERT INTO users (id, idp_subject, ...) VALUES (gen_random_uuid(), $1, ...) ON CONFLICT (idp_subject) DO NOTHING RETURNING *`

This keeps the interface signature identical (kinde_id in → `*domain.User` out), so `GetOrProvision` in the usecase and the middleware are untouched.

### Decision: `AccountWipeArgs` carries `IDPSubject`, not internal UUID

`AccountWipeArgs{IDPSubject string}` replaces the old `{UserID string}`. `WipeAccount(ctx, idpSubject string)` resolves to the user row via `GetByIDPSubject` at the start of execution:

- **User found**: full wipe (DB data → R2 jobs → Kinde API via `idpSubject` → Booklet deletion job)
- **User not found** (unprovisioned): Kinde-only cleanup using `idpSubject` directly

This collapses the two current code paths (provisioned and unprovisioned wipe) into one job type. The job always carries a Kinde subject, which is the right key for Kinde API calls regardless of whether a DB row exists.

**Alternative considered**: Keep `UserID uuid.UUID` on the args and add a separate `KindeOnlyWipeArgs` job for the unprovisioned case. Rejected — adds a second job kind for a rare path; the resolution-by-idp_subject approach handles it cleanly.

### Decision: `BookletUserDeletionArgs` carries `IDPSubject`

Booklet identifies users by Kinde subject. The enqueued job must carry the Kinde subject, not the internal UUID, so the `bookletClient.DeleteUser` call can supply the correct value in the `DELETE /internal/users/:id` path.

### Decision: `MarkForDeletion` operates on IDP subject; `ListPublicFolders` resolves via `InternalUserResolver`

`MarkForDeletion(ctx, idpSubject string)` accepts the Kinde subject and calls `GetByIDPSubject` internally. This means both callers — `DeleteMe` (passes `AuthenticatedIDPSubjectContextKey` from context) and `DeleteAccount` (passes the URL param directly) — hand off the Kinde subject without any pre-resolution. The usecase handles the provisioned and unprovisioned paths in one place.

This is the alternative that was originally rejected in the proposal, but it was chosen during implementation because:
- `MarkForDeletion` needs the IDP subject to enqueue `AccountWipeArgs{IDPSubject}` regardless, so passing UUID would require an immediate `GetByID` → `user.IDPSubject` lookup inside the usecase anyway.
- The caller (`DeleteMe`) only has the IDP subject cheaply available via context; resolving UUID → IDP subject inside the usecase would add a round-trip.

`InternalHandler.ListPublicFolders` is the exception: it resolves via a dedicated `InternalUserResolver` interface (`GetByIDPSubject` → `user.ID`) before calling `GetPublicFoldersByUser(uuid.UUID)`, because the share usecase operates in internal UUID space. Unknown Kinde subjects return an empty folder list (200).

**Boundary in practice**: most usecases operate on `uuid.UUID`. `MarkForDeletion` and `WipeAccount` are the exceptions — they stay in IDP subject space because Kinde API calls are their primary concern.

### Decision: Pure SQL migration, no app-side data manipulation

The migration runs in a single PostgreSQL transaction:
1. Add `idp_subject TEXT` to `users`, backfill from `id`, add `UNIQUE NOT NULL` constraint
2. Add `new_id UUID DEFAULT gen_random_uuid()` to `users`
3. Add `new_user_id UUID` staging columns to all FK tables
4. Backfill FK tables: `UPDATE fk_table SET new_user_id = u.new_id FROM users u WHERE fk_table.user_id = u.id`
5. Drop FK constraints, drop old `user_id` columns, rename `new_user_id` → `user_id`, re-add FK constraints
6. On `users`: drop old `id` PK and column, rename `new_id` → `id`, add UUID primary key constraint
7. Update `ai_categorisation_logs.user_id` (no FK, same staging column pattern)

`gen_random_uuid()` is available natively in PostgreSQL 13+. No app code executes during migration.

## Risks / Trade-offs

**[Risk] In-flight River jobs carry old Kinde subject as `UserID`** → Drain the River queue before running the migration. The maintenance window should not close until the queue is confirmed empty. Any `AccountWipeArgs` jobs queued before migration with the old format will fail after — they carry a Kinde subject but `WipeAccount` after the change expects to look up by `idp_subject`, which will work if the job is re-enqueued in the new format. Manual reconciliation may be needed for any stuck jobs at cutover.

**[Risk] `users.id` type change breaks GORM model scans across the codebase** → All repositories that scan into `domain.User` will pick up the UUID automatically via GORM. Any raw SQL that casts `user_id` as `TEXT` (rather than `UUID`) will need updating. Audit raw SQL queries in repositories before implementation.

**[Risk] `ai_categorisation_logs` old rows have Kinde subjects; new rows have UUIDs** → This table has no FK constraint. After migration, old rows have the Kinde subject as `user_id`, new rows have the UUID. If any query filters by `user_id` on this table (e.g. count this month), it will use the UUID and miss old rows. Current code in `categorisation_usecase.go` scopes by `user_id` — those queries will only cover post-migration records for existing users. Acceptable for a log/count table; flagged for awareness.

## Migration Plan

1. **Maintenance window open** — set maintenance mode, block all traffic
2. **Drain River queue** — confirm no pending jobs remain
3. **Run migration** — single SQL transaction as described in Decisions
4. **Deploy new application binary** — code that uses `idp_subject`, UUID-typed `id`, updated job args
5. **Smoke test** — login, provision user, verify UUID in context, verify existing data accessible
6. **Maintenance window close**

**Rollback**: If migration fails mid-transaction, PostgreSQL rolls back automatically — no partial state. If the new binary is bad post-migration, the migration itself cannot be cleanly rolled back (the column types have changed). A pre-migration DB snapshot is the rollback path for data.

## Open Questions

~~Are there any raw SQL queries in repositories that cast `user_id` explicitly to `TEXT`?~~ Audited — none found.

~~Should `GET /me` drop the `id` field?~~ Confirmed safe. `Me.id` is declared in `frontend/src/features/auth/lib/me.ts` but never read by any component or test. The extension does not call `/me`. Drop the field from both the backend response and the FE type.
