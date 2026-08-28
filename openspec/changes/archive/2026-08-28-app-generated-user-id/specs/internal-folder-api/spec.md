## MODIFIED Requirements

### Requirement: InternalShareUsecase — GetPublicFoldersByUser

`shareUsecase` SHALL implement:

```go
GetPublicFoldersByUser(ctx context.Context, userID uuid.UUID) ([]FolderShareSummary, error)
```

The parameter type changes from `string` to `uuid.UUID` to match the internal identity space. It SHALL call `FolderShareRepository.ListByUserID(ctx, userID)` and map each result to a `FolderShareSummary`. If no rows exist it SHALL return an empty slice and no error.

#### Scenario: Returns all public folders for a user

- **WHEN** `GetPublicFoldersByUser` is called with an internal UUID for a user who has two shared folders
- **THEN** it returns a slice of two `FolderShareSummary` values with the correct `FolderID`, `Token`, and `FolderName` for each

#### Scenario: Returns empty slice when user has no public folders

- **WHEN** `GetPublicFoldersByUser` is called with an internal UUID for a user with no `folder_shares` rows
- **THEN** it returns an empty slice and no error

---

### Requirement: InternalHandler

The `handler` package SHALL define `InternalHandler` in `handler/internal.go`. The `InternalShareUsecase` interface SHALL be updated to accept `uuid.UUID` for `GetPublicFoldersByUser`. `InternalHandler` SHALL additionally depend on a `UserResolver` interface:

```go
type UserResolver interface {
    GetByIDPSubject(ctx context.Context, idpSubject string) (*domain.User, error)
}
```

`ListPublicFolders(c echo.Context) error` SHALL:
1. Read `:user_id` path param (the Kinde subject sent by the caller)
2. Call `UserResolver.GetByIDPSubject(ctx, userID)` to resolve to an internal UUID
3. On `ErrUserNotFound`: return `200 OK` with `{"folder_list": []}` (the caller's user is unknown to Bookleaf — treat as having no public folders)
4. On success: call `GetPublicFoldersByUser(ctx, user.ID)` and return the result

#### Scenario: ListPublicFolders resolves Kinde subject to UUID before querying

- **WHEN** `GET /internal/users/:user_id/public-folders` is called with a Kinde subject and the user exists
- **THEN** the handler resolves the subject to an internal UUID
- **AND** queries folders by that UUID
- **AND** returns `200 OK` with the folder list

#### Scenario: ListPublicFolders returns empty list for unknown Kinde subject

- **WHEN** `GET /internal/users/:user_id/public-folders` is called with a Kinde subject that has no matching user row
- **THEN** the response is `200 OK` with `{"folder_list": []}`

#### Scenario: ListPublicFolders returns empty list when user has no public folders

- **WHEN** `GET /internal/users/:user_id/public-folders` is called for a known user with no `folder_shares` rows
- **THEN** the response is `200 OK` with `{"folder_list": []}`
