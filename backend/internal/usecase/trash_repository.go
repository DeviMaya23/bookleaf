package usecase

import (
	"context"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

type TrashRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Image, error)
	GetDeletedByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Image, error)
	SoftDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	ListTrashed(ctx context.Context, userID uuid.UUID, name *string, sortField *string, direction *string, cursor *ImageCursor, limit int) ([]*domain.Image, error)
	ListAllTrashed(ctx context.Context, userID uuid.UUID) ([]*domain.Image, error)
	ListExpiredTrash(ctx context.Context, olderThan time.Time) ([]*domain.Image, error)
	HardDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	// FilterOwnedImageIDs returns the subset of ids that exist (non-deleted) and belong to userID.
	FilterOwnedImageIDs(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) ([]uuid.UUID, error)
}
