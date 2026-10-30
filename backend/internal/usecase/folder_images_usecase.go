package usecase

import (
	"context"
	"fmt"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
)

type FolderImageThumbnail struct {
	ImageID      uuid.UUID
	ThumbnailURL string
}

type FolderImagesUsecase interface {
	ListFolderImages(ctx context.Context, userID uuid.UUID, folderID uuid.UUID) ([]FolderImageThumbnail, error)
}

// folderImagesFolder is the narrow folder lookup FolderImagesUsecase needs.
type folderImagesFolder interface {
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Folder, error)
}

// folderImagesImage is the narrow image listing FolderImagesUsecase needs.
type folderImagesImage interface {
	ListByFolder(ctx context.Context, userID uuid.UUID, folderID uuid.UUID, sortField *string, direction *string) ([]*domain.Image, error)
}

type folderImagesUsecase struct {
	folderRepo folderImagesFolder
	imageRepo  folderImagesImage
	store      StorageService
	tel        *observability.Telemetry
}

func NewFolderImagesUsecase(folderRepo folderImagesFolder, imageRepo folderImagesImage, store StorageService, tel *observability.Telemetry) FolderImagesUsecase {
	return &folderImagesUsecase{
		folderRepo: folderRepo,
		imageRepo:  imageRepo,
		store:      store,
		tel:        tel,
	}
}

func (u *folderImagesUsecase) ListFolderImages(ctx context.Context, userID uuid.UUID, folderID uuid.UUID) ([]FolderImageThumbnail, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.ListFolderImages")
	defer span.End()

	if _, err := u.folderRepo.GetByID(ctx, folderID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	images, err := u.imageRepo.ListByFolder(ctx, userID, folderID, nil, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("list folder images: %w", err)
	}

	thumbnails := make([]FolderImageThumbnail, 0, len(images))
	for _, img := range images {
		if img.ThumbnailPath == nil {
			continue
		}
		url, err := u.store.GenerateDeterministicPresignedGetURL(ctx, *img.ThumbnailPath)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("generate presigned thumbnail url: %w", err)
		}
		thumbnails = append(thumbnails, FolderImageThumbnail{
			ImageID:      img.ID,
			ThumbnailURL: url,
		})
	}

	return thumbnails, nil
}

// ensure folderImagesUsecase satisfies the interface at compile time.
var _ FolderImagesUsecase = (*folderImagesUsecase)(nil)

