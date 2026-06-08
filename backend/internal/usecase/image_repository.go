package usecase

import (
	"context"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

type ImageRepository interface {
	Create(ctx context.Context, image *domain.Image) (*domain.Image, error)
	// List returns non-deleted images ordered by (created_at DESC, id DESC), fetching limit+1 rows.
	// folderID filters via JOIN on image_folders; unfiled filters images with no image_folders row via LEFT JOIN.
	// Results include Tags and ImageFolders preloaded.
	List(ctx context.Context, userID string, folderID *uuid.UUID, unfiled bool, tagID *uuid.UUID, name *string, sortField *string, direction *string, cursor *ImageCursor, limit int) ([]*domain.Image, error)
	// GetByID returns a non-deleted image with Tags and ImageFolders preloaded.
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)
	// Update selectively updates scalar fields for the image. folder_id is not a valid key; use SyncImageFolders or SetImageFolder.
	Update(ctx context.Context, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error)
	// SetImageFolder assigns or removes a folder membership. folderID nil removes the row; non-nil upserts it.
	SetImageFolder(ctx context.Context, imageID uuid.UUID, folderID *uuid.UUID) error
	// SyncImageFolders diffs current memberships against folderIDs and applies deletes/inserts in a transaction.
	SyncImageFolders(ctx context.Context, imageID uuid.UUID, folderIDs []uuid.UUID) error
	// MoveImageFolder atomically removes image from fromFolderID and adds to toFolderID.
	MoveImageFolder(ctx context.Context, imageID uuid.UUID, fromFolderID *uuid.UUID, toFolderID *uuid.UUID) error
	// UpdateImageFolderPosition updates the position of an image within a specific folder.
	// Returns ErrRecordNotFound if no row exists for (imageID, folderID).
	UpdateImageFolderPosition(ctx context.Context, imageID uuid.UUID, folderID uuid.UUID, position string) error
}
