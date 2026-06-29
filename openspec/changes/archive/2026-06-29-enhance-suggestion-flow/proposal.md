## Why

The folder suggestion agent currently uses a `get_folder_list` tool, which forces one guaranteed round trip before the LLM can reason about folder placement. The two context tools (`get_folder_top_labels`, `get_folder_image_samples`) were implemented but never wired into the dispatch loop, leaving the agent unable to actually use them.

## What Changes

- Remove `get_folder_list` from `folderSuggestionToolParams` and from the dispatch switch
- Pre-fetch the folder list before the agentic loop and inject it into the initial user prompt alongside image metadata, eliminating one guaranteed round trip
- Wire `get_folder_top_labels` and `get_folder_image_samples` into the dispatch switch so the agent can use them
- Add invalid-input handling for both tools: UUID parse failure and DB query errors return a lenient tool result message so the LLM can retry
- Add a shared invalid-input counter (cap: 3) across the loop; hard-fail if exceeded

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `agent-folder-context-tools`: Add requirements for invalid input handling when the agent provides a bad `folder_id` to either context tool, and the cap-based hard-fail behavior.

## Impact

- `backend/internal/agent/agent_prompt.go`: remove `get_folder_list` tool param
- `backend/internal/agent/agent_service.go`: pre-fetch folder list, remove `get_folder_list` switch case, add `get_folder_top_labels` and `get_folder_image_samples` cases with invalid-input counter logic, remove `//nolint:unused` from the two methods
