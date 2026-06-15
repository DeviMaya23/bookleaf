package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
)

const shareTokenBytes = 16

// ShareFolderRepository is the narrow folder lookup ShareUsecase needs,
// satisfied implicitly by the existing folderRepository.
type ShareFolderRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Folder, error)
}

// ShareImageRepository is the narrow ordered image listing ShareUsecase
// needs, satisfied implicitly by the existing imageRepository.
type ShareImageRepository interface {
	ListByFolder(ctx context.Context, userID string, folderID uuid.UUID, sortField *string, direction *string) ([]*domain.Image, error)
}

type SharedImage struct {
	Title        string
	ThumbnailURL *string
	FullResURL   string
	DownloadURL  string
	Width        *int
	Height       *int
}

type SharedFolder struct {
	Name   string
	Notes  *string
	Images []SharedImage
}

type shareUsecase struct {
	folderShareRepo FolderShareRepository
	folderRepo      ShareFolderRepository
	imageRepo       ShareImageRepository
	store           StorageService
	tel             *observability.Telemetry
}

func NewShareUsecase(
	folderShareRepo FolderShareRepository,
	folderRepo ShareFolderRepository,
	imageRepo ShareImageRepository,
	store StorageService,
	tel *observability.Telemetry,
) *shareUsecase {
	return &shareUsecase{
		folderShareRepo: folderShareRepo,
		folderRepo:      folderRepo,
		imageRepo:       imageRepo,
		store:           store,
		tel:             tel,
	}
}

func generateShareToken() (string, error) {
	b := make([]byte, shareTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate share token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (u *shareUsecase) CreateShare(ctx context.Context, folderID uuid.UUID, userID string) (string, bool, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.CreateShare")
	defer span.End()

	if _, err := u.folderRepo.GetByID(ctx, folderID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", false, err
	}

	existing, err := u.folderShareRepo.GetByFolderID(ctx, folderID)
	switch {
	case err == nil:
		return existing.Token, false, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		// no existing share — fall through to create one
	default:
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", false, err
	}

	token, err := generateShareToken()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", false, err
	}

	created, err := u.folderShareRepo.Create(ctx, folderID, token)
	if err != nil {
		if isUniqueConstraintViolation(err) {
			existing, getErr := u.folderShareRepo.GetByFolderID(ctx, folderID)
			if getErr != nil {
				span.RecordError(getErr)
				span.SetStatus(codes.Error, getErr.Error())
				return "", false, getErr
			}
			return existing.Token, false, nil
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", false, fmt.Errorf("create folder share: %w", err)
	}

	return created.Token, true, nil
}

func (u *shareUsecase) GetShare(ctx context.Context, folderID uuid.UUID, userID string) (string, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.GetShare")
	defer span.End()

	if _, err := u.folderRepo.GetByID(ctx, folderID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	share, err := u.folderShareRepo.GetByFolderID(ctx, folderID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	return share.Token, nil
}

func (u *shareUsecase) DeleteShare(ctx context.Context, folderID uuid.UUID, userID string) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.DeleteShare")
	defer span.End()

	if _, err := u.folderRepo.GetByID(ctx, folderID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if err := u.folderShareRepo.DeleteByFolderID(ctx, folderID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("delete folder share: %w", err)
	}

	return nil
}

func (u *shareUsecase) thumbnailURL(ctx context.Context, path *string) *string {
	if path == nil {
		return nil
	}
	url, err := u.store.GeneratePresignedGetURL(ctx, *path, presignedGetTTL)
	if err != nil {
		return nil
	}
	return &url
}

func (u *shareUsecase) GetSharedFolderInfo(ctx context.Context, token string) (*domain.Folder, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.GetSharedFolderInfo")
	defer span.End()

	share, err := u.folderShareRepo.GetByToken(ctx, token)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return &share.Folder, nil
}

func (u *shareUsecase) GetSharedFolder(ctx context.Context, token string) (*SharedFolder, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.GetSharedFolder")
	defer span.End()

	share, err := u.folderShareRepo.GetByToken(ctx, token)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	rawImages, err := u.imageRepo.ListByFolder(ctx, share.Folder.UserID, share.FolderID, nil, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("list shared folder images: %w", err)
	}

	images := make([]SharedImage, len(rawImages))
	for i, img := range rawImages {
		fullResURL, err := u.store.GeneratePresignedGetURL(ctx, img.R2Path, presignedGetTTL)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("generate presigned url: %w", err)
		}
		filename := img.Title + "." + downloadFileExtension(img.MIMEType)
		downloadURL, err := u.store.GeneratePresignedDownloadURL(ctx, img.R2Path, filename, presignedGetTTL)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("generate presigned download url: %w", err)
		}
		images[i] = SharedImage{
			Title:        img.Title,
			ThumbnailURL: u.thumbnailURL(ctx, img.ThumbnailPath),
			FullResURL:   fullResURL,
			DownloadURL:  downloadURL,
			Width:        img.Width,
			Height:       img.Height,
		}
	}

	return &SharedFolder{
		Name:   share.Folder.Name,
		Notes:  share.Folder.Description,
		Images: images,
	}, nil
}
