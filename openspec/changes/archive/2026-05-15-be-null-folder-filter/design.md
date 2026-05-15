## Context

`GET /images` currently reads `folder_id` from the query string and parses it as a UUID. If the param is absent or empty, no folder filter is applied and all images are returned. There is no supported way to request only unfoldered images (WHERE folder_id IS NULL).

The change touches three layers: handler (read the new param), usecase params (carry the flag), and repository (emit the correct WHERE clause).

## Goals / Non-Goals

**Goals:**
- `unfiled=true` on `GET /images` returns only images where `folder_id IS NULL`
- `unfiled=false` or absent preserves existing behavior
- When `unfiled=true`, `folder_id` param is ignored

**Non-Goals:**
- Applying `unfiled` to any other endpoint (`GET /images/trash`, etc.)
- Changing existing `folder_id` UUID parsing behavior

## Decisions

### 1. Add `Unfiled bool` to `ListImagesParams` instead of restructuring `FolderID`

`ListImagesParams.FolderID` stays as `*uuid.UUID` (nil = no filter, non-nil = filter by folder). A new `Unfiled bool` field is added alongside it. When `Unfiled = true`, the repository ignores `FolderID` and emits `WHERE folder_id IS NULL` instead.

**Alternative considered**: Tri-state via a `FolderFilter` struct with `Apply bool` and `FolderID *uuid.UUID`. Rejected — two fields that interact implicitly are harder to read than a single, named boolean.

### 2. Handler reads `unfiled` param and sets `Unfiled` on params

```go
if c.QueryParam("unfiled") == "true" {
    params.Unfiled = true
}
```

`folder_id` parsing is left unchanged and runs regardless, but the repository ignores it when `Unfiled = true`.

### 3. Repository checks `Unfiled` first

```go
if params.Unfiled {
    query = query.Where("folder_id IS NULL")
} else if params.FolderID != nil {
    query = query.Where("folder_id = ?", *params.FolderID)
}
```

## Risks / Trade-offs

- **`ListImagesParams` is a public type** → all call sites (handler, tests) must be updated. Mitigation: grep for all usages before implementing.
- **`unfiled=true` and `folder_id=<uuid>` sent together** → `unfiled` wins and `folder_id` is silently ignored. This is intentional and documented in the spec.
