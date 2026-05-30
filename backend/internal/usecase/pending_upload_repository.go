package usecase

import (
	"context"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

type PendingUploadRepository interface {
	Create(ctx context.Context, pending *domain.PendingUpload) (*domain.PendingUpload, error)
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.PendingUpload, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListStale(ctx context.Context, olderThan time.Time) ([]*domain.PendingUpload, error)
	Transaction(ctx context.Context, fn func(pendingRepo PendingUploadRepository, imageRepo ImageRepository) error) error
}
