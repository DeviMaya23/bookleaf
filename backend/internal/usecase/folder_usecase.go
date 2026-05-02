package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
)

var ErrInvalidFolderName = errors.New("folder name is required")

type FolderUsecase interface {
	Create(ctx context.Context, userID, name string, parentID *uuid.UUID) (*domain.Folder, error)
	List(ctx context.Context, userID string) ([]*domain.Folder, error)
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Folder, error)
	Update(ctx context.Context, id uuid.UUID, userID, name string, parentID *uuid.UUID) (*domain.Folder, error)
	Delete(ctx context.Context, id uuid.UUID, userID string) error
}

type folderUsecase struct {
	folderRepo FolderRepository
}

func NewFolderUsecase(folderRepo FolderRepository) FolderUsecase {
	return &folderUsecase{
		folderRepo: folderRepo,
	}
}

func (u *folderUsecase) Create(ctx context.Context, userID, name string, parentID *uuid.UUID) (*domain.Folder, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrInvalidFolderName
	}

	return u.folderRepo.Create(ctx, &domain.Folder{
		UserID:   userID,
		Name:     name,
		ParentID: parentID,
	})
}

func (u *folderUsecase) List(ctx context.Context, userID string) ([]*domain.Folder, error) {
	return u.folderRepo.List(ctx, userID)
}

func (u *folderUsecase) GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Folder, error) {
	return u.folderRepo.GetByID(ctx, id, userID)
}

func (u *folderUsecase) Update(ctx context.Context, id uuid.UUID, userID, name string, parentID *uuid.UUID) (*domain.Folder, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrInvalidFolderName
	}

	return u.folderRepo.Update(ctx, &domain.Folder{
		ID:       id,
		UserID:   userID,
		Name:     name,
		ParentID: parentID,
	})
}

func (u *folderUsecase) Delete(ctx context.Context, id uuid.UUID, userID string) error {
	return u.folderRepo.DeleteWithCascade(ctx, id, userID)
}

var _ FolderUsecase = (*folderUsecase)(nil)
