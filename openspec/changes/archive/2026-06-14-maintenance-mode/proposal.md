## Why

During deployments that include breaking changes (e.g. DB migrations), the authenticated app (`/app`) can be left in an inconsistent state and surface confusing raw errors to users. A maintenance mode toggle lets the team proactively show a clear "down for maintenance" message during these windows, flip it on/off via a single CLI command (no new infrastructure, no console access), and still let devs/QA verify the deploy via a bypass before re-opening to everyone.

## What Changes

- New backend middleware (registered after CORS, before auth) gates all routes except `/health` behind a `MAINTENANCE_MODE` env var. When active, it responds `503` with a JSON body and an `X-Bookleaf-Maintenance: true` response header.
- New bypass: the middleware checks an `X-Bookleaf-Bypass` request header against a `MAINTENANCE_BYPASS_TOKEN` env var (set once at deploy). A match skips the maintenance gate entirely, regardless of `MAINTENANCE_MODE`.
- **Modified (be-cors)**: `X-Bookleaf-Bypass` added to the CORS `Access-Control-Allow-Headers` list alongside `Authorization` and `Content-Type`.
- Frontend `apiFetch` (`frontend/src/lib/api.ts`):
  - Attaches `X-Bookleaf-Bypass: <token>` from `localStorage` if a token has been set (manual devtools step for devs/QA).
  - Detects `X-Bookleaf-Maintenance: true` on responses and updates shared state.
- A wrapper around the `/app` routes renders a new `MaintenancePage` (copy TBD — placeholder text for now) instead of the normal app shell while maintenance state is active.
- Public/static pages (`/`, `/about`, `/privacy`, `/ai-notes`) are unaffected — they don't call the API and are explicitly out of scope.
- Toggle is operational, not new infra: `gcloud run services update bookleaf-backend --region=australia-southeast1 --update-env-vars=MAINTENANCE_MODE=true|false`, documented as convenience scripts.

## Capabilities

### New Capabilities
- `maintenance-mode`: Backend middleware that gates requests behind a `MAINTENANCE_MODE` env var, including the `X-Bookleaf-Bypass` / `MAINTENANCE_BYPASS_TOKEN` bypass and the `/health` exclusion.
- `fe-maintenance-mode`: Frontend detection of the maintenance signal via `apiFetch`, the bypass header attached from `localStorage`, and the `/app`-route `MaintenancePage` gating.

### Modified Capabilities
- `be-cors`: `Access-Control-Allow-Headers` SHALL additionally permit `X-Bookleaf-Bypass`.

## Impact

- **Backend**: new middleware in `internal/handler/middleware/`, registered in `cmd/server/main.go`'s `initEcho`/route setup; new `MAINTENANCE_MODE` and `MAINTENANCE_BYPASS_TOKEN` config fields (`internal/platform/config`); CORS `AllowHeaders` updated.
- **Frontend**: `frontend/src/lib/api.ts` (`apiFetch`) — shared by `features/auth/lib/me.ts`, `lib/folders.ts`, `lib/tags.ts`, `lib/images.ts`; new shared maintenance-state module; new `MaintenancePage` component; route wrapper around `/app` in `App.tsx`/`AppLayout`.
- **Deploy/runbook**: new toggle commands (env var update via `gcloud run services update`), no Cloud Run config/infra changes beyond the two new env vars.
- **No new HTTP endpoints** — no Bruno collection changes needed. No database changes.
