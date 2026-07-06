## 1. Backend — Remove Accept-Suggestion Endpoint

- [x] 1.1 Remove `AcceptSuggestion` from the `UploadUsecase` interface in `internal/handler/image_upload.go`
- [x] 1.2 Remove `acceptSuggestionRequest` type and `AcceptSuggestion` handler method from `internal/handler/image_upload.go`
- [x] 1.3 Remove `AcceptSuggestion` method from `imageUploadUsecase` in `internal/usecase/image_upload_usecase.go`
- [x] 1.4 Remove `POST /images/:id/accept-suggestion` route from `cmd/server/main.go`
- [x] 1.5 Delete `bruno/images/accept-suggestion.bru`

## 2. Backend — Remove suggested_folder_name from Image Response

- [x] 2.1 Remove `SuggestedFolderName *string` field from `ImageDetail` struct in `internal/usecase/image_usecase.go`
- [x] 2.2 Delete the `suggestedFolderName()` helper function from `internal/usecase/image_usecase.go`
- [x] 2.3 Remove the `SuggestedFolderName: suggestedFolderName(image.AILabels)` mapping line from the `GetImage` usecase
- [x] 2.4 Remove `SuggestedFolderName *string` from the image handler response struct in `internal/handler/image.go` and its mapping line

## 3. Backend — Tests

- [x] 3.1 Remove any handler or usecase tests covering `AcceptSuggestion`
- [x] 3.2 Run `go test ./...` and fix any compilation errors or failing tests
- [x] 3.3 Run `golangci-lint run` and fix any issues

## 4. Frontend — Remove Folder Suggestion Flow

- [x] 4.1 Delete `frontend/src/app-shell/useVisionSuggestion.ts`
- [x] 4.2 In `AppLayout.tsx`: remove the `useVisionSuggestion` import, the `checkVision` const, the `checkVision(imageDetail.id)` call in the drag-drop handler, and the `onUploadSuccess={checkVision}` prop on `<UploadModal>`
- [x] 4.3 Remove the `onUploadSuccess` prop (type definition + usage) from `UploadModal.tsx`
- [x] 4.4 In `lib/images.ts`: remove `acceptSuggestion()` function and remove `suggested_folder_name` from the `GetImageResponse` type

## 5. Frontend — Gate Categorisation Toggle on vision_enabled

- [x] 5.1 In `AdvancedSection.tsx`, disable the AI auto-categorisation `Switch` when `vision_enabled` is `false` (in addition to when the mutation is pending)
- [x] 5.2 Add a tooltip or inline hint on the categorisation row indicating the toggle requires Smart Features to be enabled, visible when `vision_enabled` is `false`
- [x] 5.3 Update the section heading label from "AI folder suggestions" to "Smart Features" and update the toggle description text and tooltip to reflect the new framing

## 6. Frontend — Tests

- [x] 6.1 Remove the `useVisionSuggestion` mock from `AppLayout.test.tsx`
- [x] 6.2 In `AdvancedSection.test.tsx`: update any copy-based assertions to use `data-testid`/role; add a test asserting the categorisation switch is disabled when `vision_enabled` is `false`
- [x] 6.3 Remove `suggested_folder_name` from mock return values in `BatchUploadModal.test.tsx`, `upload.test.ts`, and `UploadModal.test.tsx`
- [x] 6.4 Run `npm run build` and `npm run lint` and fix any issues
