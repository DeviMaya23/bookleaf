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

func (r *folderRepository) List(ctx context.Context, userID string) ([]*domain.Folder, error) {
	var folders []*domain.Folder
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&folders).Error; err != nil {
		return nil, fmt.Errorf("list folders: %w", err)
	}

	return folders, nil
}

func (r *folderRepository) FindByName(ctx context.Context, userID, name string) (*domain.Folder, error) {
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

func (r *folderRepository) GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Folder, error) {
	var folder domain.Folder
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&folder).Error; err != nil {
		return nil, fmt.Errorf("select folder: %w", err)
	}

	return &folder, nil
}

func (r *folderRepository) Update(ctx context.Context, folder *domain.Folder) (*domain.Folder, error) {
	existing, err := r.GetByID(ctx, folder.ID, folder.UserID)
	if err != nil {
		return nil, err
	}

	existing.Name = folder.Name
	existing.ParentID = folder.ParentID

	if err := r.db.WithContext(ctx).
		Model(existing).
		Select("name", "parent_id").
		Updates(existing).Error; err != nil {
		return nil, fmt.Errorf("update folder: %w", err)
	}

	return existing, nil
}

func (r *folderRepository) CountImagesByFolder(ctx context.Context, id uuid.UUID, userID string) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Table("images").
		Where("folder_id = ? AND user_id = ?", id, userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count folder images: %w", err)
	}

	return int(count), nil
}

func (r *folderRepository) DeleteWithCascade(ctx context.Context, id uuid.UUID, userID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Folder{}).
			Where("parent_id = ? AND user_id = ?", id, userID).
			Update("parent_id", nil).Error; err != nil {
			return fmt.Errorf("clear child folders parent: %w", err)
		}

		if err := tx.Table("images").
			Where("folder_id = ? AND user_id = ?", id, userID).
			Update("folder_id", nil).Error; err != nil {
			return fmt.Errorf("clear images folder: %w", err)
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

var _ usecase.FolderRepository = (*folderRepository)(nil)
