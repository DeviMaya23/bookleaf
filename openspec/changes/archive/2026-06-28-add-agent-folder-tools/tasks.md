## 1. Refactor: extract extractLabels helper

- [x] 1.1 In `agent_formatter.go`, add `extractLabels(aiLabels json.RawMessage, threshold float64) ([]string, error)` that unmarshals into `[]domain.Label` and returns descriptions with `float64(score) >= threshold`
- [x] 1.2 Refactor `formatImageLabels` to call `extractLabels` instead of the inline struct, keeping its existing output shape unchanged
- [x] 1.3 Update `agent_formatter_test.go`: add unit tests for `extractLabels` (labels above threshold returned, all below returns empty, invalid JSON returns error)

## 2. Extend AgentImageRepository

- [x] 2.1 Add `ListByFolder(ctx context.Context, userID string, folderID uuid.UUID) ([]*domain.Image, error)` to the `AgentImageRepository` interface in `agent_service.go`

## 3. Implement get_folder_top_labels

- [x] 3.1 In `agent_formatter.go`, add `formatFolderTopLabels(folderID uuid.UUID, images []*domain.Image, threshold float64) (string, error)` that counts top 5 vision labels and top 5 user tags by frequency and marshals the result
- [x] 3.2 In `agent_service.go`, add `getFolderTopLabels(ctx context.Context, userID string, folderID uuid.UUID) (string, error)` that calls `ListByFolder` then `formatFolderTopLabels`
- [x] 3.3 In `agent_prompt.go`, add the `get_folder_top_labels` tool param with `folder_id` as its required input
- [x] 3.4 In `agent_formatter_test.go`, add unit tests for `formatFolderTopLabels` (images with labels and tags returns correct top counts; labels below threshold are excluded; folder with no images returns zero counts)

## 4. Implement get_folder_image_samples

- [x] 4.1 In `agent_formatter.go`, add `formatFolderImageSamples(images []*domain.Image, threshold float64) (string, error)` that takes up to 5 images and marshals each as `{image_title, image_notes, image_source_url, image_vision_labels}`
- [x] 4.2 In `agent_service.go`, add `getFolderImageSamples(ctx context.Context, userID string, folderID uuid.UUID) (string, error)` that calls `ListByFolder` with `direction="asc"`, slices to `[:5]`, then calls `formatFolderImageSamples`
- [x] 4.3 In `agent_prompt.go`, add the `get_folder_image_samples` tool param with `folder_id` as its required input
- [x] 4.4 In `agent_formatter_test.go`, add unit tests for `formatFolderImageSamples` (more than 5 images returns only 5 oldest; nil Description and SourceURL serialise as empty strings; labels below threshold are excluded)

## 5. Lint

- [x] 5.1 Run `golangci-lint run ./backend/...` and fix any issues
