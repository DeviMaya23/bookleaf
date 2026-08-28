package usecase

import (
	"context"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

type TagRepository interface {
	Create(ctx context.Context, tag *domain.Tag) (*domain.Tag, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Tag, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Tag, error)
	Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, name string) (*domain.Tag, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	ReplaceImageTags(ctx context.Context, imageID uuid.UUID, tagIDs []uuid.UUID) error
	// DeleteAllByUserID permanently deletes all of a user's tags.
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}
