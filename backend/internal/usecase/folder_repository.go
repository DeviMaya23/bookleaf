package usecase

import (
	"context"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

type FolderRepository interface {
	Create(ctx context.Context, folder *domain.Folder) (*domain.Folder, error)
	List(ctx context.Context, userID string) ([]*domain.Folder, error)
	FindByName(ctx context.Context, userID, name string) (*domain.Folder, error)
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Folder, error)
	Update(ctx context.Context, folder *domain.Folder) (*domain.Folder, error)
	CountImagesByFolder(ctx context.Context, id uuid.UUID, userID string) (int, error)
	DeleteWithCascade(ctx context.Context, id uuid.UUID, userID string) error
}
