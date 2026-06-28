## Why

Vision labelling already extracts semantic labels from every uploaded image. This change closes the loop: after labelling succeeds, an AI agent uses those labels to suggest and apply a folder placement automatically — removing the manual step for users who opt in.

## What Changes

- Add `ai_categorisation_enabled` boolean to the `User` domain (default `false`; no UI toggle for now — enabled manually per user)
- Add `ANTHROPIC_MODEL` to app config so the model used by `AgentService` is configurable without a code change
- Add `CategoriseImageArgs` job args and `CategorisationWorker` to the River job queue
- Enqueue a `CategoriseImageArgs` job at the end of `ProcessVisionLabelling` when `user.ai_categorisation_enabled` is true
- Add `ai_categorisation_logs` table to persist the agent's reasoning output (image, user, reasoning, outcome fields, timestamp)
- Add `CategorisationLogRepository` and wire it into `CategorisationUsecase`
- `CategorisationUsecase.CategoriseImage` now: checks the user flag, calls the agent, persists the log, then assigns the folder
- Wire `CategorisationUsecase` and `CategorisationWorker` properly in `main.go` (currently dead code)

## Capabilities

### New Capabilities

- `ai-categorisation`: The `CategorisationUsecase.CategoriseImage` method — feature flag check, agent call, log persistence, and folder assignment
- `ai-categorisation-logs`: The `ai_categorisation_logs` table schema, migration, and `CategorisationLogRepository`

### Modified Capabilities

- `user-domain`: Add `ai_categorisation_enabled` field to `User` struct and a new `golang-migrate` migration
- `vision-api-labelling`: `ProcessVisionLabelling` enqueues a `CategoriseImageArgs` job on success when the user flag is enabled
- `async-job-queue`: Add `CategoriseImageArgs` job kind and `CategorisationWorker` registration

## Impact

- `internal/domain/user.go` — new field
- `internal/platform/config/` — new `ANTHROPIC_MODEL` config field
- `internal/agent/agent_service.go` — model sourced from struct field instead of hardcoded constant
- `internal/usecase/job_args.go` — new `CategoriseImageArgs`
- `internal/usecase/image_upload_usecase.go` — enqueue categorise job in `ProcessVisionLabelling`
- `internal/usecase/categorisation_usecase.go` — add `userRepo` and `logRepo` dependencies
- `internal/worker/categorise.go` — new file
- `internal/repository/` — new `CategorisationLogRepository` SQL implementation
- `cmd/server/main.go` — wire categorisation usecase and worker; pass model config to `AgentService`
- New DB migration files for `ai_categorisation_enabled` column and `ai_categorisation_logs` table
