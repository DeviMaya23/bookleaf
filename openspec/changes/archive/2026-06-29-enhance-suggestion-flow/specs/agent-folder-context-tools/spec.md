## ADDED Requirements

### Requirement: get_folder_top_labels returns lenient error on invalid folder_id input
The system SHALL return a descriptive error string as the tool result when `get_folder_top_labels` receives an unparseable `folder_id` or encounters a DB query error, allowing the LLM to retry rather than crashing the loop.

- On UUID parse failure, the tool result SHALL be `"invalid folder ID format"`.
- On DB query error, the tool result SHALL be `"could not retrieve folder data for the given ID"`.
- In both cases the `invalidInputCount` SHALL be incremented.

#### Scenario: Unparseable folder_id returns lenient error
- **WHEN** `get_folder_top_labels` is dispatched with a `folder_id` that cannot be parsed as a UUID
- **THEN** the tool result returned to the LLM is `"invalid folder ID format"` and `invalidInputCount` is incremented

#### Scenario: DB error returns lenient error
- **WHEN** `get_folder_top_labels` is dispatched with a valid UUID but the repository query fails
- **THEN** the tool result returned to the LLM is `"could not retrieve folder data for the given ID"` and `invalidInputCount` is incremented

---

### Requirement: get_folder_image_samples returns lenient error on invalid folder_id input
The system SHALL return a descriptive error string as the tool result when `get_folder_image_samples` receives an unparseable `folder_id` or encounters a DB query error, allowing the LLM to retry rather than crashing the loop.

- On UUID parse failure, the tool result SHALL be `"invalid folder ID format"`.
- On DB query error, the tool result SHALL be `"could not retrieve folder data for the given ID"`.
- In both cases the `invalidInputCount` SHALL be incremented.

#### Scenario: Unparseable folder_id returns lenient error
- **WHEN** `get_folder_image_samples` is dispatched with a `folder_id` that cannot be parsed as a UUID
- **THEN** the tool result returned to the LLM is `"invalid folder ID format"` and `invalidInputCount` is incremented

#### Scenario: DB error returns lenient error
- **WHEN** `get_folder_image_samples` is dispatched with a valid UUID but the repository query fails
- **THEN** the tool result returned to the LLM is `"could not retrieve folder data for the given ID"` and `invalidInputCount` is incremented

---

### Requirement: Agent loop hard-fails when invalid input cap is reached
The system SHALL maintain a shared `invalidInputCount` across the full agentic loop. If the count reaches 3, the loop SHALL return an error immediately without making another LLM call.

#### Scenario: Third invalid input hard-fails the loop
- **WHEN** `invalidInputCount` reaches 3 after incrementing on a bad tool input
- **THEN** `GetFolderSuggestion` returns an error and the loop terminates

#### Scenario: Fewer than 3 invalid inputs allows the loop to continue
- **WHEN** `invalidInputCount` is below 3
- **THEN** the loop continues and the lenient tool result is returned to the LLM for retry

### Requirement: GetFolderSuggestion emits a structured log on completion
The system SHALL emit a single structured log entry via `zap` (using `LoggerFromContext`) when `GetFolderSuggestion` exits, whether by success, hard-fail, or early error.

The log entry SHALL include:
- `user_id`: the requesting user's ID
- `image_id`: the image UUID being categorised
- `tool_call_count`: total number of tool calls dispatched during the loop (int)
- `tools_used`: ordered list of tool names dispatched, including submit tools and invalid-input calls ([]string)

The log SHALL fire via `defer` so it is emitted even when the function returns early (e.g., folder pre-fetch failure, image fetch failure). In those cases `tool_call_count` SHALL be 0 and `tools_used` SHALL be empty.

#### Scenario: Successful suggestion logs all dispatched tools
- **WHEN** `GetFolderSuggestion` completes successfully after dispatching one or more tools
- **THEN** the log entry includes the correct `tool_call_count` and `tools_used` reflecting every tool called in order, including the submit tool

#### Scenario: Early error before loop logs zero tool calls
- **WHEN** `GetFolderSuggestion` returns an error before entering the loop (e.g., folder pre-fetch or image fetch fails)
- **THEN** the log entry is still emitted with `tool_call_count: 0` and `tools_used: []`

#### Scenario: Invalid input calls appear in tools_used
- **WHEN** a context tool is dispatched with a bad `folder_id` before a successful call
- **THEN** that tool name appears in `tools_used` for each invocation, including the failed ones

## REMOVED Requirements

### Requirement: get_folder_list tool is available to the agent
**Reason**: Folder list is now pre-fetched before the loop and injected into the initial user prompt, eliminating the need for a dedicated tool and removing one guaranteed round trip.
**Migration**: No external migration needed. The folder list is delivered as part of the initial user prompt in `GetFolderSuggestion`.
