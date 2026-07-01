## Purpose

Defines `CategorisationUsecase` — the usecase responsible for using an AI agent to assign an image to an existing or new folder, and persisting the agent's reasoning to the log table.

---

## Requirements

### Requirement: CategorisationUsecase Dependencies

The system SHALL define `CategorisationUsecase` in `internal/usecase/categorisation_usecase.go` with the following dependencies:

- `agentService *agent.AgentService` — calls the LLM agent for a folder suggestion
- `imageRepo categorisationImageRepository` — fetches and updates images
- `folderRepo categorisationFolderRepository` — creates new folders
- `logRepo categorisationLogRepository` — persists categorisation log entries
- `tel *observability.Telemetry` — tracing

`NewCategorisationUsecase` SHALL accept these in the order above and return `*CategorisationUsecase`.

#### Scenario: Constructor wires all dependencies

- **WHEN** the Go package is compiled
- **THEN** `NewCategorisationUsecase` accepts `agentService`, `imageRepo`, `folderRepo`, `logRepo`, and `tel` and returns a `*CategorisationUsecase` without compilation errors

---

### Requirement: CategoriseImage Method

`CategorisationUsecase.CategoriseImage(ctx context.Context, userID string, imageID uuid.UUID) error` SHALL execute the following steps in order:

1. Call `logRepo.GetByImageID(ctx, imageID)` to check for an existing log entry for this image
2. Call `imageRepo.GetByID(ctx, imageID, userID)` to fetch the image; if the image already has a `FolderID` set, return nil without calling the agent
3. If no existing log entry: call `agentService.GetFolderSuggestion(ctx, userID, imageID)` and persist a `CategorisationLog` entry via `logRepo.Create` with `imageID`, `userID`, `Suggestion.Reasoning`, `Suggestion.FolderID` (if set), and `Suggestion.NewFolderName` (if set)
4. If an existing log entry was found: use it as the suggestion source, skipping the LLM call entirely
5. If `Suggestion.FolderID` is non-empty: parse the UUID and call `imageRepo.SetImageFolder`
6. If `Suggestion.NewFolderName` is non-empty: optionally parse `Suggestion.NewFolderParentID`, call `folderRepo.Create`, then call `imageRepo.SetImageFolder` with the new folder's ID
7. If neither field is set: return an error (`"agent returned suggestion with no folder id or new folder name"`)

The guard at step 2 prevents the agent from overriding an explicit folder placement made by the user at upload time. The idempotency check at step 1 ensures that on River retries caused by downstream failures, the Anthropic API is not called again. Failures at any step SHALL return a non-nil error.

#### Scenario: Image already has a folder — agent not called, nil returned

- **WHEN** `CategoriseImage` is called for an image that already has a `FolderID` set
- **THEN** `agentService.GetFolderSuggestion` is NOT called
- **AND** no log entry is created
- **AND** nil is returned

#### Scenario: Agent suggests existing folder — folder assigned and log persisted

- **WHEN** no prior log entry exists and `GetFolderSuggestion` returns a `Suggestion` with a valid `FolderID`
- **THEN** a `CategorisationLog` row is inserted with the reasoning and `folder_id`
- **AND** `imageRepo.SetImageFolder` is called with the parsed folder UUID
- **AND** nil is returned

#### Scenario: Agent suggests new folder — folder created, assigned, and log persisted

- **WHEN** no prior log entry exists and `GetFolderSuggestion` returns a `Suggestion` with a non-empty `NewFolderName` and no `FolderID`
- **THEN** a `CategorisationLog` row is inserted with the reasoning and `new_folder_name`
- **AND** `folderRepo.Create` is called with the new folder name and optional parent ID
- **AND** `imageRepo.SetImageFolder` is called with the newly created folder's ID
- **AND** nil is returned

#### Scenario: Retry reuses existing log entry — LLM not called again

- **WHEN** a `CategorisationLog` entry already exists for the image (from a prior attempt)
- **THEN** `agentService.GetFolderSuggestion` is NOT called
- **AND** the existing log entry's suggestion is used to proceed with folder assignment

#### Scenario: Agent returns empty suggestion — error returned

- **WHEN** `GetFolderSuggestion` returns a `Suggestion` with both `FolderID` and `NewFolderName` empty
- **THEN** `CategoriseImage` returns a non-nil error

#### Scenario: Log persisted before folder assignment fails

- **WHEN** `GetFolderSuggestion` returns a valid suggestion but `imageRepo.SetImageFolder` returns an error
- **THEN** the `CategorisationLog` entry has already been inserted
- **AND** a retry will use the existing log entry and skip the LLM call

---

### Requirement: Configurable Anthropic Model

`AgentService` SHALL store the Anthropic model as a `string` field set at construction time via `NewAgentService`. The model SHALL be sourced from `config.AnthropicModel` (env var `ANTHROPIC_MODEL`, defaulting to `claude-haiku-4-5-20251001`). The hardcoded `anthropic.ModelClaudeHaiku4_5` constant SHALL be removed from `GetFolderSuggestion`.

#### Scenario: Model defaults to Haiku when env var is unset

- **WHEN** `ANTHROPIC_MODEL` is not set in the environment
- **THEN** `AgentService` uses `claude-haiku-4-5-20251001` as the model for all API calls

#### Scenario: Model is overridden by env var

- **WHEN** `ANTHROPIC_MODEL=claude-sonnet-4-5` is set in the environment
- **THEN** `AgentService` uses `claude-sonnet-4-5` for all API calls
