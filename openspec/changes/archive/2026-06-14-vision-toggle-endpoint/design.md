## Context

`vision_enabled` already exists end-to-end on the read side: column on `users` (migration 000004), `domain.User.VisionEnabled`, and `GET /me` returns it. There is no write path. `AdvancedSection.tsx` renders the toggle but hardcodes `disabled` on the `Switch`.

The codebase has an established partial-update pattern for `PATCH /folders/:id` and `PATCH /images/:id`:
- Handler request struct uses `json.RawMessage` per field to distinguish "field present" from "field absent".
- Handler unmarshals present fields into a typed params struct (e.g. `*string`, `*uuid.UUID`).
- Usecase builds a `fields map[string]any` from non-nil params and calls `repo.Update(ctx, id, ..., fields)`.
- Repository runs `.Model(&T{}).Where(...).Updates(fields)` via GORM.

This change follows that pattern for `/me`, with one simplification: `/me` has no separate `userID` filter since `id` IS the authenticated user's ID (`middleware.AuthenticatedUserIDFromContext`).

## Goals / Non-Goals

**Goals:**
- Let the authenticated user update `vision_enabled` via `PATCH /me`.
- Reuse the existing RawMessage → params → fields-map → `.Updates()` pattern so this PATCH endpoint looks and behaves like the others.
- Keep the writable surface limited to `vision_enabled` — no generic "update any user column" capability.
- Unblock the `AdvancedSection` toggle in the settings modal.

**Non-Goals:**
- No new user-settings resource/namespace (e.g. `/me/settings`). Only one field is writable today; if more user preferences appear later, this endpoint and its `fields` map can grow incrementally.
- No changes to `pending_kinde_deletion` or account-deletion flows.
- No audit logging / history of preference changes.

## Decisions

### 1. `PATCH /me` with `{ "vision_enabled": <raw> }`

`updateMeRequest` declares only `VisionEnabled json.RawMessage`. The handler:
- If absent → `400 Bad Request` ("vision_enabled is required"). Unlike folders/images (which have multiple optional fields and tolerate a no-op PATCH), `/me` currently has exactly one writable field, so an empty body is always a client error rather than a meaningful no-op.
- If present, unmarshal into `bool`. A non-boolean value (string, number, object) → `400 Bad Request` ("vision_enabled must be a boolean").

### 2. Generic `Update(ctx, id string, fields map[string]any) (*domain.User, error)` on `UserRepository`

Mirrors `folder_repository.Update` / `image_repository.Update`:
```go
result := r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", id).Updates(fields)
```
GORM's map-based `.Updates()` writes `false` values (unlike struct-based updates, which skip Go zero values) — important since toggling *off* (`vision_enabled: false`) must actually persist.

Considered a targeted `SetVisionEnabled(ctx, id string, enabled bool) error` instead — rejected because it doesn't pair naturally with the RawMessage/fields-map decode already happening at the handler/usecase layers, and would be the only targeted single-field `Update` on a repo otherwise following the generic-map pattern (folder, image).

### 3. Whitelisting via the request struct

Because `updateMeRequest` only declares `vision_enabled`, the `fields` map passed to `.Updates()` can never contain `id`, `pending_kinde_deletion`, or timestamp columns — there's no separate allowlist to maintain. If a future field is added to `updateMeRequest`, it's automatically the only other key that can appear.

### 4. Response shape matches `GET /me`

`PATCH /me` returns `{ "id": ..., "vision_enabled": ... }` (the post-update row), so the frontend can use the same `Me` type and either replace the `me` query cache directly or refetch.

### 5. Frontend wiring

`AdvancedSection.tsx` gets a `useMutation` calling a new `updateMe(getToken, { vision_enabled })` in `frontend/src/features/auth/lib/me.ts`. On success, the mutation response (the updated `Me`) is written into the `['me']` query cache via `queryClient.setQueryData`, avoiding a refetch. The `Switch` becomes `checked={visionEnabled} onCheckedChange={...} disabled={isPending}` — disabled only while the mutation is in flight, not permanently.

## Risks / Trade-offs

- **[Risk]** Empty/zero `fields` map reaching `.Updates({})` is a GORM no-op that may report `RowsAffected == 0`, which `folder_repository.Update`/`image_repository.Update` treat as `ErrRecordNotFound`. → **Mitigation**: Decision 1 rejects an empty body at the handler before it reaches the usecase/repo, so this path is never hit for `/me`.
- **[Risk]** Toggling `vision_enabled` off while a vision-labelling job is in-flight for that user. → **Mitigation**: out of scope for this change — existing `worker/vision.go` jobs already enqueued will complete; only *new* uploads are affected by the flag (per `vision-api-labelling` capability, unchanged here).

## Open Questions

None — shape confirmed during exploration.
