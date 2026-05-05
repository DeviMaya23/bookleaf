package repository

import (
	"context"
	"fmt"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type imageRepository struct {
	db *gorm.DB
}

func NewImageRepository(db *gorm.DB) usecase.ImageRepository {
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

func (r *imageRepository) List(ctx context.Context, userID string, folderID *uuid.UUID) ([]*domain.Image, error) {
	var images []*domain.Image

	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC")

	if folderID != nil {
		query = query.Where("folder_id = ?", *folderID)
	}

	if err := query.Find(&images).Error; err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}

	return images, nil
}

func (r *imageRepository) GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error) {
	var image domain.Image
	if err := r.db.WithContext(ctx).
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

	return r.GetByID(ctx, id, userID)
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

func (r *imageRepository) ListTrashed(ctx context.Context, userID string) ([]*domain.Image, error) {
	var images []*domain.Image
	if err := r.db.WithContext(ctx).
		Unscoped().
		Where("deleted_at IS NOT NULL AND user_id = ?", userID).
		Order("deleted_at DESC").
		Find(&images).Error; err != nil {
		return nil, fmt.Errorf("list trashed images: %w", err)
	}

	return images, nil
}

var _ usecase.ImageRepository = (*imageRepository)(nil)
