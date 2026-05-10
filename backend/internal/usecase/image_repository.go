package usecase

import (
	"context"
	"encoding/json"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

type ImageRepository interface {
	Create(ctx context.Context, image *domain.Image) (*domain.Image, error)
	List(ctx context.Context, userID string, folderID *uuid.UUID) ([]*domain.Image, error)
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)
	GetDeletedByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)
	UpdateThumbnailPath(ctx context.Context, id uuid.UUID, thumbnailPath string) error
	UpdateAILabels(ctx context.Context, id uuid.UUID, labels json.RawMessage) error
	Update(ctx context.Context, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error)
	SoftDelete(ctx context.Context, id uuid.UUID, userID string) error
	Restore(ctx context.Context, id uuid.UUID, userID string) error
	ListTrashed(ctx context.Context, userID string) ([]*domain.Image, error)
	CountByFolderID(ctx context.Context, folderID uuid.UUID) (int64, error)
}
