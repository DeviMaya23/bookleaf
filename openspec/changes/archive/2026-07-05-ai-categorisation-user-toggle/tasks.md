## 1. Repository

- [x] 1.1 Add `CountByUserAndMonth(ctx context.Context, userID string, year, month int) (int, error)` to the `categorisationLogRepository` interface in `categorisation_usecase.go`
- [x] 1.2 Implement `CountByUserAndMonth` in `categorisation_log_repository.go` — `SELECT COUNT(*)` scoped to `user_id` and the given UTC calendar month

## 2. Categorisation Usecase

- [x] 2.1 Add `const categorisationMonthlyLimit = 50` in `categorisation_usecase.go`
- [x] 2.2 Add `CountThisMonth(ctx context.Context, userID string) (int, error)` method to `CategorisationUsecase` — delegates to `CountByUserAndMonth` with `time.Now().UTC()` year and month
- [x] 2.3 Add the monthly limit check to `CategoriseImage`: call `CountThisMonth` before the existing log check; if count >= limit, return `nil` immediately without calling the agent or creating a log entry

## 3. User Usecase

- [x] 3.1 Add `AICategorisationEnabled *bool` to `UpdateUserPreferencesParams` in `user_usecase.go`
- [x] 3.2 Handle `AICategorisationEnabled` in `UpdatePreferences` — add to the `fields` map when non-nil

## 4. Me Handler

- [x] 4.1 Define `CategorisationCountUsecase` interface in `me.go` with `CountThisMonth(ctx context.Context, userID string) (int, error)`
- [x] 4.2 Add the interface as a field on `MeHandler`; update `NewMeHandler` signature to accept it
- [x] 4.3 Update `GetMe` to call `CountThisMonth` and include `ai_categorisation_count_this_month` in the JSON response
- [x] 4.4 Add `AICategorisationEnabled json.RawMessage` to `updateMeRequest` struct
- [x] 4.5 Parse and pass `AICategorisationEnabled` in `UpdateMe`; update the empty-body guard to cover all three fields
- [x] 4.6 Update `UpdateMe` response to include `ai_categorisation_count_this_month` (call `CountThisMonth` after the update)

## 5. Wiring

- [x] 5.1 Pass `categorisationUsecase` as the new `CategorisationCountUsecase` argument to `NewMeHandler` in `main.go`
- [x] 5.2 Update `bruno/update-me.bru` to include `"ai_categorisation_enabled": true` in the example body

## 6. Backend Tests

- [x] 6.1 `CategoriseImage`: scenario where monthly count >= 50 — assert agent is not called and `nil` is returned
- [x] 6.2 `CategoriseImage`: scenario where count query errors — assert error is returned and agent is not called
- [x] 6.3 `UpdatePreferences`: scenario where `AICategorisationEnabled` is set — assert the field is persisted
- [x] 6.4 Handler `GET /me`: assert `ai_categorisation_count_this_month` is present in the response body
- [x] 6.5 Handler `PATCH /me`: scenario with `{ "ai_categorisation_enabled": true }` — assert `200 OK` and field is updated
- [x] 6.6 Handler `PATCH /me`: scenario with `{ "ai_categorisation_enabled": "yes" }` — assert `400 Bad Request`

## 7. Frontend Types

- [x] 7.1 Add `ai_categorisation_count_this_month: number` to the `Me` interface in `me.ts`
- [x] 7.2 Add `ai_categorisation_enabled?: boolean` to `UpdateMeParams` in `me.ts`

## 8. AdvancedSection

- [x] 8.1 Add a second mutation in `AdvancedSection.tsx` for `ai_categorisation_enabled`
- [x] 8.2 Render the AI auto-categorisation toggle row below the AI folder suggestions row, with `Switch` bound to `ai_categorisation_enabled` and a `"X / 50 this month"` counter; mute the counter text when count is 0
- [x] 8.3 Unit test: toggle reflects `ai_categorisation_enabled` from the `me` cache
- [x] 8.4 Unit test: toggling the switch fires `PATCH /me` with `{ "ai_categorisation_enabled": <new value> }`

## 9. ProfileMenu Badge

- [x] 9.1 In `ProfileMenu.tsx`, derive `showBadge` from the `me` query cache: `ai_categorisation_enabled && ai_categorisation_count_this_month >= 50`
- [x] 9.2 Add a dismissal state using a `localStorage` key `categorisation_limit_dismissed_<YYYY-MM>` (UTC); on mount, read the key and suppress the badge if present
- [x] 9.3 On `DropdownMenuTrigger` click, write the localStorage key for the current UTC month to dismiss the badge
- [x] 9.4 Render a small red dot badge overlaid on the `Avatar` when `showBadge` is true and not dismissed
- [x] 9.5 Unit test: badge renders when limit is hit and key is not in localStorage
- [x] 9.6 Unit test: badge does not render when count < 50 or `ai_categorisation_enabled` is false
- [x] 9.7 Unit test: clicking the trigger sets the localStorage key and hides the badge

## 10. Build & Lint

- [x] 10.1 Run `golangci-lint run ./...` from `backend/` and fix any issues
- [x] 10.2 Run `npm run build` and `npm run lint` from `frontend/` and fix any issues
