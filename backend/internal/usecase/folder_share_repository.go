package usecase

import (
	"context"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

type FolderShareListItem struct {
	FolderID   uuid.UUID
	Token      string
	FolderName string
}

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
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*FolderShareListItem, error)
}
