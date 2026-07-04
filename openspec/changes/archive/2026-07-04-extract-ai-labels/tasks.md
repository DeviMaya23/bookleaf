## 1. Migration

- [x] 1.1 Create `backend/migration/000021_create_image_labels.up.sql`: `CREATE TABLE image_labels` with `id`, `image_id` (FK → images ON DELETE CASCADE), `label TEXT`, `score FLOAT4`; add index on `image_id`; backfill from `ai_labels` using `jsonb_array_elements` with `jsonb_typeof` guard
- [x] 1.2 Create `backend/migration/000021_create_image_labels.down.sql`: `DROP TABLE image_labels`

## 2. Domain

- [x] 2.1 Add `ImageLabel` struct to `internal/domain/` with GORM tags (`id`, `image_id`, `label`, `score`) and `BeforeCreate` hook that generates a UUID if `ID` is nil

## 3. Repository — Write Side

- [x] 3.1 Add `UpdateLabels(ctx context.Context, id uuid.UUID, rawJSON json.RawMessage, labels []domain.ImageLabel) error` to `imageRepository`: within a single `r.db.WithContext(ctx).Transaction`, update `images.ai_labels` to `rawJSON`, delete existing `image_labels` rows for `image_id`, then bulk-insert the new label rows
- [x] 3.2 Update `UploadImageRepository` interface in `internal/usecase/`: remove `UpdateAILabels`, add `UpdateLabels` with the same signature

## 4. Repository — Agent Read Side

- [x] 4.1 Add `GetImageWithLabels(ctx, id, userID, threshold) (*domain.Image, []string, error)` to `imageRepository`: single LEFT JOIN on `images` + `image_labels WHERE score >= threshold`; return image and label descriptions ordered by score descending
- [x] 4.2 Define `FolderAggregate` struct in `internal/domain/` with `ImageCount int`, `TopVisionLabels []string`, `TopUserTags []string`; change `GetFolderTopLabels` return type from `([]string, error)` to `(*domain.FolderAggregate, error)`; update implementation to return all three signals — image count via `COUNT(DISTINCT images.id)`, vision labels via `image_labels` aggregation (score >= threshold, GROUP BY label, ORDER BY count DESC / label ASC, LIMIT topN), user tags via `image_tags → tags` aggregation (same ordering, same limit)
- [x] 4.3 Add `GetFolderImageSamples(ctx, userID, folderID, threshold, limit) ([]*domain.Image, map[uuid.UUID][]string, error)` to `imageRepository`: fetch up to `limit` images by `created_at DESC` (query 1); fetch `image_labels` for those IDs WHERE `score >= threshold` (query 2); return images and label map
- [x] 4.4 Update `AgentImageRepository` interface in `internal/agent/agent_service.go`: update `GetFolderTopLabels` return type to `(*domain.FolderAggregate, error)`; remove `ListByFolder` from the interface

## 5. Usecase — ProcessVisionLabelling

- [x] 5.1 Update `ProcessVisionLabelling` in `image_upload_usecase.go`: convert `[]domain.Label` to `[]domain.ImageLabel`, then call `imageRepo.UpdateLabels(ctx, imageID, labelsJSON, imageLabels)` instead of `imageRepo.UpdateAILabels`

## 6. Agent Formatter

- [x] 6.1 Remove `extractLabels` from `agent_formatter.go` and `agent_formatter_test.go`
- [x] 6.2 Update `formatImageLabels` signature to `(title string, labels []string) (string, error)` — remove the `extractLabels` call, use `labels` directly
- [x] 6.3 Update `formatFolderTopLabels` signature to `(folderID uuid.UUID, folder *domain.Folder, imageCount int, topVisionLabels []string, topUserTags []string) (string, error)` — drop `tagCount map[string]int` param; use `topUserTags` directly instead of calling `topN`; remove `topN` helper from `agent_formatter.go` if it has no remaining callers
- [x] 6.4 Update `formatFolderImageSamples` signature to `(images []*domain.Image, labelMap map[uuid.UUID][]string) (string, error)` — remove the `images[:5]` slice (caller handles limit) and `extractLabels` call; use `labelMap[img.ID]` for each image's labels

## 7. Agent Service

- [x] 7.1 Update `GetFolderSuggestion` in `agent_service.go`: replace `imageRepo.GetByID` + `formatImageLabels(img, threshold)` with `imageRepo.GetImageWithLabels(ctx, imageID, userID, VISION_LABEL_SCORE_THRESHOLD)` then `formatImageLabels(img.Title, labels)`
- [x] 7.2 Update `getFolderTopLabels` in `agent_service.go`: call `imageRepo.GetFolderTopLabels(ctx, userID, folderID, VISION_LABEL_SCORE_THRESHOLD, 5)` and unpack `agg.ImageCount`, `agg.TopVisionLabels`, `agg.TopUserTags`; remove `imageRepo.ListByFolder` call and tag-count loop; pass unpacked values to `formatFolderTopLabels`
- [x] 7.3 Update `getFolderImageSamples` in `agent_service.go`: call `imageRepo.GetFolderImageSamples(ctx, userID, folderID, VISION_LABEL_SCORE_THRESHOLD, 5)`; pass results to updated `formatFolderImageSamples`

## 8. Tests

- [x] 8.1 Update `image_upload_usecase_test.go`: replace `UpdateAILabels` mock with `UpdateLabels`; assert it is called with correct `rawJSON` and `[]domain.ImageLabel` args in `ProcessVisionLabelling` scenarios
- [x] 8.2 Update `agent_service_test.go`: update `AgentImageRepository` mock to remove `GetByID` and add `GetImageWithLabels`, `GetFolderTopLabels`, `GetFolderImageSamples`; rewrite affected test scenarios
- [x] 8.3 Update `agent_formatter_test.go`: remove `extractLabels` test cases; update `formatImageLabels`, `formatFolderTopLabels`, `formatFolderImageSamples` test cases to use the new signatures

## 9. Lint

- [x] 9.1 Run `golangci-lint run ./...` from `backend/` and fix any issues
