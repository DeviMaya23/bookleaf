## MODIFIED Requirements

### Requirement: FolderShareRepository Interface and Implementation

The `usecase` package SHALL define a `FolderShareRepository` interface:

The `usecase` package SHALL also define:

```go
type FolderShareListItem struct {
    FolderID   uuid.UUID
    Token      string
    FolderName string
}
```

```go
type FolderShareRepository interface {
    Create(ctx context.Context, folderID uuid.UUID, token string) (*domain.FolderShare, error)
    GetByFolderID(ctx context.Context, folderID uuid.UUID) (*domain.FolderShare, error)
    // GetByFolderIDWithFolder preloads the associated Folder.
    GetByFolderIDWithFolder(ctx context.Context, folderID uuid.UUID) (*domain.FolderShare, error)
    // GetByToken preloads the associated Folder.
    GetByToken(ctx context.Context, token string) (*domain.FolderShare, error)
    // DeleteByFolderID does not error when no row exists for the given folder.
    DeleteByFolderID(ctx context.Context, folderID uuid.UUID) error
    // ListByUserID returns all folder_shares rows whose associated folder belongs to userID,
    // including the folder name. Returns an empty slice (not an error) when no rows match.
    ListByUserID(ctx context.Context, userID string) ([]*FolderShareListItem, error)
}
```

`GetByFolderID` and `GetByToken` SHALL return `gorm.ErrRecordNotFound` when no row matches. `GetByToken` and `GetByFolderIDWithFolder` SHALL preload the associated `Folder`. `DeleteByFolderID` SHALL not error when no row exists for the given folder. `ListByUserID` SHALL return an empty slice (not an error) when no rows match.

The system SHALL implement this interface in `repository/folder_share_repository.go` using GORM. `ListByUserID` SHALL use an explicit `SELECT folder_shares.folder_id, folder_shares.token, folders.name AS folder_name` with `Table("folder_shares")`, JOIN `folder_shares` with `folders` on `folder_id`, filter by `folders.user_id = ?`, and `Scan` into `[]*usecase.FolderShareListItem`. `GetByFolderIDWithFolder` SHALL Preload `"Folder"` before querying by `folder_id`.

#### Scenario: GetByFolderID returns not-found for an unshared folder

- **WHEN** `GetByFolderID` is called with a folder ID that has no `folder_shares` row
- **THEN** it returns `gorm.ErrRecordNotFound`

#### Scenario: GetByToken preloads the folder

- **WHEN** `GetByToken` is called with a valid token
- **THEN** the returned `FolderShare.Folder` is populated with the associated folder's `Name`, `Description`, and `UserID`

#### Scenario: GetByToken returns not-found for an unknown token

- **WHEN** `GetByToken` is called with a token that does not exist
- **THEN** it returns `gorm.ErrRecordNotFound`

#### Scenario: DeleteByFolderID is idempotent

- **WHEN** `DeleteByFolderID` is called for a folder with no existing `folder_shares` row
- **THEN** it returns no error

#### Scenario: GetByFolderIDWithFolder preloads the folder

- **WHEN** `GetByFolderIDWithFolder` is called with a folder ID that has a `folder_shares` row
- **THEN** the returned `FolderShare.Folder` is populated with the associated folder's `Name`, `Description`, and `UserID`

#### Scenario: GetByFolderIDWithFolder returns not-found for an unshared folder

- **WHEN** `GetByFolderIDWithFolder` is called with a folder ID that has no `folder_shares` row
- **THEN** it returns `gorm.ErrRecordNotFound`

#### Scenario: ListByUserID returns all shares for a user

- **WHEN** `ListByUserID` is called for a user who has two shared folders
- **THEN** it returns a slice of two `FolderShareListItem` values, each with the correct `FolderID`, `Token`, and `FolderName`

#### Scenario: ListByUserID returns empty slice for a user with no public folders

- **WHEN** `ListByUserID` is called for a user with no `folder_shares` rows
- **THEN** it returns an empty slice and no error
