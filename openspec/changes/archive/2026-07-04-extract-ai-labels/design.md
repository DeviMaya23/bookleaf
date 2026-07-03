## Context

Vision labels are currently stored as a JSONB blob in `images.ai_labels`. Every consumer — primarily the agent layer — must fetch the full Image row, unmarshal the JSON via `extractLabels`, and filter/aggregate in Go. This makes labels invisible to SQL and forces all scoring logic into application code.

The goal is to project that same data into a normalised `image_labels` table so that SQL can own filtering and aggregation. The JSONB column is retained as the raw record of what the Vision API returned.

`agent_formatter.go` exists as the LLM presentation layer — it takes data and shapes it into the JSON strings the agent sends to the model. That responsibility does not change. What changes is where the data comes from: instead of receiving a `*domain.Image` and calling `extractLabels` on its JSONB field, formatter functions receive pre-fetched, already-filtered label strings from SQL queries.

---

## Goals / Non-Goals

**Goals:**
- Add `image_labels` table with a SQL migration that also backfills from existing `ai_labels` data
- Dual-write on every new vision labelling run (JSONB + rows), atomically
- Replace `extractLabels` and the Go-side label filtering in the agent layer with SQL JOIN queries — no extra DB round trips vs. today
- Keep `agent_formatter.go` as the LLM presentation layer; only update its function signatures to accept pre-fetched data

**Non-Goals:**
- Changing the Vision API integration or label scoring logic
- Exposing `image_labels` via any API endpoint
- Removing `ai_labels` from the `Image` struct or column (kept as raw backup)
- Frontend or extension changes

---

## Decisions

### 1. Dual-write is a single repo method, not a usecase-level transaction

`ProcessVisionLabelling` currently calls `UpdateAILabels` (one repo method, one query). Rather than coordinating two repo calls inside the usecase, a new single method — `UpdateLabels(ctx, imageID, rawJSON, []domain.ImageLabel)` — wraps both the JSONB update and the `image_labels` inserts inside `r.db.Transaction`. This keeps transaction scope inside the repository layer, consistent with how all other multi-write operations work in this codebase (see `imageRepository.SyncImageFolders`, `pendingUploadRepository.Transaction`).

`UpdateAILabels` is removed from the `UploadImageRepository` interface; `UpdateLabels` replaces it.

### 2. Backfill in SQL migration, not Go

`jsonb_array_elements` can unnest the existing `ai_labels` arrays directly inside the migration:

```sql
INSERT INTO image_labels (image_id, label, score)
SELECT images.id,
       elem->>'Description',
       (elem->>'Score')::float4
FROM images,
     LATERAL jsonb_array_elements(ai_labels) AS elem
WHERE ai_labels IS NOT NULL
  AND jsonb_typeof(ai_labels) = 'array';
```

The `jsonb_typeof` guard handles any rows where `ai_labels` is a non-array value. No Go backfill job needed.

### 3. Agent repo methods are JOIN queries — no extra round trips

The three agent reads become three SQL methods on `AgentImageRepository`:

| Old (Go-side) | New (SQL-side) |
|---|---|
| `GetByID` → `extractLabels(img.AILabels, threshold)` | `GetImageWithLabels(imageID, userID, threshold)` — LEFT JOIN on `image_labels`, returns `*domain.Image` + filtered label strings |
| `ListByFolder` → Go label/tag frequency counts, `len(images)` | `GetFolderTopLabels(userID, folderID, threshold, topN)` — three queries (image count, vision label aggregation, user tag aggregation) in one repo method, returns `*domain.FolderAggregate` |
| `ListByFolder` (desc, slice to 5) → `extractLabels` per image | `GetFolderImageSamples(userID, folderID, threshold, limit)` — two queries: fetch top-N images by `created_at DESC`, then fetch their labels in one `WHERE image_id IN (...)` batch |

`GetByID` and `ListByFolder` are both removed from `AgentImageRepository`. `GetFolderTopLabels` returns `*domain.FolderAggregate` (a query read-model in `internal/domain/` — placed there to avoid an import cycle, since `repository` already imports `usecase` which imports `agent`).

`GetFolderImageSamples` uses two queries (not one) to avoid a complex `array_agg` with GORM. Since limit is always 5, this is a fixed two-query cost — same total round trips as today's single `ListByFolder` call.

### 4. `agent_formatter.go` is kept; only `extractLabels` is removed

`extractLabels` is the only function in `agent_formatter.go` that touches `json.RawMessage` / JSONB. It is removed. The three formatter functions that called it — `formatImageLabels`, `formatFolderTopLabels`, `formatFolderImageSamples` — are retained but their signatures change:

| Function | Old signature | New signature |
|---|---|---|
| `formatImageLabels` | `(image *domain.Image, threshold float64)` | `(title string, labels []string)` |
| `formatFolderTopLabels` | `(folderID, folder, images []*domain.Image, threshold)` | `(folderID, folder, imageCount int, topVisionLabels []string, tagCount map[string]int)` |
| `formatFolderImageSamples` | `(images []*domain.Image, threshold float64)` | `(images []*domain.Image, labelMap map[uuid.UUID][]string)` |

All other helpers in `agent_formatter.go` (`formatFolderList`, `findFolder`, `getFolderPath`, `isDeniedSourceURL`, `sourceURLDenyList`) are unchanged.

### 5. `ListUnlabelled` stays as `ai_labels IS NULL`

The JSONB column remains the canonical signal that the vision job has run. Switching to `NOT EXISTS (SELECT 1 FROM image_labels ...)` adds complexity for no correctness gain — both are written atomically in the same transaction.

---

## Risks / Trade-offs

- **Dual-write divergence**: if a bug causes the transaction to partially succeed, `ai_labels` and `image_labels` could drift. Mitigation: single DB transaction wraps both writes; the backfill migration seeds existing data before any new writes happen.
- **Backfill migration time**: `jsonb_array_elements` over the full `images` table could be slow for large datasets. Mitigation: add `CREATE INDEX` after the insert so the index build doesn't slow the backfill; migration runs once at deploy.
- **JSON key casing**: `domain.Label` has no json struct tags, so Go marshals `Description` and `Score` with capital letters. The backfill SQL uses `elem->>'Description'` and `elem->>'Score'` to match. This is a hidden coupling — low risk since the backfill runs once.

---

## Migration Plan

1. Deploy migration: creates `image_labels` table, backfills from `ai_labels`, adds index on `image_id`
2. Deploy application code: `ProcessVisionLabelling` dual-writes; agent reads from `image_labels`; formatter receives pre-fetched data
3. No rollback complexity — `ai_labels` is unchanged throughout; reverting the code drop leaves the app reading JSONB again

---

## Open Questions

- None — all decisions resolved during exploration.
