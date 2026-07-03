## Why

Vision labels are currently stored as a JSONB blob in `images.ai_labels`, which makes them opaque to SQL — every consumer must unmarshal and filter in Go. Normalising them into an `image_labels` table makes labels directly queryable, enabling SQL-side filtering and aggregation and eliminating the JSONB unmarshaling work in the agent layer.

## What Changes

- Add a new `image_labels` table (`image_id`, `label`, `score`) with a SQL migration that also backfills existing data from `ai_labels` using `jsonb_array_elements`
- Add a new `ImageLabel` domain type and GORM model
- `ProcessVisionLabelling` dual-writes: continues to update `ai_labels` (kept as raw backup) and additionally inserts rows into `image_labels`, wrapped in a transaction
- `AgentImageRepository` gains three new JOIN-based query methods; the agent service calls these to fetch pre-resolved label data and passes it to the existing formatter functions
- `extractLabels` is removed from `agent_formatter.go` — it is the only function in the formatter that touches JSONB; the formatter functions themselves are retained and have their signatures updated to accept pre-fetched label data instead of raw `*domain.Image`

## Capabilities

### New Capabilities

- `image-label-table`: `image_labels` table, `ImageLabel` domain type, migration (including SQL backfill), and repository write method called by `ProcessVisionLabelling`

### Modified Capabilities

- `vision-api-labelling`: `ProcessVisionLabelling` now dual-writes to `image_labels` in addition to `ai_labels`; both writes are wrapped in a transaction
- `agent-folder-context-tools`: `extractLabels` removed; `AgentImageRepository` gains three new SQL-backed methods; formatter function signatures updated to take pre-fetched data

## Impact

- **Backend — domain**: new `ImageLabel` struct in `internal/domain/`
- **Backend — repository**: new `UpdateLabels` write method on `imageRepository`; new `GetImageWithLabels`, `GetFolderTopLabels`, and `GetFolderImageSamples` methods added to `AgentImageRepository` interface and implemented on `imageRepository`
- **Backend — usecase**: `ProcessVisionLabelling` wraps its two write calls in a transaction
- **Backend — agent**: `extractLabels` removed from `agent_formatter.go`; `formatImageLabels`, `formatFolderTopLabels`, `formatFolderImageSamples` signatures updated; `agent_service.go` calls new repo methods and passes results to formatter
- **Migration**: new numbered migration file (`image_labels` table + `jsonb_array_elements` backfill)
- **No frontend or extension changes**
