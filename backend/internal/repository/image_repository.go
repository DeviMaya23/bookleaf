package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"roci.dev/fracdex"
)

type imageRepository struct {
	db *gorm.DB
}

func NewImageRepository(db *gorm.DB) *imageRepository {
	return &imageRepository{
		db: db,
	}
}

func (r *imageRepository) Create(ctx context.Context, image *domain.Image) (*domain.Image, error) {
	if err := r.db.WithContext(ctx).Create(image).Error; err != nil {
		return nil, fmt.Errorf("insert image: %w", err)
	}

	return image, nil
}

func (r *imageRepository) List(ctx context.Context, userID string, unfiled bool, folderIDs []uuid.UUID, tagIDs []uuid.UUID, mimeTypes []string, name *string, sortField *string, direction *string, cursor *usecase.ImageCursor, limit int) ([]*domain.Image, error) {
	var images []*domain.Image
	dispatch := usecase.ResolveSort(sortField, direction)

	query := r.db.WithContext(ctx).
		Where("images.user_id = ?", userID).
		Order(dispatch.OrderClause).
		Limit(limit + 1).
		Preload("Tags").
		Preload("ImageFolders")

	if unfiled {
		query = query.
			Joins("LEFT JOIN image_folders ON image_folders.image_id = images.id").
			Where("image_folders.image_id IS NULL")
	}

	if len(folderIDs) > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM image_folders WHERE image_folders.image_id = images.id AND image_folders.folder_id IN (?))", folderIDs)
	}

	if len(tagIDs) > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM image_tags WHERE image_tags.image_id = images.id AND image_tags.tag_id IN (?))", tagIDs)
	}

	if len(mimeTypes) > 0 {
		query = query.Where("images.mime_type IN (?)", mimeTypes)
	}

	if name != nil && *name != "" {
		query = query.Where("images.title ILIKE ?", "%"+*name+"%")
	}

	if cursor != nil {
		if dispatch.Column == "title" && cursor.Title != nil {
			query = query.Where(fmt.Sprintf("(images.title, images.id) %s (?, ?)", dispatch.WhereOperator), *cursor.Title, cursor.ID)
		} else {
			query = query.Where(fmt.Sprintf("(images.created_at, images.id) %s (?, ?)", dispatch.WhereOperator), cursor.CreatedAt, cursor.ID)
		}
	}

	if err := query.Find(&images).Error; err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}

	return images, nil
}

func (r *imageRepository) ListByFolder(ctx context.Context, userID string, folderID uuid.UUID, sortField *string, direction *string) ([]*domain.Image, error) {
	var images []*domain.Image
	dispatch := usecase.ResolveSort(sortField, direction)

	query := r.db.WithContext(ctx).
		Where("images.user_id = ?", userID).
		Joins("JOIN image_folders ON image_folders.image_id = images.id AND image_folders.folder_id = ?", folderID).
		Preload("Tags").
		Preload("ImageFolders")

	if sortField != nil {
		query = query.Order(dispatch.OrderClause)
	} else {
		query = query.Order("image_folders.position ASC")
	}

	if err := query.Find(&images).Error; err != nil {
		return nil, fmt.Errorf("list images by folder: %w", err)
	}

	return images, nil
}

func (r *imageRepository) SetImageFolder(ctx context.Context, imageID uuid.UUID, folderID *uuid.UUID) error {
	if folderID == nil {
		if err := r.db.WithContext(ctx).
			Where("image_id = ?", imageID).
			Delete(&domain.ImageFolder{}).Error; err != nil {
			return fmt.Errorf("delete image folder: %w", err)
		}
		return nil
	}

	var maxPos string
	r.db.WithContext(ctx).
		Model(&domain.ImageFolder{}).
		Where("folder_id = ?", *folderID).
		Select("COALESCE(MAX(position), '')").
		Scan(&maxPos)
	position, err := fracdex.KeyBetween(maxPos, "")
	if err != nil {
		return fmt.Errorf("generate position key: %w", err)
	}

	row := domain.ImageFolder{
		ImageID:  imageID,
		FolderID: *folderID,
		Position: position,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return fmt.Errorf("insert image folder: %w", err)
	}
	return nil
}

func (r *imageRepository) SyncImageFolders(ctx context.Context, imageID uuid.UUID, folderIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current []domain.ImageFolder
		if err := tx.Where("image_id = ?", imageID).Find(&current).Error; err != nil {
			return fmt.Errorf("list image folders: %w", err)
		}

		currentSet := make(map[uuid.UUID]struct{}, len(current))
		for _, row := range current {
			currentSet[row.FolderID] = struct{}{}
		}

		desiredSet := make(map[uuid.UUID]struct{}, len(folderIDs))
		toAdd := make([]uuid.UUID, 0)
		for _, folderID := range folderIDs {
			if _, seen := desiredSet[folderID]; seen {
				continue
			}
			desiredSet[folderID] = struct{}{}
			if _, exists := currentSet[folderID]; !exists {
				toAdd = append(toAdd, folderID)
			}
		}

		toDelete := make([]uuid.UUID, 0)
		for folderID := range currentSet {
			if _, keep := desiredSet[folderID]; !keep {
				toDelete = append(toDelete, folderID)
			}
		}

		if len(toDelete) > 0 {
			if err := tx.
				Where("image_id = ? AND folder_id IN ?", imageID, toDelete).
				Delete(&domain.ImageFolder{}).Error; err != nil {
				return fmt.Errorf("delete image folders: %w", err)
			}
		}

		for _, folderID := range toAdd {
			var maxPos string
			if err := tx.Model(&domain.ImageFolder{}).
				Where("folder_id = ?", folderID).
				Select("COALESCE(MAX(position), '')").
				Scan(&maxPos).Error; err != nil {
				return fmt.Errorf("select image folder position: %w", err)
			}
			position, err := fracdex.KeyBetween(maxPos, "")
			if err != nil {
				return fmt.Errorf("generate position key: %w", err)
			}
			row := domain.ImageFolder{
				ImageID:  imageID,
				FolderID: folderID,
				Position: position,
			}
			if err := tx.Create(&row).Error; err != nil {
				return fmt.Errorf("insert image folder: %w", err)
			}
		}

		return nil
	})
}

func (r *imageRepository) MoveImageFolder(ctx context.Context, imageID uuid.UUID, fromFolderID *uuid.UUID, toFolderID *uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if fromFolderID != nil {
			if err := tx.
				Where("image_id = ? AND folder_id = ?", imageID, *fromFolderID).
				Delete(&domain.ImageFolder{}).Error; err != nil {
				return fmt.Errorf("delete image folder: %w", err)
			}
		}

		if toFolderID != nil {
			var maxPos string
			if err := tx.Model(&domain.ImageFolder{}).
				Where("folder_id = ?", *toFolderID).
				Select("COALESCE(MAX(position), '')").
				Scan(&maxPos).Error; err != nil {
				return fmt.Errorf("select image folder position: %w", err)
			}
			position, err := fracdex.KeyBetween(maxPos, "")
			if err != nil {
				return fmt.Errorf("generate position key: %w", err)
			}
			row := domain.ImageFolder{
				ImageID:  imageID,
				FolderID: *toFolderID,
				Position: position,
			}
			if err := tx.
				Where(domain.ImageFolder{ImageID: imageID, FolderID: *toFolderID}).
				FirstOrCreate(&row).Error; err != nil {
				return fmt.Errorf("insert image folder: %w", err)
			}
		}

		return nil
	})
}

func (r *imageRepository) UpdateImageFolderPosition(ctx context.Context, imageID uuid.UUID, folderID uuid.UUID, position string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.ImageFolder{}).
		Where("image_id = ? AND folder_id = ?", imageID, folderID).
		Update("position", position)
	if result.Error != nil {
		return fmt.Errorf("update image folder position: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update image folder position: %w", gorm.ErrRecordNotFound)
	}
	return nil
}

func (r *imageRepository) GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error) {
	var image domain.Image
	if err := r.db.WithContext(ctx).
		Preload("Tags").
		Preload("ImageFolders").
		Where("id = ? AND user_id = ?", id, userID).
		First(&image).Error; err != nil {
		return nil, fmt.Errorf("select image: %w", err)
	}

	return &image, nil
}

func (r *imageRepository) GetDeletedByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error) {
	var image domain.Image
	if err := r.db.WithContext(ctx).
		Unscoped().
		Preload("ImageFolders").
		Where("id = ? AND user_id = ? AND deleted_at IS NOT NULL", id, userID).
		First(&image).Error; err != nil {
		return nil, fmt.Errorf("select deleted image: %w", err)
	}

	return &image, nil
}

func (r *imageRepository) UpdateThumbnailPath(ctx context.Context, id uuid.UUID, thumbnailPath string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Image{}).
		Where("id = ?", id).
		Update("thumbnail_path", thumbnailPath)
	if result.Error != nil {
		return fmt.Errorf("update image thumbnail_path: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update image thumbnail_path: %w", gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *imageRepository) UpdateAILabels(ctx context.Context, id uuid.UUID, labels json.RawMessage) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Image{}).
		Where("id = ?", id).
		Update("ai_labels", labels)
	if result.Error != nil {
		return fmt.Errorf("update image ai_labels: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update image ai_labels: %w", gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *imageRepository) Update(ctx context.Context, id uuid.UUID, userID string, fields map[string]any) (*domain.Image, error) {
	result := r.db.WithContext(ctx).
		Model(&domain.Image{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(fields)
	if result.Error != nil {
		return nil, fmt.Errorf("update image: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("update image: %w", gorm.ErrRecordNotFound)
	}

	return r.GetByID(ctx, id, userID) // includes Preload("Tags")
}

func (r *imageRepository) SoftDelete(ctx context.Context, id uuid.UUID, userID string) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&domain.Image{})
	if result.Error != nil {
		return fmt.Errorf("soft delete image: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("soft delete image: %w", gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *imageRepository) Restore(ctx context.Context, id uuid.UUID, userID string) error {
	result := r.db.WithContext(ctx).
		Unscoped().
		Model(&domain.Image{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NOT NULL", id, userID).
		Update("deleted_at", nil)
	if result.Error != nil {
		return fmt.Errorf("restore image: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("restore image: %w", gorm.ErrRecordNotFound)
	}

	return nil
}

func (r *imageRepository) ListTrashed(ctx context.Context, userID string, name *string, cursor *usecase.ImageCursor, limit int) ([]*domain.Image, error) {
	var images []*domain.Image

	query := r.db.WithContext(ctx).
		Unscoped().
		Where("deleted_at IS NOT NULL AND user_id = ?", userID).
		Order("deleted_at ASC, id ASC").
		Limit(limit + 1)

	if name != nil && *name != "" {
		query = query.Where("images.title ILIKE ?", "%"+*name+"%")
	}

	if cursor != nil && cursor.DeletedAt != nil {
		query = query.Where("(deleted_at, id) > (?, ?)", cursor.DeletedAt, cursor.ID)
	}

	if err := query.Find(&images).Error; err != nil {
		return nil, fmt.Errorf("list trashed images: %w", err)
	}

	return images, nil
}

func (r *imageRepository) ListAllTrashed(ctx context.Context, userID string) ([]*domain.Image, error) {
	var images []*domain.Image
	if err := r.db.WithContext(ctx).
		Unscoped().
		Where("deleted_at IS NOT NULL AND user_id = ?", userID).
		Order("deleted_at ASC, id ASC").
		Find(&images).Error; err != nil {
		return nil, fmt.Errorf("list all trashed images: %w", err)
	}
	return images, nil
}

func (r *imageRepository) CountByFolderID(ctx context.Context, folderID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Image{}).
		Joins("JOIN image_folders ON image_folders.image_id = images.id").
		Where("image_folders.folder_id = ?", folderID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count images by folder: %w", err)
	}

	return count, nil
}

func (r *imageRepository) ListExpiredTrash(ctx context.Context, olderThan time.Time) ([]*domain.Image, error) {
	var images []*domain.Image
	if err := r.db.WithContext(ctx).
		Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at < ?", olderThan).
		Find(&images).Error; err != nil {
		return nil, fmt.Errorf("list expired trash: %w", err)
	}
	return images, nil
}

func (r *imageRepository) HardDelete(ctx context.Context, id uuid.UUID, userID string) error {
	result := r.db.WithContext(ctx).
		Unscoped().
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&domain.Image{})
	if result.Error != nil {
		return fmt.Errorf("hard delete image: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("hard delete image: %w", gorm.ErrRecordNotFound)
	}
	return nil
}

var _ usecase.ImageRepository = (*imageRepository)(nil)
