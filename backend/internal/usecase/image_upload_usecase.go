package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	stdimage "image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"strings"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/devi/bookleaf/internal/storage"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

var (
	ErrInvalidImageTitle = errors.New("image title is required")
	ErrInvalidMIMEType   = errors.New("mime type is required")
)

type UploadImageRepository interface {
	Create(ctx context.Context, image *domain.Image) (*domain.Image, error)
	GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)
	SetImageFolder(ctx context.Context, imageID uuid.UUID, folderID *uuid.UUID) error
	UpdateAILabels(ctx context.Context, id uuid.UUID, labels json.RawMessage) error
	UpdateThumbnailPath(ctx context.Context, id uuid.UUID, thumbnailPath string) error
}

const uploadURLTTL = 15 * time.Minute

type VisionService interface {
	AnnotateImage(ctx context.Context, imageBytes []byte) ([]domain.Label, error)
}

type UploadInitResult struct {
	ID                 uuid.UUID
	UploadURL          string
	R2Path             string
	ThumbnailUploadURL string
	ThumbnailKey       string
}

type CompleteUploadResult struct {
	ImageID uuid.UUID
}

type imageUploadUsecase struct {
	imageRepo         UploadImageRepository
	pendingUploadRepo PendingUploadRepository
	folderRepo        ImageFolderRepository
	userRepo          UserRepository
	store             StorageService
	visionService     VisionService
	enqueuer          JobEnqueuer
	tel               *observability.Telemetry
	uploadCount       metric.Int64Counter
}

func NewImageUploadUsecase(
	imageRepo UploadImageRepository,
	pendingUploadRepo PendingUploadRepository,
	folderRepo ImageFolderRepository,
	userRepo UserRepository,
	store StorageService,
	visionService VisionService,
	enqueuer JobEnqueuer,
	tel *observability.Telemetry,
) *imageUploadUsecase {
	uploadCount, _ := tel.Meter.Int64Counter(
		"r2.upload.count",
		metric.WithDescription("Total number of upload completion requests"),
	)

	return &imageUploadUsecase{
		imageRepo:         imageRepo,
		pendingUploadRepo: pendingUploadRepo,
		folderRepo:        folderRepo,
		userRepo:          userRepo,
		store:             store,
		visionService:     visionService,
		enqueuer:          enqueuer,
		tel:               tel,
		uploadCount:       uploadCount,
	}
}

func (u *imageUploadUsecase) InitiateUpload(ctx context.Context, userID, title, mimeType string, sourceURL *string, folderID *uuid.UUID, description *string) (*UploadInitResult, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.InitiateUpload")
	defer span.End()

	if strings.TrimSpace(title) == "" {
		return nil, ErrInvalidImageTitle
	}
	if strings.TrimSpace(mimeType) == "" {
		return nil, ErrInvalidMIMEType
	}

	id := uuid.New()
	r2Path := fmt.Sprintf("users/%s/images/%s%s", userID, id.String(), storage.MimeTypeToExt(mimeType))

	if folderID != nil {
		if _, err := u.folderRepo.GetByID(ctx, *folderID, userID); err != nil {
			observability.LoggerFromContext(ctx, u.tel.Logger).Info("initiate upload folder fallback applied",
				zap.String("event", "image.initiate_upload.folder_fallback"),
				zap.String("image_id", id.String()),
				zap.String("user_id", userID),
				zap.String("requested_folder_id", folderID.String()),
				zap.Error(err),
			)
			folderID = nil
		}
	}

	pending, err := u.pendingUploadRepo.Create(ctx, &domain.PendingUpload{
		ID:          id,
		UserID:      userID,
		Title:       title,
		Description: description,
		MIMEType:    mimeType,
		SourceURL:   sourceURL,
		R2Path:      r2Path,
		FolderID:    folderID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("create pending upload record: %w", err)
	}

	observability.LoggerFromContext(ctx, u.tel.Logger).Info("upload initiated",
		zap.String("event", "r2.upload.started"),
		zap.String("image_id", pending.ID.String()),
		zap.String("user_id", userID),
		zap.String("mime_type", mimeType),
		zap.String("r2_key", r2Path),
	)

	uploadURL, err := u.store.GeneratePresignedPutURL(ctx, r2Path, mimeType, uploadURLTTL)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("generate upload url: %w", err)
	}

	thumbnailKey := fmt.Sprintf("users/%s/thumbnails/%s.jpg", userID, id.String())
	thumbnailUploadURL, err := u.store.GeneratePresignedPutURL(ctx, thumbnailKey, "image/jpeg", uploadURLTTL)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("generate thumbnail upload url: %w", err)
	}

	return &UploadInitResult{
		ID:                 pending.ID,
		UploadURL:          uploadURL,
		R2Path:             r2Path,
		ThumbnailUploadURL: thumbnailUploadURL,
		ThumbnailKey:       thumbnailKey,
	}, nil
}

func (u *imageUploadUsecase) CompleteUpload(ctx context.Context, id uuid.UUID, userID string) (*CompleteUploadResult, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.CompleteUpload")
	defer span.End()

	start := time.Now()
	result := &CompleteUploadResult{ImageID: id}

	pending, err := u.pendingUploadRepo.GetByID(ctx, id, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		u.uploadCount.Add(ctx, 1, metric.WithAttributes(attribute.String("r2.status", "error")))
		return nil, err
	}

	u.uploadCount.Add(ctx, 1, metric.WithAttributes(attribute.String("r2.status", "success")))
	observability.LoggerFromContext(ctx, u.tel.Logger).Info("upload completed",
		zap.String("event", "r2.upload.completed"),
		zap.String("image_id", id.String()),
		zap.String("user_id", userID),
		zap.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)

	width, height, fileSize, err := u.extractImageMetadata(ctx, pending.R2Path)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("extract image metadata: %w", err)
	}

	thumbnailKey := fmt.Sprintf("users/%s/thumbnails/%s.jpg", pending.UserID, pending.ID.String())

	img := &domain.Image{
		ID:            pending.ID,
		UserID:        pending.UserID,
		Title:         pending.Title,
		Description:   pending.Description,
		SourceURL:     pending.SourceURL,
		R2Path:        pending.R2Path,
		MIMEType:      pending.MIMEType,
		FileSize:      &fileSize,
		ThumbnailPath: &thumbnailKey,
	}
	if width > 0 {
		widthValue := width
		img.Width = &widthValue
	}
	if height > 0 {
		heightValue := height
		img.Height = &heightValue
	}

	if err := u.pendingUploadRepo.Transaction(ctx, func(pendingRepo PendingUploadRepository, imageRepo ImageRepository) error {
		createdImage, createErr := imageRepo.Create(ctx, img)
		if createErr != nil {
			return createErr
		}
		if pending.FolderID != nil {
			if setErr := imageRepo.SetImageFolder(ctx, createdImage.ID, pending.FolderID); setErr != nil {
				return setErr
			}
		}
		return pendingRepo.Delete(ctx, pending.ID)
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if err := u.enqueuer.Insert(ctx, VisionArgs{
		ImageID: pending.ID,
		UserID:  pending.UserID,
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("enqueue vision labelling: %w", err)
	}

	return result, nil
}

func (u *imageUploadUsecase) AcceptSuggestion(ctx context.Context, imageID uuid.UUID, userID string, suggestedFolderName string) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.AcceptSuggestion")
	defer span.End()

	if _, err := u.imageRepo.GetByID(ctx, imageID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	suggestedFolderName = strings.TrimSpace(suggestedFolderName)

	folder, err := u.folderRepo.FindByName(ctx, userID, suggestedFolderName)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if folder == nil {
		folder, err = u.folderRepo.Create(ctx, &domain.Folder{
			UserID: userID,
			Name:   suggestedFolderName,
		})
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
	}

	if err := u.imageRepo.SetImageFolder(ctx, imageID, &folder.ID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (u *imageUploadUsecase) CleanupStaleUploads(ctx context.Context, threshold time.Duration) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.CleanupStaleUploads")
	defer span.End()

	logger := observability.LoggerFromContext(ctx, u.tel.Logger)

	stale, err := u.pendingUploadRepo.ListStale(ctx, time.Now().Add(-threshold))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("list stale uploads: %w", err)
	}

	for _, img := range stale {
		if err := u.store.DeleteObject(ctx, img.R2Path); err != nil {
			logger.Warn("failed to delete stale R2 object",
				zap.String("event", "r2.stale.delete_failed"),
				zap.String("image_id", img.ID.String()),
				zap.String("r2_path", img.R2Path),
				zap.Error(err),
			)
		}
		if err := u.pendingUploadRepo.Delete(ctx, img.ID); err != nil {
			logger.Warn("failed to delete stale pending upload record",
				zap.String("event", "r2.stale.pending_delete_failed"),
				zap.String("image_id", img.ID.String()),
				zap.Error(err),
			)
		}
	}

	if len(stale) > 0 {
		logger.Info("stale upload cleanup complete",
			zap.String("event", "r2.stale.cleanup_complete"),
			zap.Int("cleaned", len(stale)),
		)
	}

	return nil
}

func (u *imageUploadUsecase) extractImageMetadata(ctx context.Context, r2Path string) (width, height int, fileSize int64, err error) {
	src, err := u.store.GetObject(ctx, r2Path)
	if err != nil {
		return 0, 0, 0, err
	}
	defer src.Close()

	rawBytes, err := io.ReadAll(src)
	if err != nil {
		return 0, 0, 0, err
	}

	if cfg, _, decodeErr := stdimage.DecodeConfig(bytes.NewReader(rawBytes)); decodeErr == nil {
		width = cfg.Width
		height = cfg.Height
	}

	return width, height, int64(len(rawBytes)), nil
}

func (u *imageUploadUsecase) ProcessVisionLabelling(ctx context.Context, imageID uuid.UUID, userID string) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("fetch user: %w", err)
	}

	if !user.VisionEnabled || u.visionService == nil {
		return nil
	}

	img, err := u.imageRepo.GetByID(ctx, imageID, userID)
	if err != nil {
		return fmt.Errorf("fetch image: %w", err)
	}

	if img.ThumbnailPath == nil {
		return fmt.Errorf("image has no thumbnail path")
	}

	src, err := u.store.GetObject(ctx, *img.ThumbnailPath)
	if err != nil {
		return fmt.Errorf("fetch thumbnail bytes: %w", err)
	}
	defer src.Close()

	imgBytes, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("read image bytes: %w", err)
	}

	visionCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	labels, err := u.visionService.AnnotateImage(visionCtx, imgBytes)
	if err != nil {
		return fmt.Errorf("annotate image: %w", err)
	}

	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		return fmt.Errorf("marshal labels: %w", err)
	}

	if err := u.imageRepo.UpdateAILabels(ctx, imageID, labelsJSON); err != nil {
		return fmt.Errorf("save labels: %w", err)
	}

	return nil
}
