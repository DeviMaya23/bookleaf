package usecase

import (
	"context"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

type ImageFolderRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Folder, error)
	FindByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Folder, error)
	Create(ctx context.Context, folder *domain.Folder) (*domain.Folder, error)
}
