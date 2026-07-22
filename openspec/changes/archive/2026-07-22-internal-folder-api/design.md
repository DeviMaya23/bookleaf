## Context

Bookleaf exposes all user-facing routes behind Kinde JWT auth. A second backend application needs to query folder data (public folder lists, contents with presigned URLs, and public status) on behalf of its own users — without a Kinde token, since the caller is a trusted service, not an end user.

The existing public routes (`GET /share/:token`, `GET /share/:token/export`) already demonstrate the pattern of bypassing auth for specific endpoints. The `MaintenanceMiddleware` demonstrates the shared-secret-header pattern already in use in this codebase.

The `FolderShareRepository` already has `GetByFolderID` and `GetByToken`. The `shareUsecase` already has `GetSharedFolder` (token-based) and the full presigned URL assembly pipeline. The new internal endpoints extend these with folder-ID-based and user-ID-based entry points.

## Goals / Non-Goals

**Goals:**
- Protect three new `/internal/*` endpoints with a shared secret header
- Expose folder public status, folder contents (with presigned URLs), and user's public folder list to a trusted backend caller
- Reuse the existing share usecase and repository infrastructure where possible

**Non-Goals:**
- Separate port or process isolation for internal routes (infra-level routing assumed)
- Rate limiting or per-caller identity on the internal routes
- Any frontend or extension changes

## Decisions

### D1: Shared secret header, not JWT

The caller is a backend service with no Kinde identity. Options:
- **Shared secret header** (chosen): simple, already precedented by `X-Bookleaf-Bypass` / `MaintenanceMiddleware`. One env var, one middleware.
- **mTLS**: strong, but requires cert infrastructure not present in this stack.
- **Kinde M2M token**: Bookleaf already has M2M credentials but they're used for outbound Kinde management calls, not inbound auth. Repurposing them adds coupling.

Header name: `X-Bookleaf-Internal-Secret`. Config key: `INTERNAL_API_SECRET`.

The secret must be non-empty — if the env var is absent, startup fails rather than silently accepting all requests.

### D2: `/internal` route group on the same port

Options:
- **Same port, route-group middleware** (chosen): zero infra changes, consistent with how maintenance bypass works.
- **Separate port**: stronger isolation but requires provisioning a second listener and network rules.

Same-port is acceptable because: (a) infra-level routing is assumed to restrict access, and (b) the shared secret remains the auth mechanism regardless of port.

### D3: `GetByFolderIDWithFolder` as a new repo method (Option A)

Endpoint 2 (folder contents by folder ID) needs the folder's `UserID` to call `imageRepo.ListByFolder`. Options:
- **New `GetByFolderIDWithFolder` repo method** (chosen): single DB query with Preload, matches the existing `GetByToken` pattern exactly.
- **Two calls** (`GetByFolderID` + `folderRepo.GetByID`): two queries, requires a userID argument we don't have at the internal layer.

The new method follows the same shape as `GetByToken` but queries by `folder_id` instead of `token`.

### D4: No new domain type for endpoint 3 response

`GET /internal/folders/:folder_id/status` returns `{"token": "..."}` on 200. This reuses the existing `shareTokenResponse` struct already defined in `handler/share.go`. No new type needed.

The handler reads the token from `FolderShareRepository.GetByFolderID` directly via a thin usecase method (`CheckFolderPublicStatus`), translating `gorm.ErrRecordNotFound` → 404.

### D5: `InternalHandler` is separate from `ShareHandler`

The three internal endpoints share usecase dependencies with `ShareHandler` but serve a different caller context (no user auth, different ownership semantics). Keeping them in `handler/internal.go` behind a distinct `InternalShareUsecase` interface avoids polluting `ShareHandler` with methods that have no auth context.

### D6: `INTERNAL_API_SECRET` is required at startup

Unlike `MAINTENANCE_BYPASS_TOKEN` (optional — maintenance mode can be enabled without a bypass), the internal secret has no meaningful fallback. An empty secret would accept all requests, which is a security hole. The config loader returns an error if the value is empty.

## Risks / Trade-offs

- **Secret rotation requires restart**: Changing `INTERNAL_API_SECRET` requires redeploying Bookleaf. No hot-reload. Acceptable given deployment cadence.
- **No per-request caller identity**: All requests with the correct secret are indistinguishable. If the calling app is ever compromised, the only remediation is rotating the secret and redeploying. Acceptable for two-service setup.
- **Presigned URL TTL**: Endpoint 2 returns presigned URLs with the existing `presignedGetTTL`. If the calling app's FE caches these links longer than the TTL, they'll expire. The caller must handle this.
- **`folder_id` without `user_id` on endpoint 3**: `CheckFolderPublicStatus` looks up only in `folder_shares`, not `folders`. A folder_id that exists but is private returns 404; a folder_id that doesn't exist at all also returns 404. The caller cannot distinguish these two cases — intentional, as the distinction doesn't matter for the use case.

## Migration Plan

1. Add `INTERNAL_API_SECRET` to all environment configs before deploying (startup will fail without it).
2. Deploy backend — new routes are immediately live behind the secret.
3. Configure the calling backend app with the secret value.
4. No rollback complexity: removing the route group reverts the change cleanly.

## Open Questions

None — all decisions resolved in explore session.
