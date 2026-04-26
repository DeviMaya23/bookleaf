## Context

The Echo server is running but `internal/domain/` is empty. No database connection exists yet. This change lays the domain layer foundation — GORM structs and SQL migrations — so repository implementations can be written in the next change.

## Goals / Non-Goals

**Goals:**
- GORM structs for `User`, `Folder`, `Image` in `internal/domain/`
- `golang-migrate` wired up with a `migration/` directory
- Up/down SQL migrations for `users`, `folders`, `images` tables
- UUIDs as primary keys for `folders` and `images`; `users` uses Clerk's string ID as PK

**Non-Goals:**
- Database connection / GORM `AutoMigrate` (belongs in server wiring change)
- Repository interfaces or implementations
- Any HTTP handlers or usecases
- BucketConfig model (separate concern, separate change)

## Decisions

### Decision 1: UUIDs as primary keys for Folder and Image; Clerk string ID for User

`Folder` and `Image` use `uuid.UUID` (from `github.com/google/uuid`) as PKs — they may be referenced across user contexts (shared links) and UUIDs avoid enumeration. GORM's `BeforeCreate` hook populates the UUID before insert.

`User.ID` is Clerk's generated user ID (e.g. `user_abc123`), a `TEXT` primary key with no separate `uuid` field. Clerk owns the identity layer; duplicating a UUID alongside it adds no value.

### Decision 2: golang-migrate with SQL files, not GORM AutoMigrate

`golang-migrate` with plain SQL gives explicit, reviewable, reversible migrations. `AutoMigrate` is convenient but doesn't support down migrations and can silently drop columns. SQL files in `migration/` with the naming convention `000001_create_users.up.sql` / `000001_create_users.down.sql`.

### Decision 3: AILabels as JSONB

Store AI-generated labels as `jsonb` in the `ai_labels` column rather than a separate `image_labels` table. Rationale: labels are read-only after write, never queried individually, and BYOV means many users won't have them at all. JSONB avoids a join on the hot gallery query path.

### Decision 4: Migration order — users → folders → images

`folders` references `users`; `images` references both `folders` and `users`. Migration numbers must reflect this: `000001_create_users`, `000002_create_folders`, `000003_create_images`.

### Decision 5: User struct is minimal for now

`User` only needs `ID` (Clerk's string ID, the PK). Name, email, and avatar come from Clerk at runtime and don't need to be mirrored in the DB. No separate UUID field exists on `users`.

`user_id` FK columns on `folders` and `images` are `TEXT NOT NULL` referencing `users(id)`.

### Decision 6: All DB columns use snake_case

All column names follow `snake_case` convention (enforced by GORM's default naming strategy). Go struct field names follow Go conventions (`UserID`, `R2Path`, `AILabels`, etc.) and map to their `snake_case` equivalents (`user_id`, `r2_path`, `ai_labels`).

## Risks / Trade-offs

- [Self-referencing FK on folders] PostgreSQL handles this fine, but deleting a parent folder without cascading will fail. → Mitigation: add `ON DELETE RESTRICT` for now; a "delete folder + re-home images" usecase can be added later.
- [JSONB for AILabels] Not queryable with standard GORM methods. → Acceptable: AI labels are display-only in MVP scope; no filtering by label is required yet.
