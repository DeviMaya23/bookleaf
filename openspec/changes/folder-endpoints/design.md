## Context

The folder domain struct and DB migration already exist (`000002_create_folders`). The existing struct uses GORM soft-delete (`DeletedAt gorm.DeletedAt`) and the migration includes a `deleted_at` column. The project uses Echo + GORM with a handler → usecase → repository layering, identical to the existing `me-endpoint` + user module pattern. Auth is handled via Kinde JWT middleware that injects the authenticated user ID onto the Echo context.

Images currently have `ON DELETE SET NULL` for `folder_id` at the DB level. Child folders have `ON DELETE RESTRICT` on `parent_id`. Both constraints will be changed to `ON DELETE RESTRICT` so the DB acts as a safety net against direct SQL manipulation, while all cascade null-out logic lives explicitly in the application.

## Goals / Non-Goals

**Goals:**
- CRUD HTTP endpoints at `/folders` scoped to the authenticated user
- Hard delete with explicit application-level cascade: null out child folders' `parent_id` and images' `folder_id` before deleting; DB FK constraints are `ON DELETE RESTRICT` on both as a safety net
- Remove soft-delete from the domain struct and DB schema (new migration, not modify existing)

**Non-Goals:**
- Recursive/nested folder tree queries (flat list only)
- Folder sharing or multi-user access
- Image operations (images module not yet implemented)
- Pagination or filtering on list endpoint

## Decisions

### D1: Follow existing handler/usecase/repository layering

**Decision**: New folder files live in `internal/handler/folder.go`, `internal/usecase/folder_usecase.go`, `internal/usecase/folder_repository.go` (interface), `internal/repository/folder_repository.go`.

**Why over a `folder/` sub-package**: Stays consistent with the existing flat-package structure (`handler`, `usecase`, `repository`). Avoids introducing a new structural pattern mid-project.

### D2: All cascade null-outs are explicit in the usecase; DB FKs are RESTRICT for both

**Decision**: Before deleting a folder, the usecase calls repository methods that:
1. `UPDATE folders SET parent_id = NULL WHERE parent_id = $folderID`
2. `UPDATE images SET folder_id = NULL WHERE folder_id = $folderID`
3. `DELETE FROM folders WHERE id = $folderID AND user_id = $userID`

All three steps run in a single transaction. Both the `folders.parent_id` and `images.folder_id` FK constraints are `ON DELETE RESTRICT`.

**Why**: Keeping both null-outs in application code makes the delete behaviour fully readable and testable without consulting the DB schema. `ON DELETE RESTRICT` on both FKs ensures the DB will reject any direct SQL deletion that bypasses the application, acting as a safety net against manual or accidental data manipulation.

### D3: Remove soft-delete via a new migration (000005)

**Decision**: Add `000005_remove_folders_soft_delete` migration that drops `deleted_at` and its index from `folders`. Also remove `DeletedAt gorm.DeletedAt` from the `Folder` struct.

**Why over modifying 000002**: Migration 000002 may already have been applied on developer machines. A new migration is safe to run incrementally; modifying 000002 requires a full teardown.

### D4: All folder queries scoped by userID

**Decision**: Every repository method takes a `userID string` parameter and includes `WHERE user_id = $n` in all queries (including the cascade UPDATE before delete).

**Why**: Prevents one user from reading or modifying another user's folders. The authenticated userID is extracted from the Echo context in the handler and passed down through the usecase.

## Risks / Trade-offs

- **Stale GORM soft-delete queries**: GORM automatically appends `deleted_at IS NULL` to queries when the model has `DeletedAt gorm.DeletedAt`. After the field is removed, any queries that relied on that implicit filter will return previously soft-deleted rows (which don't exist yet in this project). Low risk at this stage, but worth noting. → Mitigation: the migration drops the column entirely so no stale rows can exist.
- **Race condition on delete**: Between the cascade UPDATEs and the DELETE, another request could create a new child folder or image. → Mitigation: all three statements run in a single transaction, so the DB sees an atomic delete.

## Migration Plan

1. Remove `DeletedAt gorm.DeletedAt` from `internal/domain/folder.go`
2. Add `000005_remove_folders_soft_delete.up.sql` / `.down.sql` — drops `deleted_at` column and index from `folders`, and changes `images.folder_id` FK from `ON DELETE SET NULL` to `ON DELETE RESTRICT`
3. Re-run `make migrate-up` locally
4. Implement repository, usecase, handler, and wire into `main.go`
