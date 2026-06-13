package usecase

import "context"

// AccountRepos bundles the repository interfaces needed for the account-deletion
// data wipe, all bound to the same database transaction.
type AccountRepos struct {
	Images         ImageRepository
	Folders        FolderRepository
	Tags           TagRepository
	PendingUploads PendingUploadRepository
	Users          UserRepository
}

// AccountRepository provides a single database transaction spanning the
// repositories needed to wipe a user's account data.
type AccountRepository interface {
	Transaction(ctx context.Context, fn func(AccountRepos) error) error
}
