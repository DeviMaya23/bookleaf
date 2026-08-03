package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/devi/bookleaf/internal/domain"
)

// ErrUserNotFound is returned by UserRepository.GetByID when the user does not exist.
var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	GetOrCreate(ctx context.Context, id string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	// SetAccountState updates account_state for the given user. Returns ErrUserNotFound if no row matches.
	SetAccountState(ctx context.Context, id string, state domain.AccountState) error
	// MarkPurged sets account_state = 'purged' and purged_at = purgedAt. Returns ErrUserNotFound if no row matches.
	MarkPurged(ctx context.Context, id string, purgedAt time.Time) error
	// ListByAccountState returns all users with the given account_state.
	ListByAccountState(ctx context.Context, state domain.AccountState) ([]*domain.User, error)
	// ListPurgedBefore returns all users with account_state = 'purged' and purged_at < threshold.
	ListPurgedBefore(ctx context.Context, threshold time.Time) ([]*domain.User, error)
	// HardDelete permanently deletes the user's row.
	HardDelete(ctx context.Context, id string) error
	// Update applies a partial update to the user's row and returns the updated record.
	Update(ctx context.Context, id string, fields map[string]any) (*domain.User, error)
}
