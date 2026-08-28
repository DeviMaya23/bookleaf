package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type folderRepository struct {
	db *gorm.DB
}

func NewFolderRepository(db *gorm.DB) usecase.FolderRepository {
	return &folderRepository{
		db: db,
	}
}

func (r *folderRepository) Create(ctx context.Context, folder *domain.Folder) (*domain.Folder, error) {
	if err := r.db.WithContext(ctx).Create(folder).Error; err != nil {
		return nil, fmt.Errorf("insert folder: %w", err)
	}

	return folder, nil
}

func (r *folderRepository) List(ctx context.Context, userID uuid.UUID) ([]*domain.Folder, error) {
	var folders []*domain.Folder
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&folders).Error; err != nil {
		return nil, fmt.Errorf("list folders: %w", err)
	}

	return folders, nil
}

func (r *folderRepository) FindByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Folder, error) {
	var folder domain.Folder
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND name ILIKE ?", userID, strings.TrimSpace(name)).
		First(&folder).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("select folder by name: %w", err)
	}

	return &folder, nil
}

func (r *folderRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Folder, error) {
	var folder domain.Folder
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&folder).Error; err != nil {
		return nil, fmt.Errorf("select folder: %w", err)
	}

	return &folder, nil
}

func (r *folderRepository) Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, fields map[string]any) (*domain.Folder, error) {
	result := r.db.WithContext(ctx).
		Model(&domain.Folder{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(fields)
	if result.Error != nil {
		return nil, fmt.Errorf("update folder: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("update folder: %w", gorm.ErrRecordNotFound)
	}

	return r.GetByID(ctx, id, userID)
}

func (r *folderRepository) CountImagesByFolder(ctx context.Context, id uuid.UUID, userID uuid.UUID) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Image{}).
		Joins("JOIN image_folders ON image_folders.image_id = images.id").
		Where("image_folders.folder_id = ? AND images.user_id = ?", id, userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count folder images: %w", err)
	}

	return int(count), nil
}

func (r *folderRepository) DeleteWithCascade(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Folder{}).
			Where("parent_id = ? AND user_id = ?", id, userID).
			Update("parent_id", nil).Error; err != nil {
			return fmt.Errorf("clear child folders parent: %w", err)
		}

		result := tx.Where("id = ? AND user_id = ?", id, userID).Delete(&domain.Folder{})
		if result.Error != nil {
			return fmt.Errorf("delete folder: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("delete folder: %w", gorm.ErrRecordNotFound)
		}

		return nil
	})
}

func (r *folderRepository) ClearAllParents(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&domain.Folder{}).
		Where("user_id = ?", userID).
		Update("parent_id", nil).Error; err != nil {
		return fmt.Errorf("clear folder parents: %w", err)
	}
	return nil
}

func (r *folderRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&domain.Folder{}).Error; err != nil {
		return fmt.Errorf("delete all folders by user: %w", err)
	}
	return nil
}

var _ usecase.FolderRepository = (*folderRepository)(nil)
