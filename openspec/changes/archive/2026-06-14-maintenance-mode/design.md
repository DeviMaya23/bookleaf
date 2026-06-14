## Context

The backend (`backend/cmd/server/main.go`) registers `Recover` and `CORS` globally via `e.Use(...)` in `initEcho`, then in `initApp` builds a `protected := e.Group("")` group carrying the Kinde auth middleware and `LoggingMiddleware`. Only `/health` (and `/metrics` when OTEL is enabled) sit outside `protected` — every other route the frontend calls (`/me`, `/folders`, `/tags`, `/images*`, etc.) is in that group.

On the frontend, every API call funnels through `apiFetch` (`frontend/src/lib/api.ts`), which is called from plain `lib/*.ts` modules (`folders.ts`, `tags.ts`, `images.ts`, `features/auth/lib/me.ts`) — not from React components directly. The existing global-state precedent (`useTheme.tsx`) uses React Context + Provider, but Context can't be read or written from outside the component tree, so it doesn't fit a flag that needs to be set from `apiFetch`.

## Goals / Non-Goals

**Goals:**
- Gate every `/app`-relevant API response behind a `MAINTENANCE_MODE` env var, without affecting `/health`/`/metrics`.
- Allow a designated bypass (`X-Bookleaf-Bypass` header + `MAINTENANCE_BYPASS_TOKEN`) to skip the gate entirely.
- Surface the gated state to the FE via a response header, and have the `/app` shell swap to a `MaintenancePage` automatically — including automatically recovering once maintenance is turned off.
- Keep the toggle a single `gcloud run services update --update-env-vars` command — no new infrastructure.

**Non-Goals:**
- Gating public/static pages (`/`, `/about`, `/privacy`, `/ai-notes`) — they don't call the API.
- Final maintenance message copy (placeholder text only).
- Sub-second toggle latency (Cloud Run revision rollout is "tens of seconds," accepted as part of the existing deploy window).
- Hardening the bypass token against timing attacks — it's an operational convenience secret, not a session credential.

## Decisions

### 1. Middleware scope: the `protected` group, not a global `e.Use`
Register the maintenance middleware as `protected.Use(maintenanceMiddleware)`, placed **before** `authMiddleware` in the chain.

- Because `/health` and `/metrics` are registered directly on `e` and never join `protected`, they're excluded automatically — no path-based exclusion logic needed.
- Running it before `authMiddleware` means a request during maintenance gets a `503` without paying for JWT validation / `GetOrProvision` DB round-trips — except bypassed requests, which fall through to auth as normal.
- *Alternative considered*: global `e.Use(...)` with an explicit `if c.Path() == "/health" { return next(c) }` check. Rejected — adds a path-string check that has to be kept in sync if more public routes are added later, where group placement handles it structurally.

### 2. Response shape
On gate: `503` status, JSON body `{"error":"maintenance"}`, header `X-Bookleaf-Maintenance: true`. The header is the FE's detection signal (not the body), so it works uniformly even for non-JSON-consuming calls (e.g. image downloads).

### 3. Bypass check
Compare `X-Bookleaf-Bypass` request header to `cfg.Maintenance.BypassToken` with plain string equality, and only treat it as a bypass if `BypassToken` is non-empty (so an unset token never matches an empty header). Simple equality is acceptable per the Non-Goals — this token gates access to a maintenance *message*, not data.

### 4. Config additions
New `MaintenanceConfig` group in `internal/platform/config`:
- `MAINTENANCE_MODE` → `Enabled bool`, parsed via `strconv.ParseBool`, defaults to `false` if unset/invalid (so existing deploys are unaffected without env changes).
- `MAINTENANCE_BYPASS_TOKEN` → `BypassToken string`, optional, defaults to `""`.

### 5. CORS: add `X-Bookleaf-Bypass` to `AllowHeaders`
Modifies `be-cors`'s `Access-Control-Allow-Headers` list (currently `Authorization`, `Content-Type`) to include `X-Bookleaf-Bypass`, so the browser permits the FE to send it on every request, maintenance or not.

### 6. FE state: a `useSyncExternalStore`-backed module store, not Context
New `frontend/src/lib/maintenanceStore.ts` exposing `useMaintenanceActive(): boolean` and `setMaintenanceActive(active: boolean): void`. `apiFetch` calls the setter directly (it's a plain function, not a hook); `AppLayout` calls the hook to read it.

- *Alternative considered*: Context + Provider (the `useTheme` pattern). Rejected — `apiFetch` runs outside the component tree and can't access Context.
- *Alternative considered*: React Query global `onError`/cache entry. Rejected — not all `apiFetch` callers necessarily go through React Query, and this couples a cross-cutting concern to a specific data-fetching library's internals.
- `useSyncExternalStore` is a standard React hook (no new dependency), used here for exactly its intended purpose: subscribing a component to state that's mutated from outside React.

### 7. `apiFetch` changes
- Before the request: if `localStorage.getItem('bookleaf-maintenance-bypass')` is set, attach it as `X-Bookleaf-Bypass`.
- After the response: if `X-Bookleaf-Maintenance` header equals `"true"`, call `setMaintenanceActive(true)`; otherwise call `setMaintenanceActive(false)`. This makes recovery automatic — the next successful call after maintenance is turned off flips the flag back.

### 8. Gating point: top of `AppLayout`
All four `/app/*` routes (`index`, `unsorted`, `trash`, `folders/:folderId`) render `<AppLayout />`. Adding the `useMaintenanceActive()` check at the top of `AppLayout` (returning `<MaintenancePage />` instead of the normal layout when active) covers all of them with a single change, no new route-level wrapper component.

- *Alternative considered*: a dedicated `MaintenanceGate` wrapper route around the `/app` subtree. Rejected as unnecessary indirection — `AppLayout` is already the single shared element for that subtree.

## Risks / Trade-offs

- **[Cold-load flash]** On a fresh page load during maintenance, `AppLayout` renders normally until the first `apiFetch` call resolves with the header, causing a brief flash of the normal shell/loading state before `MaintenancePage` appears. → Accepted as a minor, one-time visual glitch; not worth an eager pre-flight call for this change.
- **[Bypass token persistence]** The bypass token sits in `localStorage` indefinitely once set. → Acceptable since it's a low-stakes operational secret (gates a message, not data); rotated by changing `MAINTENANCE_BYPASS_TOKEN` and redeploying if needed.
- **[Toggle latency]** Flipping `MAINTENANCE_MODE` still takes a Cloud Run revision rollout (tens of seconds), not instant. → Accepted; bundled into the existing deploy window for the breaking change itself.
- **[`X-Bookleaf-Maintenance` only set on gated responses]** A backend that's down for reasons *other* than the maintenance flag (e.g. crashed, unreachable) won't set this header, so `MaintenancePage` won't show — the user sees a network-error state instead. → Acceptable; this feature specifically targets the *planned* maintenance window, not arbitrary outages.

## Migration Plan

1. Deploy backend first: new middleware + config are inert by default (`MAINTENANCE_MODE` unset → `false`; `MAINTENANCE_BYPASS_TOKEN` unset → empty, never matches). No behavior change for existing traffic.
2. Deploy frontend: `apiFetch`/store/`AppLayout` changes are inert until the backend ever sends `X-Bookleaf-Maintenance: true`.
3. One-time ops setup: generate a `MAINTENANCE_BYPASS_TOKEN` value and set it as a Cloud Run env var (e.g. via the same `gcloud run services update --update-env-vars` toggle command), share with devs/QA via an existing secret-sharing channel.
4. To use during a breaking deploy: `gcloud run services update bookleaf-backend --region=australia-southeast1 --update-env-vars=MAINTENANCE_MODE=true`, perform the breaking deploy/migration, verify via the bypass token, then flip `MAINTENANCE_MODE=false`.
5. Rollback: setting `MAINTENANCE_MODE=false` (or omitting it) fully reverts behavior; no data migrations involved.

## Resolved

- **`MaintenancePage` copy** — placeholder for now (user will adjust later):
  > "Bookleaf is down for scheduled maintenance. We'll be back shortly — thanks for your patience."
- **`MAINTENANCE_BYPASS_TOKEN` value** — generated and added to the (gitignored) root `.env` for local dev, with a placeholder documented in `.env.example`. Set the same value as a Cloud Run env var manually (`gcloud run services update --update-env-vars=MAINTENANCE_BYPASS_TOKEN=...`) as part of the one-time ops setup.
