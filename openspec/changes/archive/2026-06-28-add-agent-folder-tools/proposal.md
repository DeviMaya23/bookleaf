## Why

The folder suggestion agent currently only sees an image's own metadata when deciding where to place it. Giving the agent folder-level context — what kinds of images a folder typically contains — would produce better suggestions without requiring a full scan of every image at call time.

## What Changes

- Add `extractLabels` helper to the agent formatter, replacing the inline struct in `formatImageLabels` with `domain.Label` and centralising the label-filtering logic
- Extend `AgentImageRepository` interface with `ListByFolder`
- Add `get_folder_top_labels` tool: aggregates image count, top 5 vision labels, and top 5 user tags for a given folder
- Add `get_folder_image_samples` tool: returns metadata for the 5 oldest images in a folder (title, notes, source URL, vision labels)
- Register both tools in the agent tool param list; add dispatch cases in the service loop

## Capabilities

### New Capabilities

- `agent-folder-context-tools`: Two read-only agent tools that give the LLM aggregated and sampled context about a folder's contents, to be wired into the suggestion flow in a subsequent change

### Modified Capabilities

_(none — existing suggestion flow and its specs are unchanged)_

## Impact

- `backend/internal/agent/agent_service.go` — `AgentImageRepository` interface extended; two new service methods added
- `backend/internal/agent/agent_formatter.go` — `extractLabels` helper extracted; two new format functions added; `formatImageLabels` refactored to use `extractLabels`
- `backend/internal/agent/agent_prompt.go` — two new `ToolParam` entries added
- `backend/internal/agent/agent_formatter_test.go` — new unit tests for `extractLabels`, `formatFolderTopLabels`, `formatFolderImageSamples`
- No new dependencies; `ListByFolder` already exists on `imageRepository`
