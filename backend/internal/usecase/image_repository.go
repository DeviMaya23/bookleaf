package usecase

import (
	"context"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

type ImageRepository interface {
	Create(ctx context.Context, image *domain.Image) (*domain.Image, error)
	// List returns non-deleted images ordered by (created_at DESC, id DESC), fetching limit+1 rows.
	// unfiled filters images with no image_folders row via LEFT JOIN; folderIDs/tagIDs filter via
	// correlated EXISTS subqueries (match-any, at most once per image); mimeTypes filters via IN.
	// Results include Tags and ImageFolders preloaded.
	List(ctx context.Context, userID string, unfiled bool, folderIDs []uuid.UUID, tagIDs []uuid.UUID, mimeTypes []string, name *string, sortField *string, direction *string, cursor *ImageCursor, limit int) ([]*domain.Image, error)
	// ListByFolder returns all non-deleted images in folderID owned by userID, ordered by
	// image_folders.position ASC (or by sortField/direction when provided). No cursor or limit.
	// Results include Tags and ImageFolders preloaded.
	ListByFolder(ctx context.Context, userID string, folderID uuid.UUID, sortField *string, direction *string) ([]*domain.Image, error)
	// GetByID returns a non-deleted image with Tags and ImageFolders preloaded.
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)
	// Update selectively updates scalar fields for the image. folder_id is not a valid key; use SyncImageFolders or SetImageFolder.
	Update(ctx context.Context, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error)
	// SetImageFolder assigns or removes a folder membership. folderID nil removes the row; non-nil upserts it.
	SetImageFolder(ctx context.Context, imageID uuid.UUID, folderID *uuid.UUID) error
	// AddImageToFolder idempotently inserts an image_folders row for (imageID, folderID), appending a
	// fracdex position after the current max for the folder. No-op if the row already exists.
	AddImageToFolder(ctx context.Context, imageID, folderID uuid.UUID) error
	// FilterOwnedImageIDs returns the subset of ids that exist (non-deleted) and belong to userID.
	FilterOwnedImageIDs(ctx context.Context, ids []uuid.UUID, userID string) ([]uuid.UUID, error)
	// SyncImageFolders diffs current memberships against folderIDs and applies deletes/inserts in a transaction.
	SyncImageFolders(ctx context.Context, imageID uuid.UUID, folderIDs []uuid.UUID) error
	// MoveImageFolder atomically removes image from fromFolderID and adds to toFolderID.
	MoveImageFolder(ctx context.Context, imageID uuid.UUID, fromFolderID *uuid.UUID, toFolderID *uuid.UUID) error
	// UpdateImageFolderPosition updates the position of an image within a specific folder.
	// Returns ErrRecordNotFound if no row exists for (imageID, folderID).
	UpdateImageFolderPosition(ctx context.Context, imageID uuid.UUID, folderID uuid.UUID, position string) error
	// ListAllByUserID returns all of a user's images, unscoped (including soft-deleted/trashed).
	ListAllByUserID(ctx context.Context, userID string) ([]*domain.Image, error)
	// HardDeleteAllByUserID permanently deletes all of a user's images, including soft-deleted ones.
	HardDeleteAllByUserID(ctx context.Context, userID string) error
}
