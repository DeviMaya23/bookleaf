## Context

Vision labelling stores AI-generated labels on `Image.AILabels` after upload. `CategorisationUsecase.CategoriseImage` and `AgentService.GetFolderSuggestion` already exist and have been polished, but are dead code — nothing enqueues them. This change wires the full pipeline: vision success → River job → agent call → log → folder assignment.

Current state worth noting:
- `ProcessVisionLabelling` already fetches the user and checks `vision_enabled` before doing work — the new flag check follows the same pattern
- `AgentService` hardcodes `anthropic.ModelClaudeHaiku4_5` — needs to be a config-driven field
- Latest migration is `000018`. New migrations will be `000019` and `000020`
- `AnthropicAPIKey` is a top-level field on `Config` (not grouped into a sub-struct)

## Goals / Non-Goals

**Goals:**
- Enqueue a categorise job after every successful vision labelling for users with the flag enabled
- Persist the agent's reasoning to `ai_categorisation_logs` for every run (success and partial failure)
- Make the Anthropic model configurable via env var without a code change
- Keep the feature invisible to users for now (no UI, flag defaults to false)

**Non-Goals:**
- Any frontend changes or new HTTP endpoints
- A UI toggle for `ai_categorisation_enabled`
- Exposing the log table via API (deferred)
- Retry-on-agent-error strategies beyond standard River behaviour

## Decisions

### Already-in-folder guard: skip if image has an explicit placement

`CategoriseImage` fetches the image early and returns nil if `FolderID` is already set. This prevents the agent from overriding a folder the user explicitly chose at upload time. The check happens after the idempotency check (so a prior log entry is still reusable if needed) but before the agent is called.

### Flag check location: enqueue time, not worker time

The `ai_categorisation_enabled` check happens in `ProcessVisionLabelling` before enqueuing — not inside `CategoriseImage`. This avoids inserting jobs that will immediately no-op, mirrors how `vision_enabled` is checked before enqueuing the vision job (not inside the worker), and keeps `CategorisationUsecase` free of a `userRepo` dependency it doesn't otherwise need. Ownership is already verified inside `AgentService.GetFolderSuggestion` via `imageRepo.GetByID(ctx, imageID, userID)`.

### Log persistence timing: immediately after agent returns; doubles as idempotency record

`ai_categorisation_logs` is written right after `GetFolderSuggestion` returns a result, before the folder create/assign steps. This ensures the reasoning is recorded even if the subsequent DB operations fail.

The log table also serves as an idempotency guard. `CategoriseImage` checks `logRepo.GetByImageID` before calling the agent — if an entry exists from a prior attempt, the LLM call is skipped entirely and the persisted suggestion is used to proceed with folder assignment. This means River retries caused by downstream failures (folder creation, DB errors) do not re-invoke the Anthropic API or incur additional token cost.

### Log table: no FK cascade on image deletion

`ai_categorisation_logs.image_id` uses `ON DELETE SET NULL`. The log is an audit trail that may eventually be surfaced to users — it should survive image deletion so the history isn't silently lost.

### Anthropic model: top-level config field with default

`ANTHROPIC_MODEL` is added as a top-level `string` field on `Config` (alongside `AnthropicAPIKey`), defaulting to `claude-haiku-4-5-20251001`. This is consistent with how `AnthropicAPIKey` is structured today — no new sub-struct. `NewAgentService` receives the model string and stores it on the struct.

### Retry policy for CategorisationWorker

3 attempts, 30-second fixed retry delay (via `NextRetry` override). Longer than the vision job's 10s because LLM errors are typically rate limits or transient API issues that benefit from a longer cooldown. 3 attempts matches the vision job's limit.

## Risks / Trade-offs

**LLM cost per retry** — mitigated by the idempotency check: retries caused by downstream failures reuse the persisted log entry and do not call the Anthropic API again. Only a failure before the log is written (i.e. the agent call itself fails) would result in a retry that hits the API.

**Agent suggests uuid.Nil as parent folder** — if `NewFolderParentID` is empty and we pass `&uuid.Nil` to `folderRepo.Create`, the folder will be created with a nil parent (top-level), which is correct. Already handled in the existing code.

**`ai_categorisation_enabled` defaults to false** — no user can accidentally incur LLM costs without explicit manual opt-in. Acceptable trade-off for the "no UI" constraint.

**Log table grows unbounded** — no purge strategy in this change. Acceptable for now given low volume, but worth revisiting when the table is exposed via API.

## Migration Plan

Two new migrations in order:
- `000019_add_ai_categorisation_enabled_to_users` — adds `ai_categorisation_enabled BOOLEAN NOT NULL DEFAULT false` to `users`
- `000020_create_ai_categorisation_logs` — creates the `ai_categorisation_logs` table

Both are additive. No data backfill needed. Rollback drops the column / drops the table.

Schema for `ai_categorisation_logs`:
```sql
CREATE TABLE ai_categorisation_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    image_id        UUID REFERENCES images(id) ON DELETE SET NULL,
    user_id         TEXT NOT NULL,
    reasoning       TEXT NOT NULL,
    folder_id       UUID,
    new_folder_name TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```
