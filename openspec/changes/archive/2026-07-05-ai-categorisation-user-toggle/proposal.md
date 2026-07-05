## Why

`ai_categorisation_enabled` exists in the DB and the feature works end-to-end, but users have no way to opt in — it can only be flipped via direct database access. Surfacing it in settings lets users enable smart auto-filing themselves, while a hardcoded monthly cap (50 images/month) prevents token cost abuse without requiring admin intervention.

## What Changes

- `PATCH /me` accepts `ai_categorisation_enabled` as a writable preference field
- `GET /me` returns `ai_categorisation_count_this_month` — the count of categorisation runs in the current calendar month, derived from `ai_categorisation_logs`
- `CategoriseImage` usecase silently skips the agent call when the user has reached the monthly limit (50); the River job returns `nil` without error
- AdvancedSection in settings gains an AI auto-categorisation toggle with a live usage counter ("X / 50 this month") below the existing AI folder suggestions toggle
- Profile menu shows a small red dot badge on the user avatar when `ai_categorisation_enabled` is true and the monthly limit is hit; badge is dismissed via a localStorage flag keyed to the current year-month

## Capabilities

### New Capabilities

- `ai-categorisation-monthly-limit`: Monthly limit enforcement — new `CountByUserAndMonth` repository method on `ai_categorisation_logs` (no migration needed), hardcoded 50-image monthly limit constant, silent skip in `CategoriseImage` when limit is reached

### Modified Capabilities

- `me-endpoint`: GET /me adds `ai_categorisation_count_this_month`; PATCH /me now accepts `ai_categorisation_enabled` as a writable preference alongside `vision_enabled` and `folder_icons_enabled`
- `fe-vision-toggle`: AdvancedSection gains a second toggle row for AI auto-categorisation with an inline usage counter; the toggle and counter read from the `me` query cache
- `user-profile-menu`: Avatar gains a dismissable red dot badge when the monthly limit is hit; dismissal is persisted to localStorage for the current month

## Impact

- **Backend**: `me.go` (handler request struct + response shape + new usecase dependency), `user_usecase.go` (params struct + UpdatePreferences), `categorisation_usecase.go` (limit check + new CountThisMonth method), `categorisation_log_repository.go` (new CountByUserAndMonth), `main.go` (wiring)
- **Frontend**: `me.ts` (Me type + UpdateMeParams), `AdvancedSection.tsx` (new toggle + counter), `ProfileMenu.tsx` (badge logic)
- **No database migrations** — `ai_categorisation_logs` already has `user_id` and `created_at`
