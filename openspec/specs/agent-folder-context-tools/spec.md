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
