## Why

The original "folder suggestion" flow (Vision API labels → toast with Accept/Ignore → `POST /images/:id/accept-suggestion`) has been superseded by smart search and AI auto-categorisation, making it dead weight. The settings UI also has no visual gate enforcing that `ai_categorisation_enabled` requires `vision_enabled`, which is inconsistent with the actual backend dependency.

## What Changes

- **REMOVE** the folder suggestion toast flow: `useVisionSuggestion` hook, `checkVision` wiring in `AppLayout`, `onUploadSuccess` prop on `UploadModal`, and `acceptSuggestion()` API call
- **REMOVE** `POST /images/:id/accept-suggestion` endpoint, handler, and usecase method
- **REMOVE** `suggested_folder_name` from `GET /images/:id` response and from `ImageDetail` usecase struct; remove `suggestedFolderName()` helper
- **REMOVE** the `accept-suggestion.bru` Bruno file
- **UPDATE** `AdvancedSection` settings UI to visually disable the AI auto-categorisation toggle when `vision_enabled` is off, with a hint explaining the dependency
- **UPDATE** copy in `AdvancedSection` to reflect the new framing (`vision_enabled` = "Smart Features", not "AI folder suggestions")

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `vision-api-labelling`: remove `suggested_folder_name` from `GetImage` response; the BE no longer derives a suggested folder name from vision labels
- `image-endpoints`: remove `suggested_folder_name` field from the `GET /images/:id` response shape
- `fe-vision-toggle`: gate the AI auto-categorisation switch on `vision_enabled`; update section label and toggle copy to reflect "Smart Features" framing

### Removed Capabilities

- `accept-suggestion`: entire spec is obsolete — endpoint, handler, usecase method, and frontend wiring are all deleted

## Impact

**Backend**
- `internal/handler/image_upload.go`: remove `AcceptSuggestion` handler, `acceptSuggestionRequest` type, `AcceptSuggestion` from `UploadUsecase` interface
- `internal/usecase/image_upload_usecase.go`: remove `AcceptSuggestion` method
- `internal/usecase/image_usecase.go`: remove `suggestedFolderName()` helper and `SuggestedFolderName *string` from `ImageDetail`
- `internal/handler/image.go`: remove `SuggestedFolderName` from image response struct and mapping
- `cmd/server/main.go`: remove `POST /images/:id/accept-suggestion` route
- `bruno/images/accept-suggestion.bru`: delete

**Frontend**
- `app-shell/useVisionSuggestion.ts`: delete
- `app-shell/AppLayout.tsx`: remove import, `checkVision` const, two call sites
- `features/upload/components/UploadModal.tsx`: remove `onUploadSuccess` prop
- `lib/images.ts`: remove `acceptSuggestion()`, remove `suggested_folder_name` from `GetImageResponse` type
- `features/settings/components/AdvancedSection.tsx`: gate categorisation switch on `vision_enabled`, update copy

**Tests affected**
- `AppLayout.test.tsx`: remove `useVisionSuggestion` mock
- `AdvancedSection.test.tsx`: update copy assertions; add test for categorisation switch disabled state
- `BatchUploadModal.test.tsx`, `upload.test.ts`, `UploadModal.test.tsx`: remove `suggested_folder_name` from mock return values
- Handler/usecase tests covering `AcceptSuggestion`: remove
