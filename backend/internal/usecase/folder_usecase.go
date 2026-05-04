package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/observability"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
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
	tel        *observability.Telemetry
}

func NewFolderUsecase(folderRepo FolderRepository, tel *observability.Telemetry) FolderUsecase {
	return &folderUsecase{
		folderRepo: folderRepo,
		tel:        tel,
	}
}

func (u *folderUsecase) Create(ctx context.Context, userID, name string, parentID *uuid.UUID) (*domain.Folder, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.CreateFolder")
	defer span.End()

	if strings.TrimSpace(name) == "" {
		span.RecordError(ErrInvalidFolderName)
		span.SetStatus(codes.Error, ErrInvalidFolderName.Error())
		return nil, ErrInvalidFolderName
	}

	folder, err := u.folderRepo.Create(ctx, &domain.Folder{
		UserID:   userID,
		Name:     name,
		ParentID: parentID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return folder, nil
}

func (u *folderUsecase) List(ctx context.Context, userID string) ([]*domain.Folder, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.ListFolders")
	defer span.End()

	folders, err := u.folderRepo.List(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return folders, nil
}

func (u *folderUsecase) GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Folder, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.GetFolder")
	defer span.End()

	folder, err := u.folderRepo.GetByID(ctx, id, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return folder, nil
}

func (u *folderUsecase) Update(ctx context.Context, id uuid.UUID, userID, name string, parentID *uuid.UUID) (*domain.Folder, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.UpdateFolder")
	defer span.End()

	if strings.TrimSpace(name) == "" {
		span.RecordError(ErrInvalidFolderName)
		span.SetStatus(codes.Error, ErrInvalidFolderName.Error())
		return nil, ErrInvalidFolderName
	}

	folder, err := u.folderRepo.Update(ctx, &domain.Folder{
		ID:       id,
		UserID:   userID,
		Name:     name,
		ParentID: parentID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return folder, nil
}

func (u *folderUsecase) Delete(ctx context.Context, id uuid.UUID, userID string) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.DeleteFolder")
	defer span.End()

	imageCount, err := u.folderRepo.CountImagesByFolder(ctx, id, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if err := u.folderRepo.DeleteWithCascade(ctx, id, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	observability.LoggerFromContext(ctx, u.tel.Logger).Info(
		"folder deleted",
		zap.String("event", "folder.mutated"),
		zap.String("folder_id", id.String()),
		zap.String("user_id", userID),
		zap.String("operation", "deleted"),
		zap.Int("image_count", imageCount),
	)

	return nil
}

var _ FolderUsecase = (*folderUsecase)(nil)
