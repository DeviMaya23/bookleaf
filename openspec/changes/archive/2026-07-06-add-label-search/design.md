## Context

Search by image title (`GET /images?name=...`) is already implemented across `handler → usecase → repository`. The `image_labels` table exists and is populated by the vision pipeline, but is currently only queried on the labels side (frequency aggregation, image detail). The `domain.Image` struct has no `Labels []ImageLabel` association.

On the frontend, `useGalleryControls` manages all search/filter/sort state and is instantiated in `AppLayout`. `me.vision_enabled` is currently only consumed in `useVisionSuggestion` (SSE folder suggestion) and the settings modal — it is not available in `AppLayout` or the gallery hooks today.

## Goals / Non-Goals

**Goals:**
- Extend title search to also match `image_labels.label` when `search_labels=true`
- Surface a toggle in the filter dropdown, visible only when `vision_enabled = true`
- Keep the change additive and self-contained — no regressions to existing filters, sort, or pagination

**Non-Goals:**
- Label search in the trash view
- Label search in folder views (client-side filtered; labels are not in the image list payload)
- Returning matched label data in the response (surfacing *why* an image matched)
- Standalone label-only search without a name term

## Decisions

### 1. OR EXISTS subquery, no domain association

The label filter is implemented as an `OR EXISTS` subquery in `image_repository.List`, consistent with how `folderIDs` and `tagIDs` are already handled:

```sql
WHERE (
  images.title ILIKE '%term%'
  OR EXISTS (
    SELECT 1 FROM image_labels
    WHERE image_id = images.id
    AND label ILIKE '%term%'
  )
)
```

**Why not a JOIN?** A JOIN on `image_labels` would produce multiple rows per image (one per label), breaking `LIMIT` semantics and requiring a `DISTINCT` or subquery anyway.

**Why not add `Labels []ImageLabel` to `domain.Image`?** The label association is only needed if label data is returned in the list response. For pure filtering it adds unnecessary preload cost on every request. `ImageLabel` stays standalone.

The existing `name != nil && *name != ""` guard in the repository naturally makes label search a no-op when no search term is entered — no additional guard needed.

### 2. `searchLabels bool` not `*bool` on `ListImagesParams`

`unfiled bool` is already in the same struct using a plain bool. `searchLabels bool` is consistent — `false` is the unambiguous zero value meaning "don't search labels."

The repository interface signature gains a matching `searchLabels bool` parameter, placed after `name *string`:

```go
List(ctx, userID string, unfiled bool, folderIDs []uuid.UUID, tagIDs []uuid.UUID,
     mimeTypes []string, name *string, searchLabels bool,
     sortField *string, direction *string, cursor *ImageCursor, limit int)
```

### 3. `vision_enabled` sourced in AppLayout, passed as param to `useGalleryControls`

`AppLayout` is the natural owner of `me` for the gallery — it already owns `tags` and `folders` queries. Adding a `getMe` query there (same pattern as `useVisionSuggestion`) and passing `visionEnabled bool` into `useGalleryControls` keeps the hook pure and testable without internal data fetching.

**Alternative considered**: fetch inside `useGalleryControls` directly. Rejected because it couples the hook to a data source, making it harder to test and inconsistent with how tags/folders are passed in.

### 4. Toggle lives in the filter dropdown, not inline with search

The toolbar row is already cluttered. The filter dropdown has an established extensibility pattern (`filterSections` array gated by view type). A new `'labelSearch'` section type added to `filterSectionsForViewType` for `all` and `unsorted` views (when `visionEnabled`) follows the same shape as existing sections.

**Filter count**: the label search toggle does NOT increment `filterCount` (the badge on the Filters button). It's a search-scope modifier, not a data filter — incrementing the count would be misleading.

### 5. `searchLabels` excluded from folder and trash views

- **Folder view**: search is client-side only; labels are not in the list payload. Label search here would require either a separate fetch or server-side search for folder contents — out of scope.
- **Trash view**: not a primary search surface for labelled content; excluded per proposal.

`filterSectionsForViewType` already returns `[]` for `trash` and has no `labelSearch` section — no change needed there.

## Risks / Trade-offs

- **Query performance on large `image_labels` tables**: The `EXISTS` subquery with `ILIKE` is not index-aided (trigram index would help but is out of scope). For typical moodboard library sizes this is acceptable. → No mitigation needed now; a `pg_trgm` index on `image_labels.label` is a future optimization if needed.
- **`image-endpoints` spec interface change**: Adding `searchLabels bool` to `List` signature is a breaking change to the interface contract. All callers (only `imageUsecase`) must be updated. → Low risk — single caller, caught at compile time.

## Open Questions

None — all decisions resolved during exploration.
