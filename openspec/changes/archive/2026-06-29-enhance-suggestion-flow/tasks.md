## 1. Remove get_folder_list tool

- [x] 1.1 Remove the `get_folder_list` entry from `folderSuggestionToolParams` in `agent_prompt.go`

## 2. Pre-fetch folder list into initial prompt

- [x] 2.1 In `GetFolderSuggestion`, call `u.listFolders(ctx, userID)` before the loop and hard-fail if it errors
- [x] 2.2 Update the initial user prompt to include the folder list: `"Image metadata: {imageMetadata}\nFolder list: {folderList}"`

## 3. Wire context tools into dispatch loop

- [x] 3.1 Add `invalidInputCount` counter (int) before the loop in `GetFolderSuggestion`
- [x] 3.2 Add `get_folder_top_labels` case: parse `folder_id` from `variant.Input`, call `getFolderTopLabels`; on UUID parse failure return `"invalid folder ID format"` and increment counter; on DB error return `"could not retrieve folder data for the given ID"` and increment counter; check cap after incrementing
- [x] 3.3 Add `get_folder_image_samples` case: same pattern as 3.2 using `getFolderImageSamples`
- [x] 3.4 After each counter increment, if `invalidInputCount >= 3` return a hard error from the loop
- [x] 3.5 Remove the `get_folder_list` switch case
- [x] 3.6 Remove `//nolint:unused` comments from `getFolderTopLabels` and `getFolderImageSamples`

## 4. Log tool usage on completion

- [x] 4.1 Declare `toolsUsed []string` before the loop in `GetFolderSuggestion`
- [x] 4.2 Append the tool name to `toolsUsed` on each tool dispatch (all cases in the switch, including submit tools before returning)
- [x] 4.3 Add a `defer` at the top of `GetFolderSuggestion` (after `toolsUsed` is declared) that logs `user_id`, `image_id`, `tool_call_count` (`len(toolsUsed)`), and `tools_used` via `observability.LoggerFromContext(ctx, u.tel.Logger).Info`

## 5. Lint

- [x] 5.1 Run `golangci-lint run ./backend/...` and fix any issues
