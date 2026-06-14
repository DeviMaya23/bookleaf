package usecase

import (
	"context"

	"github.com/devi/bookleaf/internal/domain"
)

type UserRepository interface {
	GetOrCreate(ctx context.Context, id string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	// MarkPendingKindeDeletion sets pending_kinde_deletion = true on the user's row.
	MarkPendingKindeDeletion(ctx context.Context, id string) error
	// HardDelete permanently deletes the user's row.
	HardDelete(ctx context.Context, id string) error
	// ListPendingKindeDeletion returns all users with pending_kinde_deletion = true.
	ListPendingKindeDeletion(ctx context.Context) ([]*domain.User, error)
	// Update applies a partial update to the user's row and returns the updated record.
	Update(ctx context.Context, id string, fields map[string]any) (*domain.User, error)
}
