package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/storage"
	"github.com/devi/bookleaf/internal/thumbnail"
	"github.com/google/uuid"
)

var (
	ErrInvalidImageTitle = errors.New("image title is required")
	ErrInvalidMIMEType   = errors.New("mime type is required")
)

const (
	uploadURLTTL    = 15 * time.Minute
	presignedGetTTL = 24 * time.Hour
)

type UploadInitResult struct {
	Image     *domain.Image
	UploadURL string
}

type ImageDetail struct {
	Image    *domain.Image
	ImageURL string
}

type ImageUsecase interface {
	InitiateUpload(ctx context.Context, userID, title, mimeType string, sourceURL *string, folderID *uuid.UUID) (*UploadInitResult, error)
	CompleteUpload(ctx context.Context, id uuid.UUID, userID string) error
	ListImages(ctx context.Context, userID string, folderID *uuid.UUID) ([]*domain.Image, error)
	GetImage(ctx context.Context, id uuid.UUID, userID string) (*ImageDetail, error)
	SoftDelete(ctx context.Context, id uuid.UUID, userID string) error
	ListTrashed(ctx context.Context, userID string) ([]*domain.Image, error)
	Restore(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)
}

type imageUsecase struct {
	imageRepo  ImageRepository
	store      storage.StorageService
	thumbnails thumbnail.ThumbnailService
}

func NewImageUsecase(imageRepo ImageRepository, store storage.StorageService, thumbnails thumbnail.ThumbnailService) ImageUsecase {
	return &imageUsecase{
		imageRepo:  imageRepo,
		store:      store,
		thumbnails: thumbnails,
	}
}

func (u *imageUsecase) InitiateUpload(ctx context.Context, userID, title, mimeType string, sourceURL *string, folderID *uuid.UUID) (*UploadInitResult, error) {
	if strings.TrimSpace(title) == "" {
		return nil, ErrInvalidImageTitle
	}
	if strings.TrimSpace(mimeType) == "" {
		return nil, ErrInvalidMIMEType
	}

	id := uuid.New()
	r2Path := fmt.Sprintf("users/%s/images/%s%s", userID, id.String(), storage.MimeTypeToExt(mimeType))

	created, err := u.imageRepo.Create(ctx, &domain.Image{
		ID:        id,
		UserID:    userID,
		Title:     title,
		MIMEType:  mimeType,
		SourceURL: sourceURL,
		FolderID:  folderID,
		R2Path:    r2Path,
	})
	if err != nil {
		return nil, fmt.Errorf("create image record: %w", err)
	}

	uploadURL, err := u.store.GeneratePresignedPutURL(ctx, r2Path, mimeType, uploadURLTTL)
	if err != nil {
		return nil, fmt.Errorf("generate upload url: %w", err)
	}

	return &UploadInitResult{Image: created, UploadURL: uploadURL}, nil
}

func (u *imageUsecase) CompleteUpload(ctx context.Context, id uuid.UUID, userID string) error {
	image, err := u.imageRepo.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}

	go u.generateThumbnail(image)
	return nil
}

func (u *imageUsecase) generateThumbnail(image *domain.Image) {
	ctx := context.Background()

	src, err := u.store.GetObject(ctx, image.R2Path)
	if err != nil {
		log.Printf("thumbnail: get object %s: %v", image.R2Path, err)
		return
	}
	defer src.Close()

	thumb, err := u.thumbnails.Generate(ctx, src)
	if err != nil {
		log.Printf("thumbnail: generate for image %s: %v", image.ID, err)
		return
	}

	thumbnailKey := fmt.Sprintf("users/%s/thumbnails/%s.jpg", image.UserID, image.ID.String())

	if err := u.store.PutObject(ctx, thumbnailKey, thumb, "image/jpeg"); err != nil {
		log.Printf("thumbnail: put object %s: %v", thumbnailKey, err)
		return
	}

	if err := u.imageRepo.UpdateThumbnailPath(ctx, image.ID, thumbnailKey); err != nil {
		log.Printf("thumbnail: update thumbnail_path for image %s: %v", image.ID, err)
	}
}

func (u *imageUsecase) ListImages(ctx context.Context, userID string, folderID *uuid.UUID) ([]*domain.Image, error) {
	return u.imageRepo.List(ctx, userID, folderID)
}

func (u *imageUsecase) GetImage(ctx context.Context, id uuid.UUID, userID string) (*ImageDetail, error) {
	image, err := u.imageRepo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	imageURL, err := u.store.GeneratePresignedGetURL(ctx, image.R2Path, presignedGetTTL)
	if err != nil {
		return nil, fmt.Errorf("generate presigned url: %w", err)
	}

	return &ImageDetail{Image: image, ImageURL: imageURL}, nil
}

func (u *imageUsecase) SoftDelete(ctx context.Context, id uuid.UUID, userID string) error {
	return u.imageRepo.SoftDelete(ctx, id, userID)
}

func (u *imageUsecase) ListTrashed(ctx context.Context, userID string) ([]*domain.Image, error) {
	return u.imageRepo.ListTrashed(ctx, userID)
}

func (u *imageUsecase) Restore(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error) {
	if _, err := u.imageRepo.GetDeletedByID(ctx, id, userID); err != nil {
		return nil, err
	}

	if err := u.imageRepo.Restore(ctx, id, userID); err != nil {
		return nil, err
	}

	return u.imageRepo.GetByID(ctx, id, userID)
}

var _ ImageUsecase = (*imageUsecase)(nil)
