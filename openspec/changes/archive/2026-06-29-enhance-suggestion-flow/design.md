## Context

`GetFolderSuggestion` runs an agentic loop that calls Claude with a set of tools. Currently the loop always starts with just image metadata in the user prompt, relying on the LLM to call `get_folder_list` as its first move — a guaranteed extra round trip. Two context tools (`get_folder_top_labels`, `get_folder_image_samples`) were added to `folderSuggestionToolParams` and implemented as service methods, but are not handled in the dispatch switch and cannot be used by the agent.

## Goals / Non-Goals

**Goals:**
- Eliminate the mandatory `get_folder_list` round trip by pre-fetching folders and including them in the initial prompt
- Wire `get_folder_top_labels` and `get_folder_image_samples` into the dispatch loop
- Handle invalid `folder_id` inputs leniently so the LLM can retry, with a hard cap to avoid runaway loops

**Non-Goals:**
- Changing the folder list format or the output format of the two context tools
- Modifying how submit tools work
- Any frontend or handler changes

## Decisions

### Folder list delivery: pre-fetch into user prompt vs. keep as a tool

The folder list is always needed — without it the LLM cannot make any placement decision. Delivering it upfront as part of the initial user prompt saves one guaranteed round trip with no downside. The tool approach only made sense if the list were sometimes unnecessary, which it isn't.

**Decision**: pre-fetch `listFolders` before the loop, append to user prompt alongside image metadata. Remove `get_folder_list` from tool params and dispatch.

---

### Invalid input handling: lenient result vs. hard fail

UUID parse failure and DB query errors on the context tools are most likely caused by the LLM sending a bad or hallucinated `folder_id`. These are recoverable mistakes — the LLM can see the error message and self-correct. Hard-failing immediately would be too strict.

However, uncapped retries risk burning tokens if something goes wrong structurally. A simple counter across the full loop (cap: 3) is sufficient; no need to track per-tool or per-turn.

**Decision**: return a lenient tool result string on bad input (`"invalid folder ID format"` for parse failure, `"could not retrieve folder data for the given ID"` for DB error), increment a shared `invalidInputCount`, hard-fail at 3.

No reset on success — the cap is a total across the loop, not consecutive. Keeps the logic minimal.

---

### Input parsing for context tools

Both tools take a single `folder_id` string field in their JSON input (`variant.Input`). Inline struct unpacking (`var input struct{ FolderID string \`json:"folder_id"\` }`) is sufficient — no shared type needed for two callsites.

## Risks / Trade-offs

- **Folder list in prompt increases token cost per call** — the list is always sent upfront even if the LLM would have resolved the suggestion from image metadata alone without needing it. Acceptable trade-off: one guaranteed round trip saved outweighs the marginal token cost of the list.
- **Cap of 3 is a constant in the service** — not configurable. If tuning is needed later it's a one-line change.
