package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/observability"
	"github.com/devi/bookleaf/internal/storage"
	"github.com/devi/bookleaf/internal/thumbnail"
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
	Title    *string
	FolderID **uuid.UUID
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
	UpdateImage(ctx context.Context, id uuid.UUID, userID string, params UpdateImageParams) (*domain.Image, error)
	SoftDelete(ctx context.Context, id uuid.UUID, userID string) error
	ListTrashed(ctx context.Context, userID string) ([]*domain.Image, error)
	Restore(ctx context.Context, id uuid.UUID, userID string) (*domain.Image, error)
}

type imageUsecase struct {
	imageRepo         ImageRepository
	store             storage.StorageService
	thumbnails        thumbnail.ThumbnailService
	tel               *observability.Telemetry
	uploadCount       metric.Int64Counter
	thumbnailDuration metric.Float64Histogram
	thumbnailCount    metric.Int64Counter
}

func NewImageUsecase(imageRepo ImageRepository, store storage.StorageService, thumbnails thumbnail.ThumbnailService, tel *observability.Telemetry) ImageUsecase {
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
		tel:               tel,
		uploadCount:       uploadCount,
		thumbnailDuration: thumbnailDuration,
		thumbnailCount:    thumbnailCount,
	}
}

func (u *imageUsecase) InitiateUpload(ctx context.Context, userID, title, mimeType string, sourceURL *string, folderID *uuid.UUID) (*UploadInitResult, error) {
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

func (u *imageUsecase) CompleteUpload(ctx context.Context, id uuid.UUID, userID string) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.CompleteUpload")
	defer span.End()

	start := time.Now()

	image, err := u.imageRepo.GetByID(ctx, id, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		u.uploadCount.Add(ctx, 1, metric.WithAttributes(attribute.String("r2.status", "error")))
		return err
	}

	u.uploadCount.Add(ctx, 1, metric.WithAttributes(attribute.String("r2.status", "success")))
	observability.LoggerFromContext(ctx, u.tel.Logger).Info("upload completed",
		zap.String("event", "r2.upload.completed"),
		zap.String("image_id", id.String()),
		zap.String("user_id", userID),
		zap.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)

	go u.generateThumbnail(image)
	return nil
}

func (u *imageUsecase) generateThumbnail(image *domain.Image) {
	ctx := context.Background()
	logger := u.tel.Logger.With(
		zap.String("image_id", image.ID.String()),
		zap.String("user_id", image.UserID),
	)

	logger.Info("thumbnail job started", zap.String("event", "thumbnail.job.started"))
	start := time.Now()

	recordMetrics := func(status string) {
		elapsed := float64(time.Since(start).Milliseconds())
		attrs := metric.WithAttributes(attribute.String("r2.status", status))
		u.thumbnailDuration.Record(ctx, elapsed, attrs)
		u.thumbnailCount.Add(ctx, 1, attrs)
	}

	src, err := u.store.GetObject(ctx, image.R2Path)
	if err != nil {
		logger.Error("thumbnail job failed",
			zap.String("event", "thumbnail.job.failed"),
			zap.Error(err),
		)
		recordMetrics("error")
		return
	}
	defer src.Close()

	thumb, err := u.thumbnails.Generate(ctx, src)
	if err != nil {
		logger.Error("thumbnail job failed",
			zap.String("event", "thumbnail.job.failed"),
			zap.Error(err),
		)
		recordMetrics("error")
		return
	}

	thumbnailKey := fmt.Sprintf("users/%s/thumbnails/%s.jpg", image.UserID, image.ID.String())

	if err := u.store.PutObject(ctx, thumbnailKey, thumb, "image/jpeg"); err != nil {
		logger.Error("thumbnail job failed",
			zap.String("event", "thumbnail.job.failed"),
			zap.Error(err),
		)
		recordMetrics("error")
		return
	}

	if err := u.imageRepo.UpdateThumbnailPath(ctx, image.ID, thumbnailKey); err != nil {
		logger.Error("thumbnail job failed",
			zap.String("event", "thumbnail.job.failed"),
			zap.Error(err),
		)
		recordMetrics("error")
		return
	}

	logger.Info("thumbnail job completed",
		zap.String("event", "thumbnail.job.completed"),
		zap.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)
	recordMetrics("success")
}

func (u *imageUsecase) ListImages(ctx context.Context, userID string, folderID *uuid.UUID) ([]*domain.Image, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.ListImages")
	defer span.End()

	images, err := u.imageRepo.List(ctx, userID, folderID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return images, nil
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

func (u *imageUsecase) ListTrashed(ctx context.Context, userID string) ([]*domain.Image, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.ListTrashed")
	defer span.End()

	images, err := u.imageRepo.ListTrashed(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return images, nil
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

var _ ImageUsecase = (*imageUsecase)(nil)
