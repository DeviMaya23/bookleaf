## 1. Backend config

- [x] 1.1 Add a `MaintenanceConfig` struct (`Enabled bool`, `BypassToken string`) to `internal/platform/config/config.go` and a `Maintenance MaintenanceConfig` field on `Config`.
- [x] 1.2 Populate `Maintenance.Enabled` via `envWithDefault("MAINTENANCE_MODE", "false") == "true"` (matches the existing `OTEL_ENABLED` pattern) and `Maintenance.BypassToken` via `envWithDefault("MAINTENANCE_BYPASS_TOKEN", "")`.
- [x] 1.3 Add `config_test.go` cases: `MAINTENANCE_MODE` unset → `Enabled` false; `"true"` → true; invalid value (e.g. `"notabool"`) → false; `MAINTENANCE_BYPASS_TOKEN` unset → `BypassToken` `""`.

## 2. Backend maintenance middleware

- [x] 2.1 Create `internal/handler/middleware/maintenance.go` exporting a constructor (e.g. `NewMaintenanceMiddleware(cfg config.MaintenanceConfig) echo.MiddlewareFunc`) that:
  - Skips the gate if `X-Bookleaf-Bypass` header is non-empty and matches a non-empty `cfg.BypassToken`.
  - Otherwise, if `cfg.Enabled`, responds `503` with body `{"error":"maintenance"}` and header `X-Bookleaf-Maintenance: true`.
  - Otherwise calls `next(c)`.
- [x] 2.2 Register the middleware on the `protected` group in `cmd/server/main.go`, before `authMiddleware`.
- [x] 2.3 Table-driven unit tests in `maintenance_test.go`:
  - maintenance disabled → request passes through unchanged
  - maintenance enabled, no bypass header → `503` + `X-Bookleaf-Maintenance: true` + `{"error":"maintenance"}` body
  - maintenance enabled, no `Authorization` header → `503`, not `401`
  - maintenance enabled + matching bypass header → passes through to next middleware
  - maintenance enabled + non-matching bypass header → `503`
  - maintenance enabled + `MAINTENANCE_BYPASS_TOKEN` unset/empty + any bypass header → `503`

## 3. CORS update

- [x] 3.1 Add `X-Bookleaf-Bypass` to `AllowHeaders` in the `echomiddleware.CORSWithConfig` setup in `initEcho` (`cmd/server/main.go`).
- [x] 3.2 Extend the existing CORS test(s) to assert `X-Bookleaf-Bypass` is permitted alongside `Authorization`/`Content-Type`.

## 4. Frontend maintenance store

- [x] 4.1 Create `frontend/src/lib/maintenanceStore.ts`: a module-level boolean backed by `useSyncExternalStore`, exporting `useMaintenanceActive(): boolean` and `setMaintenanceActive(active: boolean): void`.
- [x] 4.2 Unit test: default state is `false`; calling `setMaintenanceActive(true)` updates what `useMaintenanceActive()` returns for subscribed components.

## 5. apiFetch changes

- [x] 5.1 In `frontend/src/lib/api.ts`, attach `X-Bookleaf-Bypass: <token>` when `localStorage.getItem('bookleaf-maintenance-bypass')` is non-empty.
- [x] 5.2 After the response resolves, call `setMaintenanceActive(true)` if the `X-Bookleaf-Maintenance` response header equals `"true"`, otherwise call `setMaintenanceActive(false)`.
- [x] 5.3 Extend `frontend/src/lib/api.test.ts`:
  - bypass header attached when localStorage value present; omitted when absent
  - maintenance store set to `true` when response includes `X-Bookleaf-Maintenance: true`
  - maintenance store set to `false` when response does not include the header

## 6. MaintenancePage + AppLayout gating

- [x] 6.1 Create `frontend/src/components/MaintenancePage.tsx` with placeholder copy: "Bookleaf is down for scheduled maintenance. We'll be back shortly — thanks for your patience."
- [x] 6.2 In `frontend/src/app-shell/AppLayout.tsx`, call `useMaintenanceActive()` and render `MaintenancePage` instead of the normal shell when `true`.
- [x] 6.3 Extend `frontend/src/app-shell/AppLayout.test.tsx`: renders `MaintenancePage` when maintenance state is active; renders the normal shell when inactive; switches back to the normal shell when the state transitions to `false`.

## 7. Ops documentation

- [x] 7.1 Add a short section to `README.md` documenting the maintenance toggle (`gcloud run services update bookleaf-backend --region=australia-southeast1 --update-env-vars=MAINTENANCE_MODE=true|false`) and the bypass header/token usage for QA.

## 8. Verification

- [x] 8.1 Run `golangci-lint run` in `backend/` and fix any issues.
- [x] 8.2 Run `npm run build` and `npm run lint` in `frontend/` and fix any issues.
