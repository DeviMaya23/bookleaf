## Why

The `internal/` directory mixes architectural layers (domain, usecase, repository, handler) with infrastructure adapters and cross-cutting concerns at the same level, making the architecture harder to read and giving agents no clear rule for where new packages belong. CONVENTIONS.md now defines the standard — this change brings the codebase into alignment with it.

## What Changes

- Move `internal/config/` → `internal/platform/config/`
- Move `internal/observability/` → `internal/platform/observability/`
- Move `internal/middleware/` → `internal/handler/middleware/`
- Move `internal/thumbnail/` → `pkg/thumbnail/`
- Move `StorageService`, `ThumbnailService`, and `VisionService` interfaces from their respective packages into `internal/usecase/` (consumer side), leaving only implementations in `storage/`, `thumbnail/`, and `vision/`
- Update all import paths across the codebase

## Capabilities

### New Capabilities

None. This is a structural refactor — no new domain capabilities are introduced.

### Modified Capabilities

None. No spec-level behavior changes. All existing functionality is preserved.

## Impact

- All Go files that import `internal/config`, `internal/observability`, or `internal/middleware` need import path updates
- `pkg/thumbnail` is a new top-level package; the module path changes accordingly
- Interface definitions for storage, thumbnail, and vision move to `usecase/`; implementing packages are unaffected but callers referencing the interface type need updating
- No API contract changes, no database changes, no frontend impact
