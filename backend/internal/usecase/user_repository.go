package usecase

import (
	"context"

	"github.com/devi/bookleaf/internal/domain"
)

type UserRepository interface {
	GetOrCreate(ctx context.Context, id string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
}
