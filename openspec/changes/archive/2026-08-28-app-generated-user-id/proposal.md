## Why

The app currently uses the Kinde-issued subject as `users.id`, coupling our primary key space to a third-party IdP. This makes future IdP migrations expensive and conflates two distinct identities: who the user is inside our system vs. how they authenticated externally.

## What Changes

- **BREAKING** `users.id` changes from `TEXT` (Kinde subject) to `UUID` (app-generated). All FK columns in `folders`, `images`, `tags`, `pending_uploads`, and `ai_categorisation_logs` change accordingly.
- New `users.idp_subject` column stores the Kinde subject (previously `id`). Unique, not null.
- `UserRepository` gains `GetByIDPSubject(ctx, idpSubject string)` for lookups by Kinde subject.
- Auth middleware stores the internal UUID in context (resolved via `GetOrProvision`) instead of `claims.Subject`.
- Internal API endpoints (`DELETE /internal/users/:id`, `GET /internal/users/:user_id/folders`) continue to accept the Kinde subject from callers (booklet); they resolve to internal UUID before hitting repositories.
- `WipeAccount` fetches `user.IDPSubject` from the user row to call Kinde's delete/session APIs.
- `BookletUserDeletionArgs.UserID` carries `idp_subject` (Kinde subject) since booklet identifies users by Kinde ID.
- `GET /me` response drops the `id` field (unused in FE).
- Data migration: backfill `idp_subject = id`, generate new UUIDs for all `users.id`, cascade-update FK columns. Requires maintenance window.

## Capabilities

### New Capabilities

- `user-id-ownership`: App-owned user UUID, `idp_subject` as the bridge to the external IdP, and the repository/middleware conventions for resolving between the two.

### Modified Capabilities

- `user-domain`: User entity gains `IDPSubject` field; `ID` type changes to UUID.
- `kinde-auth`: Auth middleware stores internal UUID (not Kinde subject) in request context after provisioning.
- `account-deletion`: Deletion flow uses `idp_subject` for Kinde API calls and booklet deletion job.
- `booklet-deletion-sync`: Booklet deletion job carries `idp_subject`; internal endpoints accept Kinde subject and resolve to UUID.
- `internal-folder-api`: Internal folder listing endpoint resolves incoming Kinde subject to internal UUID before querying.

## Impact

- **Backend**: `domain.User`, all repositories, auth middleware, account usecase, internal handler, booklet worker args.
- **Database**: Migration touching `users`, `folders`, `images`, `tags`, `pending_uploads`, `ai_categorisation_logs`. Requires maintenance window.
- **R2**: No file migration needed — existing `r2_path` values are stored as opaque strings and still resolve correctly. New uploads use paths under the new UUID.
- **Booklet**: No contract change — internal endpoints continue to accept Kinde subject. Booklet-side is unaffected.
- **Frontend / Extension**: `id` field removed from `GET /me` response. No other API changes.
