## Context

The Vision AI labelling flow currently resolves folder existence eagerly inside `runVisionFlow` in `image_usecase.go`. It calls `FolderRepository.FindByName` to determine if the suggested folder already exists, then surfaces `FolderID`, `FolderName`, and `IsNew` to the client. The client is then expected to decide what to do with this info — but the server has already done the lookup work. This couples folder resolution timing to upload completion and forces the client to handle three fields just to accept a suggestion.

The new design defers folder resolution to an explicit client action: `POST /images/:id/accept-suggestion`. The upload flow only returns a name string from Vision; the new endpoint handles find-or-create when the user actually accepts.

## Goals / Non-Goals

**Goals:**
- Simplify `runVisionFlow` — return only `*string` (the suggested name), remove `FindByName` call
- Remove `FolderSuggestion` struct; `CompleteUploadResult` uses `SuggestedFolderName *string`
- Flatten the `POST /images/:id/complete` response to `suggested_folder_name: string | null`
- New `POST /images/:id/accept-suggestion` endpoint handles find-or-create folder and assigns it to the image

**Non-Goals:**
- Changing how Vision labels are fetched or stored (`UpdateAILabels` is unchanged)
- Removing `FindByName` from `FolderRepository` — it remains, but is now used only by `AcceptSuggestion`
- Any changes to the folder usecase or folder endpoints

## Decisions

### `AcceptSuggestion` lives on `ImageUsecase`, not `FolderUsecase`

The operation assigns a folder to an image — the primary entity being mutated is the image. `ImageUsecase` already holds references to both `ImageRepository` and `FolderRepository`, so no new dependency wiring is needed. Moving it to `FolderUsecase` would require injecting `ImageRepository` there, inverting a dependency that doesn't exist today.

**Alternative considered**: a dedicated `SuggestionUsecase` — rejected as over-engineering for a single method.

### `FindByName` stays on `FolderRepository`

The method is still needed by `AcceptSuggestion` in `imageUsecase`. It is not removed from the interface or SQL implementation — only its call site moves (from `runVisionFlow` to `AcceptSuggestion`).

### `runVisionFlow` return type becomes `(*string, string)`

`runVisionFlow` currently returns `(*FolderSuggestion, string)`. With `FolderSuggestion` gone, it becomes `(suggestedName *string, warning string)`. The caller (`CompleteUpload`) assigns the result to `result.SuggestedFolderName`.

### `accept-suggestion` creates the folder if it does not exist

When the client calls `POST /images/:id/accept-suggestion`, the usecase calls `FolderRepository.FindByName`. If a matching folder exists it is used; if not, a new folder is created via `FolderRepository.Create`. The image is then updated with the resolved folder ID via `ImageRepository.Update`. This keeps all side effects server-side and the client request simple.

**Alternative considered**: client passes `folder_id` directly — rejected because that requires the client to know whether the folder exists, which was the original problem.

## Risks / Trade-offs

- **Breaking API change** → clients consuming `folder_suggestion.folder_id` or `folder_suggestion.is_new` from `POST /images/:id/complete` will break. Mitigation: coordinate with frontend before deploying; this is an internal API with a single client.
- **Folder created on accept** → if the user accepts the suggestion and a folder is created, then immediately rejects the assignment, an empty folder is left behind. Mitigation: acceptable for now; folder cleanup is out of scope.
- **Race on folder creation** → two concurrent accepts for the same folder name could create duplicates. Mitigation: `FindByName` runs before `Create`; a unique constraint on `(user_id, name)` at the DB level would fully protect against this, but that pre-exists this change.
