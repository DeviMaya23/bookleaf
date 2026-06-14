## 1. Backend - Repository

- [x] 1.1 Add `Update(ctx context.Context, id string, fields map[string]any) (*domain.User, error)` to `UserRepository` interface (`internal/usecase/user_repository.go`)
- [x] 1.2 Implement `Update` in `internal/repository/user_repository.go`: `.Model(&domain.User{}).Where("id = ?", id).Updates(fields)`, map `RowsAffected == 0` to `gorm.ErrRecordNotFound`, then return via `GetByID`

## 2. Backend - Usecase

- [x] 2.1 Add `UpdateVisionEnabled(ctx context.Context, id string, enabled bool) (*domain.User, error)` to `userUsecase`, building `fields := map[string]any{"vision_enabled": enabled}` and calling `userRepo.Update`
- [x] 2.2 Add `UpdateVisionEnabled` to the `UserUsecase` interface in `internal/handler/me.go`
- [x] 2.3 Unit test: `UpdateVisionEnabled` success (asserts returned user's `VisionEnabled`)
- [x] 2.4 Unit test: `UpdateVisionEnabled` repo error is propagated

## 3. Backend - Handler & Routing

- [x] 3.1 Add `updateMeRequest` struct (`VisionEnabled json.RawMessage`) to `internal/handler/me.go`
- [x] 3.2 Implement `UpdateMe` handler: bind request; if `VisionEnabled` absent → `400`; if present but not a JSON bool → `400`; otherwise call `usecase.UpdateVisionEnabled` and return `200` with `{ id, vision_enabled }`
- [x] 3.3 Register `PATCH /me` → `meHandler.UpdateMe` in `cmd/server/main.go` (protected group)
- [x] 3.4 Unit test: `PATCH /me` with `{"vision_enabled": true}` → `200`, body reflects updated value
- [x] 3.5 Unit test: `PATCH /me` with `{"vision_enabled": false}` → `200`, body reflects updated value
- [x] 3.6 Unit test: `PATCH /me` with `{}` → `400`
- [x] 3.7 Unit test: `PATCH /me` with `{"vision_enabled": "yes"}` → `400`
- [x] 3.8 Unit test: usecase error → `500`

## 4. Bruno

- [x] 4.1 Add `bruno/update-me.bru` (`PATCH {{baseUrl}}/me`, body `{ "vision_enabled": true }`, auth inherit)

## 5. Frontend

- [x] 5.1 Add `updateMe(getToken, { vision_enabled }): Promise<Me>` to `frontend/src/features/auth/lib/me.ts`
- [x] 5.2 In `AdvancedSection.tsx`, add a `useMutation` calling `updateMe`; on success, `queryClient.setQueryData(['me'], ...)` with the response
- [x] 5.3 Wire `Switch` to `checked={visionEnabled} onCheckedChange={...} disabled={isPending}`, removing the hardcoded `disabled`
- [x] 5.4 Update `AdvancedSection.test.tsx`: toggling on/off calls `PATCH /me` and updates the displayed state; switch is disabled while pending
- [x] 5.5 On mutation error, show `toast.error('Failed to update settings')` (matches `AccountSection` toast pattern)

## 6. Verification

- [x] 6.1 Run `golangci-lint run` from `backend/` and fix any issues
- [x] 6.2 Run `npm run build` and `npm run lint` from `frontend/` and fix any issues
