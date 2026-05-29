package usecase

import (
	"context"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

type TagRepository interface {
	Create(ctx context.Context, tag *domain.Tag) (*domain.Tag, error)
	ListByUserID(ctx context.Context, userID string) ([]*domain.Tag, error)
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Tag, error)
	Update(ctx context.Context, id uuid.UUID, userID string, name string) (*domain.Tag, error)
	Delete(ctx context.Context, id uuid.UUID, userID string) error
	ReplaceImageTags(ctx context.Context, imageID uuid.UUID, tagIDs []uuid.UUID) error
}
