## REMOVED Requirements

### Requirement: suggested_folder_name on GET /images/:id

**Reason**: The folder suggestion toast flow has been removed. `suggested_folder_name` was derived from `ai_labels[0].Description` and used exclusively by the `useVisionSuggestion` hook to show an accept/ignore toast. With that hook deleted, the field has no consumer and is removed from the API response.

**Migration**: Remove `SuggestedFolderName *string` from the `ImageDetail` usecase struct and the image handler response struct. Delete the `suggestedFolderName()` helper in `image_usecase.go`. Remove `suggested_folder_name` from the `GetImageResponse` type in `lib/images.ts`. No database change required — the field was always derived at read time from `ai_labels`.
