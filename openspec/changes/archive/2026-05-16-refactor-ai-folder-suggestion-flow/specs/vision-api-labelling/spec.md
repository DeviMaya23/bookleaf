## REMOVED Requirements

### Requirement: FolderSuggestion Resolution

**Reason**: Folder existence resolution is deferred to the new `accept-suggestion` endpoint. The Vision flow no longer needs to call `FolderRepository.FindByName` — it only surfaces the top label name as a plain string.

**Migration**: `runVisionFlow` now returns `(*string, string)` instead of `(*FolderSuggestion, string)`. Callers receive a `SuggestedFolderName *string` on `CompleteUploadResult`. The `FolderSuggestion` struct is deleted.
