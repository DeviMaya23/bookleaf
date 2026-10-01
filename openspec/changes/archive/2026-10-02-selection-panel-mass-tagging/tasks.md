## 1. BE — BulkTrash atomic refactor

- [x] 1.1 Add `BulkSoftDelete(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) (int64, error)` to the `TrashRepository` interface in `backend/internal/usecase/trash_repository.go`
- [x] 1.2 Implement `BulkSoftDelete` in `backend/internal/repository/image_repository.go`: single `UPDATE images SET deleted_at = NOW() WHERE id IN (?) AND user_id = ? AND deleted_at IS NULL`, return `RowsAffected`
- [x] 1.3 Update `BulkTrash` in `backend/internal/usecase/trash_usecase.go` to call `BulkSoftDelete(ctx, ownedIDs, userID)` instead of the per-image `SoftDelete` loop; derive `succeeded_count` from the returned row count
- [x] 1.4 Update unit tests for `BulkTrash` in `backend/internal/usecase/trash_usecase_test.go`

## 2. BE — Bulk tag endpoint

- [x] 2.1 Add `AppendImageTagsBulk(ctx context.Context, imageIDs []uuid.UUID, tagIDs []uuid.UUID) error` to the `TagRepository` interface in `backend/internal/usecase/tag_repository.go`
- [x] 2.2 Implement `AppendImageTagsBulk` in `backend/internal/repository/tag_repository.go`: build cartesian `(image_id, tag_id)` rows, bulk-insert in a single transaction with `ON CONFLICT DO NOTHING`
- [x] 2.3 Add `BulkAddTags(ctx context.Context, userID uuid.UUID, imageIDs []uuid.UUID, tagIDs []uuid.UUID) (int, error)` to the `ImageUsecase` interface and implement in `backend/internal/usecase/image_usecase.go`: validate tag ownership (all `tagIDs` must belong to `userID`; return error if any do not), filter `imageIDs` to owned via `FilterOwnedImageIDs`, call `AppendImageTagsBulk`, return owned image count
- [x] 2.4 Add `bulkAddTagsRequest` and handler `BulkAddTags` in `backend/internal/handler/image.go`; register `POST /images/bulk/tag` on the protected route group in `backend/cmd/server/main.go`
- [x] 2.5 Write unit tests for `BulkAddTags` usecase in `backend/internal/usecase/image_usecase_test.go`
- [x] 2.6 Write unit tests for `BulkAddTags` handler in `backend/internal/handler/image_test.go`
- [x] 2.7 Create `bruno/images/bulk-add-tags.bru`

## 3. FE — bulkAddTags lib function

- [x] 3.1 Add `bulkAddTags(getToken, imageIds, tagIds): Promise<BulkActionResult>` to `frontend/src/lib/tags.ts`, posting to `POST /images/bulk/tag`

## 4. FE — SelectionPanelBody tag section + unified Apply

- [x] 4.1 Add `pendingTags` state and `TagInput` section ("Add tags") below the folder section in `frontend/src/features/right-panel/components/SelectionPanelBody.tsx`; pass all user tags as suggestions
- [x] 4.2 Replace `onAddToFolder` prop with `onApply: (params: { folderIds: string[]; tagIds: string[] }) => void`; update Apply to be disabled when both arrays are empty; call `onApply({ folderIds, tagIds })` on click and clear both inputs
- [x] 4.3 Update `RightPanel.tsx` selection mode prop shape (`onApply` replaces `onAddToFolder`), forwarding `panelContent.onApply`

## 5. FE — AppLayout onApply handler

- [x] 5.1 Add `bulkAddTagsMutation` in `AppLayout.tsx` using `bulkAddTags`
- [x] 5.2 Replace `handleAddSelectionToFolder` with `handleApplySelection({ folderIds, tagIds })`: fire folder requests (one per folder) and tag request concurrently; invalidate images on success; show success toast and exit select mode only when all succeed; show per-operation error toast on failure without exiting select mode
- [x] 5.3 Update `panelContent` object in `AppLayout.tsx` to pass `onApply` instead of `onAddToFolder`

## 6. FE — Tests

- [x] 6.1 Update `RightPanel.test.tsx`: replace `onAddToFolder` usage with `onApply`; add scenario for Apply with tags only; add scenario for Apply with both folders and tags
- [x] 6.2 Update `AppLayout.test.tsx` to reflect the new `onApply` prop shape

## 7. FE — Build and lint

- [x] 7.1 Run `npm run build` and fix any type errors
- [x] 7.2 Run `npm run lint` and fix any lint issues

## 8. BE — Lint

- [x] 8.1 Run `golangci-lint run ./...` from `backend/` and fix any issues
