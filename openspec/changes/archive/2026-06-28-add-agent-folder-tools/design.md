## Context

The agent suggestion service (`backend/internal/agent/`) uses an agentic loop where the LLM calls tools, receives results, and iterates until it calls a submit tool. Currently the only data-fetching tool is `get_folder_list`. This change adds two new read-only tools that give the LLM folder-level context without modifying the loop or the suggestion flow — the tools are built and verified in isolation first.

The vision label data is stored as `jsonb` (`ai_labels`) on the `images` table, serialised as `[]domain.Label{Description string, Score float32}`. An inline struct in `formatImageLabels` currently duplicates this shape — `domain.Label` already exists and is wire-compatible.

`ListByFolder` on `imageRepository` already preloads `Tags []Tag` and `ImageFolders`, so both new tools can get everything they need from a single query each.

## Goals / Non-Goals

**Goals:**
- Implement `get_folder_top_labels` and `get_folder_image_samples` as callable agent tools with correct formatter output
- Extract `extractLabels` helper to eliminate the duplicated label-parsing logic
- Keep `AgentImageRepository` minimal — extend it only with `ListByFolder`
- Tools are verifiable in isolation (correct JSON output) before being wired into the suggestion flow

**Non-Goals:**
- Wiring the tools into `GetFolderSuggestion` (separate change)
- Updating the system prompt to reference the new tools (separate change)
- Pagination or limits at the DB level for `ListByFolder` — folders are expected to stay small enough that fetching all and slicing in Go is acceptable

## Decisions

**D1: Extend `AgentImageRepository` rather than introducing a new interface**

Adding `ListByFolder` to the existing `AgentImageRepository` is the minimal change. Both tools share the same fetch-and-format pattern as `listFolders`, so a second interface would add indirection with no benefit.

**D2: Count tags and labels in Go, not SQL**

`ListByFolder` already preloads `Tags` on each image. A separate aggregation query (`GROUP BY tag_id`) would be cleaner SQL but costs an extra round-trip for data already in memory. Go-side counting is simpler and avoids the extra query. Revisit if folder sizes grow large.

**D3: Slice to 5 images in Go for `get_folder_image_samples`**

`ListByFolder` with `direction="asc"` returns images sorted `created_at ASC`. Taking `[:5]` after the fetch avoids a new repo method or a raw SQL `LIMIT`. Acceptable given expected folder sizes; a dedicated limited query can be added later if needed.

**D4: Extract `extractLabels(aiLabels json.RawMessage, threshold float64) ([]string, error)`**

Both the existing `formatImageLabels` and the new sample formatter need to filter labels by score. Extracting the helper removes the inline struct duplication and lets both formatters share the same parsing path. `domain.Label.Score` is `float32`; the helper casts to `float64` for the threshold comparison.

**D5: `image_notes` maps to `image.Description`**

The agent-facing field name `image_notes` is intentional — it matches the user-facing mental model of the field better than `description` in this context. No domain changes needed.

## Risks / Trade-offs

- **Full folder fetch for aggregation** → For very large folders the in-memory count could be wasteful. Mitigation: acceptable for current scale; the `ListByFolder` call is already bounded by user data size. Add DB-side aggregation later if profiling shows it.
- **`ListByFolder` called twice if both tools are invoked in the same agent turn** → Each tool issues its own query. No deduplication. Mitigation: the agentic loop is sequential and both tools are expected to be called at most once per turn; the cost is low.
- **Label score float32 → float64 cast** → Precision loss is negligible at two decimal places. The threshold const (`0.75`) is well within safe cast range.
