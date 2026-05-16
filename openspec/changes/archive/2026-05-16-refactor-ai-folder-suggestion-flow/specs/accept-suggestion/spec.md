## ADDED Requirements

### Requirement: AcceptSuggestion Usecase Method

The system SHALL add an `AcceptSuggestion` method to the `ImageUsecase` interface in `internal/usecase/`:

```go
AcceptSuggestion(ctx context.Context, imageID uuid.UUID, userID string, suggestedFolderName string) error
```

The implementation SHALL:
1. Verify the image exists and belongs to `userID` via `ImageRepository.GetByID`
2. Call `FolderRepository.FindByName` with `userID` and `suggestedFolderName` (case-insensitive)
3. If no matching folder exists, create one via `FolderRepository.Create` with the given name and `userID`
4. Assign the resolved folder ID to the image via `ImageRepository.Update` with `FolderID`

#### Scenario: Suggestion accepted — existing folder matched

- **WHEN** `AcceptSuggestion` is called with a name matching an existing folder
- **THEN** no new folder is created
- **AND** the image's `folder_id` is updated to the matched folder's ID
- **AND** `nil` is returned

#### Scenario: Suggestion accepted — no matching folder, new one created

- **WHEN** `AcceptSuggestion` is called with a name that matches no existing folder
- **THEN** a new folder is created with the given name
- **AND** the image's `folder_id` is updated to the new folder's ID
- **AND** `nil` is returned

#### Scenario: Image not found or not owned by user

- **WHEN** `AcceptSuggestion` is called with an image ID that does not exist or belongs to a different user
- **THEN** an error is returned and no folder is created or assigned

---

### Requirement: POST /images/:id/accept-suggestion Route

The system SHALL register `POST /images/:id/accept-suggestion` on the protected Echo group in `main.go`.

The handler SHALL:
- Parse `:id` as a UUID; return `400 Bad Request` on parse failure
- Bind the request body `{ "suggested_folder_name": "<string>" }`; return `400 Bad Request` if the field is missing or empty
- Extract `userID` from the Kinde auth context
- Call `imageUsecase.AcceptSuggestion`
- Return `204 No Content` on success

#### Scenario: Suggestion accepted successfully

- **WHEN** an authenticated `POST /images/:id/accept-suggestion` request is made with a valid `suggested_folder_name`
- **THEN** the response is `204 No Content`
- **AND** the image's `folder_id` is set to the resolved folder

#### Scenario: Missing suggested_folder_name

- **WHEN** the request body omits `suggested_folder_name` or provides an empty string
- **THEN** the response is `400 Bad Request`

#### Scenario: Image not found

- **WHEN** `AcceptSuggestion` returns a not-found error
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request

- **WHEN** the request does not include a valid Bearer token
- **THEN** the response is `401 Unauthorized`
