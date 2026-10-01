package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var ErrInvalidFolderName = errors.New("folder name is required")
var ErrInvalidFolderIcon = errors.New("folder icon is not in the allowlist")

type UpdateFolderParams struct {
	Name        *string
	ParentID    **uuid.UUID
	Description **string
	Icon        **string
}

type FolderImageRepository interface {
	CountByFolderID(ctx context.Context, folderID uuid.UUID) (int64, error)
	ListByFolder(ctx context.Context, userID uuid.UUID, folderID uuid.UUID, sortField *string, direction *string) ([]*domain.Image, error)
}

type FolderDetail struct {
	Folder     *domain.Folder
	ImageCount int64
}

type folderUsecase struct {
	folderRepo FolderRepository
	imageRepo  FolderImageRepository
	store      StorageService
	tel        *observability.Telemetry
}

func NewFolderUsecase(folderRepo FolderRepository, imageRepo FolderImageRepository, store StorageService, tel *observability.Telemetry) *folderUsecase {
	return &folderUsecase{
		folderRepo: folderRepo,
		imageRepo:  imageRepo,
		store:      store,
		tel:        tel,
	}
}

func (u *folderUsecase) Create(ctx context.Context, userID uuid.UUID, name string, parentID *uuid.UUID, description *string, icon *string) (*domain.Folder, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.CreateFolder")
	defer span.End()

	if strings.TrimSpace(name) == "" {
		span.RecordError(ErrInvalidFolderName)
		span.SetStatus(codes.Error, ErrInvalidFolderName.Error())
		return nil, ErrInvalidFolderName
	}

	if icon != nil && !IsValidFolderIcon(*icon) {
		span.RecordError(ErrInvalidFolderIcon)
		span.SetStatus(codes.Error, ErrInvalidFolderIcon.Error())
		return nil, ErrInvalidFolderIcon
	}

	folder, err := u.folderRepo.Create(ctx, &domain.Folder{
		UserID:      userID,
		Name:        name,
		ParentID:    parentID,
		Description: description,
		Icon:        icon,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return folder, nil
}

func (u *folderUsecase) List(ctx context.Context, userID uuid.UUID) ([]*domain.Folder, error) {
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

func (u *folderUsecase) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*FolderDetail, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.GetFolder")
	defer span.End()

	folder, err := u.folderRepo.GetByID(ctx, id, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	count, err := u.imageRepo.CountByFolderID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &FolderDetail{
		Folder:     folder,
		ImageCount: count,
	}, nil
}

func (u *folderUsecase) Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, params UpdateFolderParams) (*domain.Folder, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.UpdateFolder")
	defer span.End()

	if params.Name != nil && strings.TrimSpace(*params.Name) == "" {
		span.RecordError(ErrInvalidFolderName)
		span.SetStatus(codes.Error, ErrInvalidFolderName.Error())
		return nil, ErrInvalidFolderName
	}

	if params.Icon != nil && *params.Icon != nil && !IsValidFolderIcon(**params.Icon) {
		span.RecordError(ErrInvalidFolderIcon)
		span.SetStatus(codes.Error, ErrInvalidFolderIcon.Error())
		return nil, ErrInvalidFolderIcon
	}

	fields := make(map[string]any)
	if params.Name != nil {
		fields["name"] = *params.Name
	}
	if params.ParentID != nil {
		fields["parent_id"] = *params.ParentID
	}
	if params.Description != nil {
		fields["description"] = *params.Description
	}
	if params.Icon != nil {
		fields["icon"] = *params.Icon
	}

	folder, err := u.folderRepo.Update(ctx, id, userID, fields)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return folder, nil
}

func (u *folderUsecase) ExportFolder(ctx context.Context, folderID uuid.UUID, userID uuid.UUID, w io.Writer) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.ExportFolder")
	defer span.End()

	images, err := u.imageRepo.ListByFolder(ctx, userID, folderID, nil, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("list images by folder: %w", err)
	}

	if err := writeImagesToZip(ctx, images, u.store, w); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (u *folderUsecase) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
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
		zap.String("user_id", userID.String()),
		zap.String("operation", "deleted"),
		zap.Int("image_count", imageCount),
	)

	return nil
}
