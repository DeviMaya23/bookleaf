## Why

`vision_enabled` controls whether Bookleaf's AI folder-suggestion-on-upload feature runs for a user. It's currently readable via `GET /me`, but there's no way to change it — the settings modal's AI toggle (`AdvancedSection`) renders the current value but is hardcoded `disabled`. Users need a way to opt in or out of this AI feature.

## What Changes

- Add a `PATCH /me` endpoint that accepts `{ "vision_enabled": <bool> }` and updates the authenticated user's `vision_enabled` flag, returning the updated user in the same shape as `GET /me`.
- Add a generic `Update(ctx, id string, fields map[string]any) (*domain.User, error)` method to `UserRepository`, mirroring the existing `folder_repository.Update` / `image_repository.Update` pattern (`.Model(&User{}).Where("id = ?", id).Updates(fields)`).
- Add a corresponding usecase method that decodes the request's `json.RawMessage` field(s) into a `fields map[string]any`, validating `vision_enabled` is a bool.
- Wire up `AdvancedSection.tsx`'s AI toggle to call `PATCH /me` on change (add an `updateMe`/`patchMe` function to `frontend/src/features/auth/lib/me.ts`, use a mutation that updates the `me` query), and remove the `disabled` prop from the `Switch`.

## Capabilities

### New Capabilities
- `fe-vision-toggle`: Settings modal AI section toggle that lets the user enable/disable `vision_enabled` via `PATCH /me`, with the switch reflecting the persisted value and pending/error states during the mutation.

### Modified Capabilities
- `me-endpoint`: Add a `PATCH /me` requirement — authenticated users can update their `vision_enabled` flag via a partial update, with validation and the updated user returned in the response.

## Impact

- **Backend**: `internal/handler/me.go` (new `UpdateMe` handler + request struct), `internal/usecase/user_usecase.go` + `internal/usecase/user_repository.go` (new usecase method + repo interface method), `internal/repository/user_repository.go` (new `Update` implementation), `cmd/server/main.go` (register `PATCH /me` route).
- **Frontend**: `frontend/src/features/auth/lib/me.ts` (new `updateMe` function), `frontend/src/features/settings/components/AdvancedSection.tsx` (wire toggle to mutation, remove `disabled`).
- **Bruno**: new request file for `PATCH /me`.
