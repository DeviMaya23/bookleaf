## 1. Domain and Config

- [x] 1.1 Add `AICategorisationEnabled bool` field to `User` struct in `internal/domain/user.go` with `gorm:"column:ai_categorisation_enabled;default:false"`
- [x] 1.2 Add migration `000019_add_ai_categorisation_enabled_to_users` (up: `ALTER TABLE users ADD COLUMN ai_categorisation_enabled BOOLEAN NOT NULL DEFAULT false`, down: drop column)
- [x] 1.3 Add `AnthropicModel string` to `Config` in `internal/platform/config/config.go`, reading from `ANTHROPIC_MODEL` env var with default `claude-haiku-4-5-20251001`

## 2. Agent Service

- [x] 2.1 Add `model string` field to `AgentService` struct; update `NewAgentService` to accept it as a parameter
- [x] 2.2 Replace the hardcoded `anthropic.ModelClaudeHaiku4_5` in `GetFolderSuggestion` with `u.model`

## 3. Categorisation Log Table and Repository

- [x] 3.1 Define `CategorisationLog` struct in `internal/domain/` with fields: `ID uuid.UUID`, `ImageID *uuid.UUID`, `UserID string`, `Reasoning string`, `FolderID *uuid.UUID`, `NewFolderName *string`, `CreatedAt time.Time`
- [x] 3.2 Add migration `000020_create_ai_categorisation_logs` (up: create table per design schema, down: drop table)
- [x] 3.3 Define `categorisationLogRepository` interface in `internal/usecase/categorisation_usecase.go` with `Create(ctx context.Context, log *domain.CategorisationLog) error`
- [x] 3.4 Implement `CategorisationLogRepository` in `internal/repository/` with a GORM `Create` insert into `ai_categorisation_logs`
- [x] 3.5 Add `GetByImageID(ctx context.Context, imageID uuid.UUID) (*domain.CategorisationLog, error)` to the `categorisationLogRepository` interface; return `nil, nil` if no entry exists
- [x] 3.6 Implement `GetByImageID` in the SQL repository — query `ai_categorisation_logs` by `image_id`, order by `created_at DESC`, limit 1

## 4. CategorisationUsecase

- [x] 4.1 Add `logRepo categorisationLogRepository` to `CategorisationUsecase` struct and `NewCategorisationUsecase` constructor
- [x] 4.2 In `CategoriseImage`, persist a `CategorisationLog` entry via `logRepo.Create` immediately after `GetFolderSuggestion` returns, before any folder assignment
- [x] 4.3 Add idempotency check at the start of `CategoriseImage`: call `logRepo.GetByImageID`; if an entry exists, skip the agent call and use the existing log entry's suggestion for folder assignment

## 5. River Job

- [x] 5.1 Add `CategoriseImageArgs` to `internal/usecase/job_args.go` with `Kind() = "categorise_image"` and `MaxAttempts() = 3`
- [x] 5.2 Create `internal/worker/categorise.go` with `CategorisationWorker` implementing `river.Worker[CategoriseImageArgs]`; override `NextRetry` for a fixed 30-second delay; call `categorisationUsecase.CategoriseImage` on work

## 6. Enqueue After Vision

- [x] 6.1 In `ProcessVisionLabelling` (`internal/usecase/image_upload_usecase.go`), after `UpdateAILabels` succeeds, check `user.AICategorisationEnabled` and enqueue a `CategoriseImageArgs` job via `u.enqueuer` if true

## 7. Wire in main.go

- [x] 7.1 Pass `cfg.AnthropicModel` to `agent.NewAgentService` in `initApp`
- [x] 7.2 Create `CategorisationLogRepository` and pass it to `NewCategorisationUsecase` (replace the `_ =` dead-code assignment)
- [x] 7.3 Register `CategorisationWorker` with `river.AddWorker` and pass the wired `CategorisationUsecase` as its dependency

## 8. Already-in-folder Guard

- [x] 8.1 In `CategoriseImage`, after the idempotency check, call `imageRepo.GetByID` and return nil if the image already has a `FolderID` set
- [x] 8.2 Add a unit test scenario: image already has a folder → agent not called, no folder assignment, nil returned

## 9. Unit Tests

- [x] 8.1 In `internal/usecase/categorisation_usecase_test.go`, write unit tests for `CategoriseImage`:
  - Idempotency: existing log found → `GetFolderSuggestion` not called, `SetImageFolder` called with log's `FolderID`
  - FolderID branch: no existing log, agent returns `FolderID` → log created, `SetImageFolder` called with parsed UUID
  - NewFolderName branch: no existing log, agent returns `NewFolderName` → log created, folder created, `SetImageFolder` called with new folder ID
  - Empty suggestion → error message asserted with `require.ErrorContains`
  - Use a fake for `logRepo` (reads pre-state to decide writes); use value-return spies for agent, imageRepo, folderRepo
- [x] 8.2 In the existing `ProcessVisionLabelling` test file, add two scenarios for the categorise enqueue:
  - `ai_categorisation_enabled = true` → enqueuer called with `CategoriseImageArgs` containing the correct image ID and user ID
  - `ai_categorisation_enabled = false` → enqueuer not called with `CategoriseImageArgs`
- [x] 8.3 In `internal/agent/agent_formatter_test.go`, write unit tests for the pure formatter functions:
  - `formatImageLabels`: correct JSON output from image with AI labels
  - `getFolderPath`: flat folder returns name only; nested folder returns full `parent > child` path

## 10. Lint

- [x] 10.1 Run `golangci-lint run ./...` from `backend/` and fix any issues
