package repository

import (
	"context"
	"fmt"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type folderShareRepository struct {
	db *gorm.DB
}

func NewFolderShareRepository(db *gorm.DB) usecase.FolderShareRepository {
	return &folderShareRepository{db: db}
}

func (r *folderShareRepository) Create(ctx context.Context, folderID uuid.UUID, token string) (*domain.FolderShare, error) {
	share := &domain.FolderShare{FolderID: folderID, Token: token}
	if err := r.db.WithContext(ctx).Create(share).Error; err != nil {
		return nil, fmt.Errorf("insert folder share: %w", err)
	}

	return share, nil
}

func (r *folderShareRepository) GetByFolderID(ctx context.Context, folderID uuid.UUID) (*domain.FolderShare, error) {
	var share domain.FolderShare
	if err := r.db.WithContext(ctx).
		Where("folder_id = ?", folderID).
		First(&share).Error; err != nil {
		return nil, fmt.Errorf("select folder share by folder id: %w", err)
	}

	return &share, nil
}

func (r *folderShareRepository) GetByToken(ctx context.Context, token string) (*domain.FolderShare, error) {
	var share domain.FolderShare
	if err := r.db.WithContext(ctx).
		Preload("Folder").
		Where("token = ?", token).
		First(&share).Error; err != nil {
		return nil, fmt.Errorf("select folder share by token: %w", err)
	}

	return &share, nil
}

func (r *folderShareRepository) DeleteByFolderID(ctx context.Context, folderID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Where("folder_id = ?", folderID).
		Delete(&domain.FolderShare{}).Error; err != nil {
		return fmt.Errorf("delete folder share: %w", err)
	}

	return nil
}

var _ usecase.FolderShareRepository = (*folderShareRepository)(nil)
