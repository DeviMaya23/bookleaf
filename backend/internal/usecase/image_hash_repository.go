package usecase

import (
	"context"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

type ImageHashRepository interface {
	ListUnhashed(ctx context.Context, limit int) ([]*domain.Image, error)
	UpdatePHash(ctx context.Context, id uuid.UUID, phash string) error
}
