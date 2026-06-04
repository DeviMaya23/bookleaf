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

type ThumbnailService interface {
	Generate(ctx context.Context, src io.Reader) (io.Reader, error)
}

type VisionService interface {
	AnnotateImage(ctx context.Context, imageBytes []byte) ([]domain.Label, error)
}

type UploadInitResult struct {
	ID        uuid.UUID
	UploadURL string
	R2Path    string
}

type CompleteUploadResult struct {
	ImageID             uuid.UUID
	SuggestedFolderName *string
	Warning             string
}

type imageUploadUsecase struct {
	imageRepo         UploadImageRepository
	pendingUploadRepo PendingUploadRepository
	folderRepo        ImageFolderRepository
	userRepo          UserRepository
	store             StorageService
	thumbnails        ThumbnailService
	visionService     VisionService
	tel               *observability.Telemetry
	uploadCount       metric.Int64Counter
	thumbnailDuration metric.Float64Histogram
	thumbnailCount    metric.Int64Counter
}

func NewImageUploadUsecase(
	imageRepo UploadImageRepository,
	pendingUploadRepo PendingUploadRepository,
	folderRepo ImageFolderRepository,
	userRepo UserRepository,
	store StorageService,
	thumbnails ThumbnailService,
	visionService VisionService,
	tel *observability.Telemetry,
) *imageUploadUsecase {
	uploadCount, _ := tel.Meter.Int64Counter(
		"r2.upload.count",
		metric.WithDescription("Total number of upload completion requests"),
	)
	thumbnailDuration, _ := tel.Meter.Float64Histogram(
		"r2.thumbnail.duration",
		metric.WithUnit("ms"),
		metric.WithDescription("Duration of thumbnail generation in milliseconds"),
	)
	thumbnailCount, _ := tel.Meter.Int64Counter(
		"r2.thumbnail.count",
		metric.WithDescription("Total number of thumbnail generation attempts"),
	)

	return &imageUploadUsecase{
		imageRepo:         imageRepo,
		pendingUploadRepo: pendingUploadRepo,
		folderRepo:        folderRepo,
		userRepo:          userRepo,
		store:             store,
		thumbnails:        thumbnails,
		visionService:     visionService,
		tel:               tel,
		uploadCount:       uploadCount,
		thumbnailDuration: thumbnailDuration,
		thumbnailCount:    thumbnailCount,
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

	return &UploadInitResult{ID: pending.ID, UploadURL: uploadURL, R2Path: r2Path}, nil
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

	thumbnailBytes, width, height, fileSize, err := u.prepareThumbnail(ctx, pending.ID, pending.UserID, pending.R2Path)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("prepare thumbnail: %w", err)
	}

	fileSizeValue := fileSize
	image := &domain.Image{
		ID:          pending.ID,
		UserID:      pending.UserID,
		Title:       pending.Title,
		Description: pending.Description,
		SourceURL:   pending.SourceURL,
		R2Path:      pending.R2Path,
		MIMEType:    pending.MIMEType,
		FileSize:    &fileSizeValue,
	}
	if width > 0 {
		widthValue := width
		image.Width = &widthValue
	}
	if height > 0 {
		heightValue := height
		image.Height = &heightValue
	}

	var created *domain.Image
	if err := u.pendingUploadRepo.Transaction(ctx, func(pendingRepo PendingUploadRepository, imageRepo ImageRepository) error {
		createdImage, createErr := imageRepo.Create(ctx, image)
		if createErr != nil {
			return createErr
		}
		created = createdImage
		if pending.FolderID != nil {
			if setErr := imageRepo.SetImageFolder(ctx, createdImage.ID, pending.FolderID); setErr != nil {
				return setErr
			}
		}
		if deleteErr := pendingRepo.Delete(ctx, pending.ID); deleteErr != nil {
			return deleteErr
		}
		return nil
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	thumbnailKey := fmt.Sprintf("users/%s/thumbnails/%s.jpg", pending.UserID, pending.ID.String())
	if created == nil {
		created = image
	}
	go u.uploadThumbnail(created, thumbnailKey, thumbnailBytes)

	result.SuggestedFolderName, result.Warning = u.runVisionFlow(ctx, id, userID, thumbnailBytes)
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

func (u *imageUploadUsecase) prepareThumbnail(ctx context.Context, imageID uuid.UUID, userID, r2Path string) ([]byte, int, int, int64, error) {
	logger := observability.LoggerFromContext(ctx, u.tel.Logger).With(
		zap.String("image_id", imageID.String()),
		zap.String("user_id", userID),
	)

	src, err := u.store.GetObject(ctx, r2Path)
	if err != nil {
		logger.Error("prepare thumbnail failed",
			zap.String("event", "thumbnail.prepare.failed"),
			zap.Error(err),
		)
		return nil, 0, 0, 0, err
	}
	defer src.Close()

	rawBytes, err := io.ReadAll(src)
	if err != nil {
		logger.Error("prepare thumbnail failed",
			zap.String("event", "thumbnail.prepare.failed"),
			zap.Error(err),
		)
		return nil, 0, 0, 0, err
	}

	width, height := 0, 0
	if cfg, _, decodeErr := stdimage.DecodeConfig(bytes.NewReader(rawBytes)); decodeErr != nil {
		logger.Warn("prepare thumbnail metadata decode failed",
			zap.String("event", "thumbnail.metadata.decode_failed"),
			zap.Error(decodeErr),
		)
	} else {
		width = cfg.Width
		height = cfg.Height
	}

	thumb, err := u.thumbnails.Generate(ctx, bytes.NewReader(rawBytes))
	if err != nil {
		logger.Error("prepare thumbnail failed",
			zap.String("event", "thumbnail.prepare.failed"),
			zap.Error(err),
		)
		return nil, 0, 0, 0, err
	}

	thumbnailBytes, err := io.ReadAll(thumb)
	if err != nil {
		logger.Error("prepare thumbnail failed",
			zap.String("event", "thumbnail.prepare.failed"),
			zap.Error(err),
		)
		return nil, 0, 0, 0, err
	}

	return thumbnailBytes, width, height, int64(len(rawBytes)), nil
}

func (u *imageUploadUsecase) runVisionFlow(ctx context.Context, imageID uuid.UUID, userID string, thumbnailBytes []byte) (suggestion *string, warning string) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		observability.LoggerFromContext(ctx, u.tel.Logger).Error("vision: failed to fetch user",
			zap.String("event", "vision.user.fetch_failed"),
			zap.String("image_id", imageID.String()),
			zap.Error(err),
		)
		return nil, "ai labelling skipped: could not fetch user"
	}

	if !user.VisionEnabled || u.visionService == nil {
		return nil, ""
	}

	visionCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	labels, err := u.visionService.AnnotateImage(visionCtx, thumbnailBytes)
	if err != nil {
		observability.LoggerFromContext(ctx, u.tel.Logger).Error("vision: annotation failed",
			zap.String("event", "vision.annotation.failed"),
			zap.String("image_id", imageID.String()),
			zap.Error(err),
		)
		return nil, "ai labelling failed"
	}

	if len(labels) == 0 {
		return nil, ""
	}

	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		observability.LoggerFromContext(ctx, u.tel.Logger).Error("vision: failed to marshal labels",
			zap.String("event", "vision.marshal.failed"),
			zap.String("image_id", imageID.String()),
			zap.Error(err),
		)
		return nil, "ai labelling failed"
	}

	if err := u.imageRepo.UpdateAILabels(ctx, imageID, labelsJSON); err != nil {
		observability.LoggerFromContext(ctx, u.tel.Logger).Error("vision: failed to save labels",
			zap.String("event", "vision.labels.save_failed"),
			zap.String("image_id", imageID.String()),
			zap.Error(err),
		)
		return nil, "ai labelling failed"
	}

	topLabel := labels[0]
	return &topLabel.Description, ""
}

func (u *imageUploadUsecase) uploadThumbnail(image *domain.Image, thumbnailKey string, thumbnailBytes []byte) {
	ctx := context.Background()
	logger := u.tel.Logger.With(
		zap.String("image_id", image.ID.String()),
		zap.String("user_id", image.UserID),
	)

	logger.Info("upload thumbnail job started", zap.String("event", "upload.thumbnail.job.started"))
	start := time.Now()

	recordMetrics := func(status string) {
		elapsed := float64(time.Since(start).Milliseconds())
		attrs := metric.WithAttributes(attribute.String("r2.status", status))
		u.thumbnailDuration.Record(ctx, elapsed, attrs)
		u.thumbnailCount.Add(ctx, 1, attrs)
	}

	if err := u.store.PutObject(ctx, thumbnailKey, bytes.NewReader(thumbnailBytes), "image/jpeg"); err != nil {
		logger.Error("upload thumbnail job failed",
			zap.String("event", "upload.thumbnail.job.failed"),
			zap.Error(err),
		)
		recordMetrics("error")
		return
	}

	if err := u.imageRepo.UpdateThumbnailPath(ctx, image.ID, thumbnailKey); err != nil {
		logger.Error("upload thumbnail job failed",
			zap.String("event", "upload.thumbnail.job.failed"),
			zap.Error(err),
		)
		recordMetrics("error")
		return
	}

	logger.Info("upload thumbnail job completed",
		zap.String("event", "upload.thumbnail.job.completed"),
		zap.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)
	recordMetrics("success")
}
