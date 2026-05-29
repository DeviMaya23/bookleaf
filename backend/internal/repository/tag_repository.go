package repository

import (
	"context"
	"fmt"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) usecase.TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) Create(ctx context.Context, tag *domain.Tag) (*domain.Tag, error) {
	if err := r.db.WithContext(ctx).Create(tag).Error; err != nil {
		return nil, fmt.Errorf("insert tag: %w", err)
	}
	return tag, nil
}

func (r *tagRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.Tag, error) {
	var tags []*domain.Tag
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("name ASC").
		Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	return tags, nil
}

func (r *tagRepository) GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Tag, error) {
	var tag domain.Tag
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&tag).Error; err != nil {
		return nil, fmt.Errorf("select tag: %w", err)
	}
	return &tag, nil
}

func (r *tagRepository) Update(ctx context.Context, id uuid.UUID, userID string, name string) (*domain.Tag, error) {
	result := r.db.WithContext(ctx).
		Model(&domain.Tag{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("name", name)
	if result.Error != nil {
		return nil, fmt.Errorf("update tag: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("update tag: %w", gorm.ErrRecordNotFound)
	}
	return r.GetByID(ctx, id, userID)
}

func (r *tagRepository) Delete(ctx context.Context, id uuid.UUID, userID string) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&domain.Tag{})
	if result.Error != nil {
		return fmt.Errorf("delete tag: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("delete tag: %w", gorm.ErrRecordNotFound)
	}
	return nil
}

func (r *tagRepository) ReplaceImageTags(ctx context.Context, imageID uuid.UUID, tagIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM image_tags WHERE image_id = ?", imageID).Error; err != nil {
			return fmt.Errorf("clear image tags: %w", err)
		}
		if len(tagIDs) == 0 {
			return nil
		}
		type imageTag struct {
			ImageID uuid.UUID
			TagID   uuid.UUID
		}
		rows := make([]imageTag, len(tagIDs))
		for i, tagID := range tagIDs {
			rows[i] = imageTag{ImageID: imageID, TagID: tagID}
		}
		if err := tx.Table("image_tags").Create(&rows).Error; err != nil {
			return fmt.Errorf("insert image tags: %w", err)
		}
		return nil
	})
}

var _ usecase.TagRepository = (*tagRepository)(nil)
