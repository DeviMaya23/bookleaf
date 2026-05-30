package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

type ImageRepository interface {
	Create(ctx context.Context, image *domain.Image) (*domain.Image, error)
	// List returns non-deleted images ordered by (created_at DESC, id DESC), fetching limit+1 rows.
	// folderID filters via JOIN on image_folders; unfiled filters images with no image_folders row via LEFT JOIN.
	// Results include Tags and ImageFolders preloaded.
	List(ctx context.Context, userID string, folderID *uuid.UUID, unfiled bool, tagID *uuid.UUID, cursor *ImageCursor, limit int) ([]*domain.Image, error)
	// GetByID returns a non-deleted image with Tags and ImageFolders preloaded.
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)
	// GetDeletedByID returns a soft-deleted image with ImageFolders preloaded.
	GetDeletedByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)
	UpdateThumbnailPath(ctx context.Context, id uuid.UUID, thumbnailPath string) error
	UpdateAILabels(ctx context.Context, id uuid.UUID, labels json.RawMessage) error
	// Update selectively updates scalar fields for the image. folder_id is not a valid key; use SetImageFolder.
	Update(ctx context.Context, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error)
	// SetImageFolder assigns or removes a folder membership. folderID nil removes the row; non-nil upserts it.
	SetImageFolder(ctx context.Context, imageID uuid.UUID, folderID *uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID, userID string) error
	Restore(ctx context.Context, id uuid.UUID, userID string) error
	ListTrashed(ctx context.Context, userID string, cursor *ImageCursor, limit int) ([]*domain.Image, error)
	// CountByFolderID counts non-deleted images with a row in image_folders for the given folder.
	// Queries via Model(&domain.Image{}) + JOIN image_folders so soft-delete scope applies automatically.
	CountByFolderID(ctx context.Context, folderID uuid.UUID) (int64, error)
	ListExpiredTrash(ctx context.Context, olderThan time.Time) ([]*domain.Image, error)
	HardDelete(ctx context.Context, id uuid.UUID, userID string) error
}
