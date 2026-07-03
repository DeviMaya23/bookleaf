## REMOVED Requirements

### Requirement: extractLabels filters vision labels by score threshold

**Reason**: Replaced by SQL-side filtering in the new `AgentImageRepository` methods. Label filtering by score threshold now happens in the database query (`WHERE score >= $threshold`) rather than in Go after unmarshaling JSONB. `extractLabels` is the only function in `agent_formatter.go` that reads `json.RawMessage`; removing it is the only deletion from that file.

**Migration**: All call sites of `extractLabels` within `agent_formatter.go` are replaced by pre-fetched `[]string` label data passed in via updated function signatures. `agent_formatter_test.go` scenarios that test `extractLabels` directly are removed.

---

## ADDED Requirements

### Requirement: AgentImageRepository.GetImageWithLabels method

The system SHALL add `GetImageWithLabels(ctx context.Context, id uuid.UUID, userID string, threshold float64) (*domain.Image, []string, error)` to the `AgentImageRepository` interface and implement it on `imageRepository`.

The method SHALL execute a single LEFT JOIN query against `images` and `image_labels`, returning:
- The `*domain.Image` (title, description, source_url, and other image fields)
- A `[]string` of label descriptions where `image_labels.score >= threshold`, ordered by score descending

`GetByID` SHALL be removed from `AgentImageRepository`; `GetImageWithLabels` replaces it. The existing `GetByID` on `imageRepository` is unaffected.

#### Scenario: Image with labels above threshold returns filtered label strings

- **WHEN** `GetImageWithLabels` is called for an image with labels at varying scores and a threshold of `VISION_LABEL_SCORE_THRESHOLD`
- **THEN** only label descriptions meeting the threshold are returned, ordered by score descending
- **AND** the returned `*domain.Image` contains the image's metadata

#### Scenario: Image with no qualifying labels returns empty slice

- **WHEN** `GetImageWithLabels` is called and no `image_labels` rows meet the threshold
- **THEN** `[]string{}` is returned for labels and the image metadata is still returned without error

#### Scenario: Image not found returns error

- **WHEN** `GetImageWithLabels` is called with an ID that does not exist for the given user
- **THEN** a non-nil error is returned

---

### Requirement: AgentImageRepository.GetFolderTopLabels method

The system SHALL add `GetFolderTopLabels(ctx context.Context, userID string, folderID uuid.UUID, threshold float64, topN int) (*FolderAggregate, error)` to the `AgentImageRepository` interface and implement it on `imageRepository`.

`FolderAggregate` is a domain type (in `internal/domain/`) used as a query read-model:
```go
type FolderAggregate struct {
    ImageCount      int
    TopVisionLabels []string
    TopUserTags     []string
}
```

The method SHALL execute a single database operation that returns all three signals:
- `ImageCount`: count of non-deleted images in the folder for the given user
- `TopVisionLabels`: top `topN` label descriptions from `image_labels` where `score >= threshold`, ordered by frequency descending then label ascending for tie-breaking
- `TopUserTags`: top `topN` tag names from `image_tags → tags`, ordered by frequency descending then name ascending for tie-breaking

`GetByID` and `ListByFolder` are both removed from `AgentImageRepository`. `GetFolderTopLabels` replaces `ListByFolder` for the `get_folder_top_labels` tool — no separate call to `ListByFolder` is needed.

#### Scenario: Returns all three signals in one call

- **WHEN** `GetFolderTopLabels` is called for a folder with images that have vision labels and user tags
- **THEN** a `*FolderAggregate` is returned with `ImageCount`, `TopVisionLabels`, and `TopUserTags` all populated in a single DB round trip

#### Scenario: Returns top N labels by frequency across folder images

- **WHEN** `GetFolderTopLabels` is called with `topN = 5` and the folder has images with multiple qualifying labels
- **THEN** `TopVisionLabels` contains up to 5 label strings, ordered by frequency descending

#### Scenario: Labels below threshold are excluded

- **WHEN** some `image_labels` rows have `score` below `threshold`
- **THEN** those labels are not counted and do not appear in `TopVisionLabels`

#### Scenario: Folder with no images returns zero count and empty slices

- **WHEN** `GetFolderTopLabels` is called for a folder that contains no images
- **THEN** `ImageCount` is 0 and both `TopVisionLabels` and `TopUserTags` are empty slices without error

#### Scenario: Fewer than topN distinct labels returns all available

- **WHEN** the folder's images have fewer qualifying distinct labels than `topN`
- **THEN** all available label strings are returned without padding

---

### Requirement: AgentImageRepository.GetFolderImageSamples method

The system SHALL add `GetFolderImageSamples(ctx context.Context, userID string, folderID uuid.UUID, threshold float64, limit int) ([]*domain.Image, map[uuid.UUID][]string, error)` to the `AgentImageRepository` interface and implement it on `imageRepository`.

The method SHALL:
1. Fetch up to `limit` non-deleted images in the folder for the given user, ordered by `created_at DESC` (one query)
2. If images are returned, fetch all `image_labels` rows for those image IDs where `score >= threshold` in a single `WHERE image_id IN (...)` query
3. Return the images and a map of `imageID → []string` of filtered label descriptions

#### Scenario: Returns up to limit images with their filtered labels

- **WHEN** `GetFolderImageSamples` is called with `limit = 5` and the folder has more than 5 images
- **THEN** exactly 5 images are returned (the 5 most recently created), each mapped to their qualifying labels

#### Scenario: Folder with fewer than limit images returns all

- **WHEN** the folder has fewer than `limit` images
- **THEN** all images are returned

#### Scenario: Labels map contains empty slice for images with no qualifying labels

- **WHEN** an image in the result has no `image_labels` rows meeting the threshold
- **THEN** that image's entry in the map is an empty `[]string{}`

#### Scenario: Empty folder returns empty slices

- **WHEN** `GetFolderImageSamples` is called for a folder with no images
- **THEN** an empty image slice and empty map are returned without error

---

## MODIFIED Requirements

### Requirement: extractLabels filters vision labels by score threshold

*(See REMOVED section above — this requirement is fully removed.)*

---

### Requirement: get_folder_top_labels aggregates folder content signals

The system SHALL provide a `get_folder_top_labels` agent tool that, given a `folder_id`, returns aggregated frequency signals useful for LLM context.

All three signals (`image_count`, `top_vision_labels`, `top_user_tags`) SHALL be sourced from a single call to `AgentImageRepository.GetFolderTopLabels`, which returns a `*FolderAggregate`. `ListByFolder` is no longer called from this tool handler.

The response shape and all output requirements are unchanged:
- `folder_id`, `folder_name`, `folder_description`, `image_count`, `top_vision_labels`, `top_user_tags`

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

Image and label data SHALL be sourced by calling `AgentImageRepository.GetFolderImageSamples` and passed to `formatFolderImageSamples`.

The response shape and all output requirements are unchanged:
- `image_name`, `image_notes`, `image_source_url`, `image_vision_labels` per image

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

### Requirement: formatImageLabels formats a single image for LLM context

The system SHALL update `formatImageLabels` in `agent_formatter.go` to accept pre-fetched data instead of a raw `*domain.Image`:

**New signature**: `formatImageLabels(title string, labels []string) (string, error)`

The function SHALL produce a JSON object with `image_name` and `vision_results` fields, identical in shape to before.

#### Scenario: Returns JSON with title and filtered label strings

- **WHEN** `formatImageLabels` is called with a title and a pre-fetched `[]string` of label descriptions
- **THEN** the returned JSON contains `image_name` equal to the title and `vision_results` equal to the label slice

---

### Requirement: formatFolderTopLabels formats folder signals for LLM context

The system SHALL update `formatFolderTopLabels` in `agent_formatter.go` to accept pre-fetched data:

**New signature**: `formatFolderTopLabels(folderID uuid.UUID, folder *domain.Folder, imageCount int, topVisionLabels []string, topUserTags []string) (string, error)`

The function SHALL produce the same JSON shape as before. Both `top_vision_labels` and `top_user_tags` are passed in as pre-ranked `[]string` slices — the function no longer calls `topN` for either. The `topN` helper MAY be removed from `agent_formatter.go` if it has no remaining callers.

#### Scenario: Returns JSON with all folder signal fields

- **WHEN** `formatFolderTopLabels` is called with pre-fetched vision labels and pre-ranked user tags
- **THEN** the returned JSON contains `folder_id`, `folder_name`, `folder_description`, `image_count`, `top_vision_labels`, and `top_user_tags`

---

### Requirement: formatFolderImageSamples formats image samples for LLM context

The system SHALL update `formatFolderImageSamples` in `agent_formatter.go` to accept pre-fetched label data:

**New signature**: `formatFolderImageSamples(images []*domain.Image, labelMap map[uuid.UUID][]string) (string, error)`

The function SHALL no longer slice `images` to 5 — the caller (`GetFolderImageSamples`) is responsible for the limit. All other behaviour (source URL deny list, nil field handling) is unchanged.

#### Scenario: Returns JSON samples using label map instead of extractLabels

- **WHEN** `formatFolderImageSamples` is called with images and a pre-fetched label map
- **THEN** each entry's `image_vision_labels` contains the labels from the map for that image ID

#### Scenario: Image absent from label map returns empty vision labels

- **WHEN** an image ID is not present in `labelMap`
- **THEN** `image_vision_labels` for that entry is an empty array
