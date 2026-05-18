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
	"github.com/devi/bookleaf/internal/observability"
	"github.com/devi/bookleaf/internal/storage"
	"github.com/devi/bookleaf/internal/thumbnail"
	"github.com/devi/bookleaf/internal/vision"
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

const (
	uploadURLTTL    = 15 * time.Minute
	presignedGetTTL = 24 * time.Hour
)

type UploadInitResult struct {
	Image     *domain.Image
	UploadURL string
}

type UpdateImageParams struct {
	Title       *string
	FolderID    **uuid.UUID
	Description *string
}

type ImageDetail struct {
	Image    *domain.Image
	ImageURL string
}

type CompleteUploadResult struct {
	ImageID             uuid.UUID
	SuggestedFolderName *string
	Warning             string
}

type ImageUsecase interface {
	InitiateUpload(ctx context.Context, userID, title, mimeType string, sourceURL *string, folderID *uuid.UUID, description *string) (*UploadInitResult, error)
	CompleteUpload(ctx context.Context, id uuid.UUID, userID string) (*CompleteUploadResult, error)
	AcceptSuggestion(ctx context.Context, imageID uuid.UUID, userID string, suggestedFolderName string) error
	ListImages(ctx context.Context, userID string, params ListImagesParams) (*ListImagesResult, error)
	GetImage(ctx context.Context, id uuid.UUID, userID string) (*ImageDetail, error)
	UpdateImage(ctx context.Context, id uuid.UUID, userID string, params UpdateImageParams) (*domain.Image, error)
	SoftDelete(ctx context.Context, id uuid.UUID, userID string) error
	ListTrashed(ctx context.Context, userID string, params ListTrashedParams) (*ListTrashedResult, error)
	Restore(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)
	CleanupStaleUploads(ctx context.Context, threshold time.Duration) error
	PurgeExpiredTrash(ctx context.Context, threshold time.Duration) error
}

type imageUsecase struct {
	imageRepo         ImageRepository
	store             storage.StorageService
	thumbnails        thumbnail.ThumbnailService
	visionService     vision.VisionService
	folderRepo        FolderRepository
	userRepo          UserRepository
	tel               *observability.Telemetry
	uploadCount       metric.Int64Counter
	thumbnailDuration metric.Float64Histogram
	thumbnailCount    metric.Int64Counter
}

func NewImageUsecase(
	imageRepo ImageRepository,
	store storage.StorageService,
	thumbnails thumbnail.ThumbnailService,
	visionService vision.VisionService,
	folderRepo FolderRepository,
	userRepo UserRepository,
	tel *observability.Telemetry,
) ImageUsecase {
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

	return &imageUsecase{
		imageRepo:         imageRepo,
		store:             store,
		thumbnails:        thumbnails,
		visionService:     visionService,
		folderRepo:        folderRepo,
		userRepo:          userRepo,
		tel:               tel,
		uploadCount:       uploadCount,
		thumbnailDuration: thumbnailDuration,
		thumbnailCount:    thumbnailCount,
	}
}

func (u *imageUsecase) InitiateUpload(ctx context.Context, userID, title, mimeType string, sourceURL *string, folderID *uuid.UUID, description *string) (*UploadInitResult, error) {
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

	created, err := u.imageRepo.Create(ctx, &domain.Image{
		ID:          id,
		UserID:      userID,
		Title:       title,
		Description: description,
		MIMEType:    mimeType,
		SourceURL:   sourceURL,
		FolderID:    folderID,
		R2Path:      r2Path,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("create image record: %w", err)
	}

	observability.LoggerFromContext(ctx, u.tel.Logger).Info("upload initiated",
		zap.String("event", "r2.upload.started"),
		zap.String("image_id", created.ID.String()),
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

	return &UploadInitResult{Image: created, UploadURL: uploadURL}, nil
}

func (u *imageUsecase) CompleteUpload(ctx context.Context, id uuid.UUID, userID string) (*CompleteUploadResult, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.CompleteUpload")
	defer span.End()

	start := time.Now()
	result := &CompleteUploadResult{ImageID: id}

	image, err := u.imageRepo.GetByID(ctx, id, userID)
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

	thumbnailBytes, width, height, fileSize, err := u.prepareThumbnail(ctx, image)
	if err != nil {
		result.Warning = "thumbnail generation failed"
		return result, nil
	}

	updateFields := map[string]any{
		"file_size":   fileSize,
		"is_uploaded": true,
		"width":       nil,
		"height":      nil,
	}
	if width > 0 {
		updateFields["width"] = width
	}
	if height > 0 {
		updateFields["height"] = height
	}
	if _, err := u.imageRepo.Update(ctx, id, userID, updateFields); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	thumbnailKey := fmt.Sprintf("users/%s/thumbnails/%s.jpg", image.UserID, image.ID.String())
	go u.uploadThumbnail(image, thumbnailKey, thumbnailBytes)

	result.SuggestedFolderName, result.Warning = u.runVisionFlow(ctx, id, userID, thumbnailBytes)
	return result, nil
}

func (u *imageUsecase) prepareThumbnail(ctx context.Context, img *domain.Image) ([]byte, int, int, int64, error) {
	logger := observability.LoggerFromContext(ctx, u.tel.Logger).With(
		zap.String("image_id", img.ID.String()),
		zap.String("user_id", img.UserID),
	)

	src, err := u.store.GetObject(ctx, img.R2Path)
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

func (u *imageUsecase) runVisionFlow(ctx context.Context, imageID uuid.UUID, userID string, thumbnailBytes []byte) (suggestion *string, warning string) {
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

func (u *imageUsecase) AcceptSuggestion(ctx context.Context, imageID uuid.UUID, userID string, suggestedFolderName string) error {
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

	if _, err := u.imageRepo.Update(ctx, imageID, userID, map[string]any{"folder_id": folder.ID}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (u *imageUsecase) uploadThumbnail(image *domain.Image, thumbnailKey string, thumbnailBytes []byte) {
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

func (u *imageUsecase) ListImages(ctx context.Context, userID string, params ListImagesParams) (*ListImagesResult, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.ListImages")
	defer span.End()

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	} else if limit > 200 {
		limit = 200
	}

	images, err := u.imageRepo.List(ctx, userID, params.FolderID, params.Unfiled, params.Cursor, limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var nextCursor *ImageCursor
	if len(images) > limit {
		images = images[:limit]
		last := images[limit-1]
		nextCursor = &ImageCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}

	return &ListImagesResult{Images: images, NextCursor: nextCursor}, nil
}

func (u *imageUsecase) GetImage(ctx context.Context, id uuid.UUID, userID string) (*ImageDetail, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.GetImage")
	defer span.End()

	image, err := u.imageRepo.GetByID(ctx, id, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	imageURL, err := u.store.GeneratePresignedGetURL(ctx, image.R2Path, presignedGetTTL)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("generate presigned url: %w", err)
	}

	return &ImageDetail{Image: image, ImageURL: imageURL}, nil
}

func (u *imageUsecase) SoftDelete(ctx context.Context, id uuid.UUID, userID string) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.SoftDelete")
	defer span.End()

	if err := u.imageRepo.SoftDelete(ctx, id, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	observability.LoggerFromContext(ctx, u.tel.Logger).Info("image mutated",
		zap.String("event", "image.mutated"),
		zap.String("image_id", id.String()),
		zap.String("user_id", userID),
		zap.String("operation", "trashed"),
	)
	return nil
}

func (u *imageUsecase) ListTrashed(ctx context.Context, userID string, params ListTrashedParams) (*ListTrashedResult, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.ListTrashed")
	defer span.End()

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	} else if limit > 200 {
		limit = 200
	}

	images, err := u.imageRepo.ListTrashed(ctx, userID, params.Cursor, limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var nextCursor *ImageCursor
	if len(images) > limit {
		images = images[:limit]
		last := images[limit-1]
		nextCursor = &ImageCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}

	return &ListTrashedResult{Images: images, NextCursor: nextCursor}, nil
}

func (u *imageUsecase) Restore(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.Restore")
	defer span.End()

	if _, err := u.imageRepo.GetDeletedByID(ctx, id, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if err := u.imageRepo.Restore(ctx, id, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	image, err := u.imageRepo.GetByID(ctx, id, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return image, nil
}

func (u *imageUsecase) UpdateImage(ctx context.Context, id uuid.UUID, userID string, params UpdateImageParams) (*domain.Image, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.UpdateImage")
	defer span.End()

	existing, err := u.imageRepo.GetByID(ctx, id, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	fields := make(map[string]any)
	if params.Title != nil {
		fields["title"] = *params.Title
	}
	if params.FolderID != nil {
		fields["folder_id"] = *params.FolderID
	}
	if params.Description != nil {
		fields["description"] = *params.Description
	}

	updated, err := u.imageRepo.Update(ctx, id, userID, fields)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if params.FolderID != nil {
		newFolderID := *params.FolderID
		oldFolderID := existing.FolderID
		folderChanged := (newFolderID == nil) != (oldFolderID == nil) ||
			(newFolderID != nil && oldFolderID != nil && *newFolderID != *oldFolderID)
		if folderChanged {
			var folderIDField interface{} = nil
			if newFolderID != nil {
				folderIDField = newFolderID.String()
			}
			observability.LoggerFromContext(ctx, u.tel.Logger).Info("image mutated",
				zap.String("event", "image.mutated"),
				zap.String("image_id", id.String()),
				zap.String("user_id", userID),
				zap.String("operation", "moved_to_folder"),
				zap.Any("folder_id", folderIDField),
			)
		}
	}

	return updated, nil
}

func (u *imageUsecase) CleanupStaleUploads(ctx context.Context, threshold time.Duration) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.CleanupStaleUploads")
	defer span.End()

	logger := observability.LoggerFromContext(ctx, u.tel.Logger)

	stale, err := u.imageRepo.ListStaleUploads(ctx, time.Now().Add(-threshold))
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
		if err := u.imageRepo.HardDelete(ctx, img.ID, img.UserID); err != nil {
			logger.Warn("failed to hard delete stale image record",
				zap.String("event", "r2.stale.hard_delete_failed"),
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

func (u *imageUsecase) PurgeExpiredTrash(ctx context.Context, threshold time.Duration) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.PurgeExpiredTrash")
	defer span.End()

	logger := observability.LoggerFromContext(ctx, u.tel.Logger)

	expired, err := u.imageRepo.ListExpiredTrash(ctx, time.Now().Add(-threshold))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("list expired trash: %w", err)
	}

	for _, img := range expired {
		if err := u.store.DeleteObject(ctx, img.R2Path); err != nil {
			logger.Warn("failed to delete R2 object for expired trash",
				zap.String("event", "r2.trash.delete_failed"),
				zap.String("image_id", img.ID.String()),
				zap.String("r2_path", img.R2Path),
				zap.Error(err),
			)
		}
		if img.ThumbnailPath != nil {
			if err := u.store.DeleteObject(ctx, *img.ThumbnailPath); err != nil {
				logger.Warn("failed to delete thumbnail for expired trash",
					zap.String("event", "r2.trash.thumbnail_delete_failed"),
					zap.String("image_id", img.ID.String()),
					zap.String("thumbnail_path", *img.ThumbnailPath),
					zap.Error(err),
				)
			}
		}
		if err := u.imageRepo.HardDelete(ctx, img.ID, img.UserID); err != nil {
			logger.Warn("failed to hard delete expired trash record",
				zap.String("event", "r2.trash.hard_delete_failed"),
				zap.String("image_id", img.ID.String()),
				zap.Error(err),
			)
		}
	}

	if len(expired) > 0 {
		logger.Info("trash purge complete",
			zap.String("event", "r2.trash.purge_complete"),
			zap.Int("purged", len(expired)),
		)
	}

	return nil
}

var _ ImageUsecase = (*imageUsecase)(nil)
