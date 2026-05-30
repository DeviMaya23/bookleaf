package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pendingUploadRepository struct {
	db *gorm.DB
}

func NewPendingUploadRepository(db *gorm.DB) usecase.PendingUploadRepository {
	return &pendingUploadRepository{db: db}
}

func (r *pendingUploadRepository) Create(ctx context.Context, pending *domain.PendingUpload) (*domain.PendingUpload, error) {
	if err := r.db.WithContext(ctx).Create(pending).Error; err != nil {
		return nil, fmt.Errorf("insert pending upload: %w", err)
	}
	return pending, nil
}

func (r *pendingUploadRepository) GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.PendingUpload, error) {
	var pending domain.PendingUpload
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&pending).Error; err != nil {
		return nil, fmt.Errorf("select pending upload: %w", err)
	}
	return &pending, nil
}

func (r *pendingUploadRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&domain.PendingUpload{}).Error; err != nil {
		return fmt.Errorf("delete pending upload: %w", err)
	}
	return nil
}

func (r *pendingUploadRepository) ListStale(ctx context.Context, olderThan time.Time) ([]*domain.PendingUpload, error) {
	var pending []*domain.PendingUpload
	if err := r.db.WithContext(ctx).
		Where("created_at < ?", olderThan).
		Find(&pending).Error; err != nil {
		return nil, fmt.Errorf("list stale pending uploads: %w", err)
	}
	return pending, nil
}

func (r *pendingUploadRepository) Transaction(ctx context.Context, fn func(pendingRepo usecase.PendingUploadRepository, imageRepo usecase.ImageRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&pendingUploadRepository{db: tx}, &imageRepository{db: tx})
	})
}

var _ usecase.PendingUploadRepository = (*pendingUploadRepository)(nil)
