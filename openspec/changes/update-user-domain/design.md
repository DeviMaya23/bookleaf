## Context

The `User` struct in `internal/domain/user.go` is a minimal shadow record anchored to Clerk's user ID. The project has moved away from BYOS/BYOV; the only app-owned state on a user is whether they have opted into AI organising. This requires adding a single boolean column to the `users` table.

The existing migration that created the `users` table is already applied; this change adds a new migration for the `vision_enabled` column.

## Goals / Non-Goals

**Goals:**
- Add `VisionEnabled bool` to the `User` GORM struct
- Add an additive SQL migration for the new column with a safe default

**Non-Goals:**
- Any handler, usecase, or repository logic that reads/writes `vision_enabled` — that belongs to the AI organising feature change
- Storage limits or per-user quotas

## Decisions

### Boolean column with DB-level default

`vision_enabled BOOLEAN NOT NULL DEFAULT false` rather than nullable.

Rationale: the field has a clear default (off) for all users, including any already in the database. `NOT NULL DEFAULT false` is safer and simpler than nullable — avoids nil checks in Go and makes the intent explicit.

### Additive migration only

No modifications to the existing users migration. A new numbered migration file adds the column via `ALTER TABLE`.

Rationale: the original migration may already be applied in dev/prod. Altering it would break idempotency. An additive migration is the standard golang-migrate pattern.

## Risks / Trade-offs

- [Zero risk for new databases] The column is added with a default, so existing rows are populated automatically.
- [Migration naming] Migration files must be numbered correctly to run after the initial users migration. File naming must follow the existing convention in the migrations directory.

## Migration Plan

1. Apply up migration: `ALTER TABLE users ADD COLUMN vision_enabled BOOLEAN NOT NULL DEFAULT false`
2. Rollback: `ALTER TABLE users DROP COLUMN vision_enabled`

No data backfill needed — default `false` covers all existing rows.
