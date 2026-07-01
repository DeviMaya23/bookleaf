# Spec: agent-folder-context-tools

## Purpose

Agent tools and supporting utilities that provide folder-level content signals for LLM reasoning. Includes label extraction helpers and two agent tools — one for aggregated frequency signals across a folder's images, and one for per-image metadata samples.

---

## Requirements

### Requirement: extractLabels filters vision labels by score threshold
The system SHALL parse a `json.RawMessage` of vision labels and return only the description strings whose score meets or exceeds the given threshold.

#### Scenario: Labels above threshold are returned
- **WHEN** `extractLabels` is called with labels where some scores are at or above threshold and some are below
- **THEN** only the descriptions of labels meeting the threshold are returned, in their original order

#### Scenario: All labels below threshold returns empty slice
- **WHEN** `extractLabels` is called and no label meets the threshold
- **THEN** an empty slice is returned with no error

#### Scenario: Invalid JSON returns error
- **WHEN** `extractLabels` is called with malformed JSON
- **THEN** an error is returned

---

### Requirement: get_folder_top_labels aggregates folder content signals
The system SHALL provide a `get_folder_top_labels` agent tool that, given a `folder_id`, fetches all images in that folder and returns aggregated frequency signals useful for LLM context.

The response SHALL include:
- `folder_id`: the input folder ID
- `folder_name`: the folder's name, or empty string if the folder is not found in the pre-fetched list
- `folder_description`: the folder's description, or empty string if unset or not found
- `image_count`: total number of images in the folder
- `top_vision_labels`: up to 5 vision label descriptions that appear most frequently across all images, counting only labels with score >= `VISION_LABEL_SCORE_THRESHOLD`
- `top_user_tags`: up to 5 tag names that appear most frequently across all images in the folder

#### Scenario: Folder with images returns aggregated signals
- **WHEN** the tool is called with a folder containing multiple images with vision labels and user tags
- **THEN** the response includes the correct image count, top 5 vision labels by frequency, and top 5 user tags by frequency

#### Scenario: Labels below threshold are excluded from top labels
- **WHEN** some images have vision labels with scores below `VISION_LABEL_SCORE_THRESHOLD`
- **THEN** those labels are not counted toward `top_vision_labels`

#### Scenario: Fewer than 5 distinct labels returns all available
- **WHEN** the folder's images collectively have fewer than 5 distinct qualifying vision labels or user tags
- **THEN** all available are returned without padding

#### Scenario: Folder with no images returns zero counts
- **WHEN** the tool is called with a folder that has no images
- **THEN** `image_count` is 0 and both top label arrays are empty

#### Scenario: Response includes folder name and description
- **WHEN** the tool is called with a valid folder ID that exists in the pre-fetched folder list
- **THEN** the response includes `folder_name` and `folder_description` matching the folder's data

---

### Requirement: get_folder_image_samples returns the 5 newest images in a folder
The system SHALL provide a `get_folder_image_samples` agent tool that, given a `folder_id`, returns metadata for up to 5 images sorted by `created_at DESC`.

Each image entry in the response SHALL include:
- `image_name`: the image's title
- `image_notes`: the image's description (may be empty string if unset)
- `image_source_url`: the image's source URL. SHALL be empty string if unset, or if the URL's host matches a search engine deny list (google.com, bing.com, duckduckgo.com, yahoo.com, yandex.com, yandex.ru, baidu.com). Search engine URLs carry no useful semantic signal for LLM reasoning.
- `image_vision_labels`: list of vision label description strings with score >= `VISION_LABEL_SCORE_THRESHOLD`

#### Scenario: Folder with more than 5 images returns only 5
- **WHEN** the tool is called with a folder containing more than 5 images
- **THEN** exactly 5 image entries are returned, corresponding to the 5 with the latest `created_at`

#### Scenario: Folder with fewer than 5 images returns all
- **WHEN** the tool is called with a folder containing fewer than 5 images
- **THEN** all images are returned

#### Scenario: Image with no description or source URL returns empty strings
- **WHEN** an image has nil `Description` or nil `SourceURL`
- **THEN** the corresponding fields in the response are empty strings, not null or omitted

#### Scenario: Source URL from a search engine is replaced with empty string
- **WHEN** an image has a `SourceURL` whose host matches an entry in the search engine deny list
- **THEN** `image_source_url` in the response is an empty string

#### Scenario: Source URL from a non-denied domain passes through unchanged
- **WHEN** an image has a `SourceURL` whose host is not in the deny list
- **THEN** `image_source_url` in the response contains the original URL

#### Scenario: Image vision labels are filtered by threshold
- **WHEN** an image has vision labels where some scores are below `VISION_LABEL_SCORE_THRESHOLD`
- **THEN** only labels meeting the threshold appear in `image_vision_labels`

---

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

---

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
