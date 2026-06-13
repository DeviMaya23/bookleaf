package repository

import (
	"context"

	"github.com/devi/bookleaf/internal/usecase"
	"gorm.io/gorm"
)

type accountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) usecase.AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) Transaction(ctx context.Context, fn func(usecase.AccountRepos) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(usecase.AccountRepos{
			Images:         &imageRepository{db: tx},
			Folders:        &folderRepository{db: tx},
			Tags:           &tagRepository{db: tx},
			PendingUploads: &pendingUploadRepository{db: tx},
			Users:          &userRepository{db: tx},
		})
	})
}

var _ usecase.AccountRepository = (*accountRepository)(nil)
