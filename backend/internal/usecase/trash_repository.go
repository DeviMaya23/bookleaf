package usecase

import (
	"context"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

type TrashRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)
	GetDeletedByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)
	SoftDelete(ctx context.Context, id uuid.UUID, userID string) error
	Restore(ctx context.Context, id uuid.UUID, userID string) error
	ListTrashed(ctx context.Context, userID string, name *string, cursor *ImageCursor, limit int) ([]*domain.Image, error)
	ListAllTrashed(ctx context.Context, userID string) ([]*domain.Image, error)
	ListExpiredTrash(ctx context.Context, olderThan time.Time) ([]*domain.Image, error)
	HardDelete(ctx context.Context, id uuid.UUID, userID string) error
}
