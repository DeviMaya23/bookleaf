## Context

The backend lives under `github.com/devi/bookleaf` (module root). All packages
being moved are within this single module, so no external consumers are affected —
only internal import paths change.

Current state that needs correcting:

| Package | Current path | Target path |
|---|---|---|
| config | `internal/config` | `internal/platform/config` |
| observability | `internal/observability` | `internal/platform/observability` |
| middleware | `internal/middleware` | `internal/handler/middleware` |
| thumbnail | `internal/thumbnail` | `pkg/thumbnail` |
| StorageService interface | `internal/storage` | `internal/usecase` |
| ThumbnailService interface | `internal/thumbnail` | `internal/usecase` |
| VisionService interface | `internal/vision` | `internal/usecase` |

Files affected by import path changes: ~20 Go files across `handler/`, `usecase/`,
`repository/`, `middleware/`, and `cmd/server/`.

## Goals / Non-Goals

**Goals:**
- Bring directory layout into alignment with CONVENTIONS.md
- Move interfaces to consumer side (usecase owns all its dependencies' contracts)
- Establish `internal/platform/` as the home for cross-cutting infrastructure
- Move `thumbnail` to `pkg/` as the first genuinely reusable utility

**Non-Goals:**
- Changing any behavior — this is a mechanical move with import updates only
- Refactoring logic inside any moved package
- Adding, removing, or modifying any interfaces beyond relocation

## Decisions

### Interface files in usecase/

`StorageService` has 6 methods and warrants its own file (`usecase/image_storage.go`).
`ThumbnailService` and `VisionService` each have 1 method — they are small enough
to inline at the top of `image_usecase.go` rather than adding separate files.

After moving, `internal/storage/`, `internal/vision/`, and `pkg/thumbnail/` become
pure implementation packages with no interface declarations.

### thumbnail moves to pkg/, storage and vision stay in internal/

`thumbnail` has zero app-specific knowledge — it takes an `io.Reader`, returns an
`io.Reader`, and depends on nothing in this codebase. It qualifies for `pkg/`.

`storage` and `vision` reference app config and credentials; they stay in `internal/`.

### One package move per task

Each directory move is its own task. This keeps diffs reviewable and makes it easy
to stop or roll back mid-refactor without leaving the codebase in a broken state.
Import path updates are bundled with the move that causes them.

## Risks / Trade-offs

[Build breaks mid-refactor] → Each task must leave the build passing before
the next one starts. Run `go build ./...` after every move.

[Missed import references] → `go build ./...` will catch any stale imports.
No manual grep needed as the compiler is authoritative.

[`internal/` visibility rules for pkg/thumbnail] → Moving thumbnail out of
`internal/` means it could theoretically be imported by code outside this module.
This is acceptable — it is intentionally a general-purpose utility.

## Migration Plan

No deployment steps required. This is a compile-time refactor with no runtime,
database, or API changes. Changes take effect when merged.

Rollback: revert the commits. No state to unwind.
