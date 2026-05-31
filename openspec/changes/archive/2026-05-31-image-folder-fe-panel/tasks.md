## 1. BE — Update Response Structs

- [x] 1.1 Replace `FolderID *uuid.UUID` and `Position *string` fields in `imageResponse` and `imageDetailResponse` structs with `FolderIDs []uuid.UUID` in `handler/image.go`
- [x] 1.2 Update `toImageResponse` to populate `FolderIDs` by iterating `item.Image.ImageFolders` (all entries)
- [x] 1.3 Remove `firstFolderID` and `firstFolderPosition` helpers from `handler/image.go`
- [x] 1.4 Update `GetImage` handler to map the new `FolderIDs` field into `imageDetailResponse`

## 2. BE — Unit Tests

- [x] 2.1 Update handler unit tests in `handler/image_test.go` to assert `folder_ids` array instead of `folder_id` (success scenario: multi-folder image returns all IDs; failure scenario: unfiled image returns empty array)

## 3. FE — Update Image Type

- [x] 3.1 Replace `folder_id: string | null` with `folder_ids: string[]` on the `Image` interface in `lib/images.ts`
- [x] 3.2 Update all test fixtures that reference `folder_id` to use `folder_ids: []` (`ImageGrid.test.tsx`, `RightPanel.test.tsx`, `images.test.ts`, `dragHandlers.test.ts`)

## 4. FE — FolderInput Component

- [x] 4.1 Create `frontend/src/components/FolderInput.tsx` — combobox multi-select with pill chips, filtered dropdown from suggestions, no inline creation (per `fe-image-folder-panel` spec)
- [x] 4.2 Write unit tests for `FolderInput` (success: selecting a suggestion calls onChange with folder appended; failure: blurring without selecting does not call onChange)

## 5. FE — RightPanel Integration

- [x] 5.1 Remove `folderName` derived value and the static Folder row from the details grid in `RightPanel.tsx`
- [x] 5.2 Initialise local `folders` state from `image.folder_ids` (resolved via `['folders']` cache), reset on `image.id` change
- [x] 5.3 Add Folders section between Source URL and Tags, rendering `FolderInput` pre-populated with resolved folder objects
- [x] 5.4 On `FolderInput` `onChange`: call `updateImage({ folder_ids: [...] })`, show success/error toast, invalidate `['images']` on success
- [x] 5.5 Update `RightPanel` unit tests — success: adding a folder calls PATCH with updated folder_ids; failure: PATCH failure shows error toast

## 6. FE — Bruno File

- [x] 6.1 Update the existing `GET /images/:id` Bruno request file to document the new `folder_ids` array response field
