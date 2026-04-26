## Why

The project no longer uses BYOS/BYOV — storage and Vision API are now app-provided. The only app-specific state that belongs on the `User` domain entity is the opt-in flag for AI organising (`vision_enabled`). The current `User` struct is a placeholder that lacks this field.

## What Changes

- Add `vision_enabled bool` column to the `users` table (default `false`)
- Add `VisionEnabled` field to the `User` GORM struct
- Add a `golang-migrate` migration to introduce the new column on existing databases

## Capabilities

### New Capabilities

_None — this change adds a field to an existing capability._

### Modified Capabilities

- `user-domain`: Add `vision_enabled` field requirement to the User struct and its migration

## Impact

- `internal/domain/user.go` — add `VisionEnabled` field
- New `golang-migrate` migration file (up: `ALTER TABLE users ADD COLUMN vision_enabled BOOLEAN NOT NULL DEFAULT false`, down: `ALTER TABLE users DROP COLUMN vision_enabled`)
- No handler, usecase, or repository changes in scope — this change is domain + migration only
