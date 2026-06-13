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
	// ListAllByUserID returns all of a user's pending uploads.
	ListAllByUserID(ctx context.Context, userID string) ([]*domain.PendingUpload, error)
	// DeleteAllByUserID permanently deletes all of a user's pending uploads.
	DeleteAllByUserID(ctx context.Context, userID string) error
}
