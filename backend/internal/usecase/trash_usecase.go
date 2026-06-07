package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type TrashUsecase interface {
	SoftDelete(ctx context.Context, id uuid.UUID, userID string) error
	ListTrashed(ctx context.Context, userID string, params ListTrashedParams) (*ListTrashedResult, error)
	Restore(ctx context.Context, id uuid.UUID, userID string) (*ImageItem, error)
	PurgeExpiredTrash(ctx context.Context, threshold time.Duration) error
	DeleteFromTrash(ctx context.Context, id uuid.UUID, userID string) error
	EmptyTrash(ctx context.Context, userID string) error
	ProcessR2Delete(ctx context.Context, r2Path string, thumbnailPath *string) error
}

type trashUsecase struct {
	imageRepo TrashRepository
	store     StorageService
	enqueuer  JobEnqueuer
	tel       *observability.Telemetry
}

func NewTrashUsecase(
	imageRepo TrashRepository,
	store StorageService,
	enqueuer JobEnqueuer,
	tel *observability.Telemetry,
) *trashUsecase {
	return &trashUsecase{
		imageRepo: imageRepo,
		store:     store,
		enqueuer:  enqueuer,
		tel:       tel,
	}
}

func (u *trashUsecase) thumbnailURL(ctx context.Context, path *string) *string {
	if path == nil {
		return nil
	}
	url, err := u.store.GeneratePresignedGetURL(ctx, *path, presignedGetTTL)
	if err != nil {
		return nil
	}
	return &url
}

func (u *trashUsecase) SoftDelete(ctx context.Context, id uuid.UUID, userID string) error {
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

func (u *trashUsecase) ListTrashed(ctx context.Context, userID string, params ListTrashedParams) (*ListTrashedResult, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.ListTrashed")
	defer span.End()

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	} else if limit > 200 {
		limit = 200
	}

	var name *string
	if params.Name != nil && *params.Name != "" {
		name = params.Name
	}

	rawImages, err := u.imageRepo.ListTrashed(ctx, userID, name, params.Cursor, limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var nextCursor *ImageCursor
	if len(rawImages) > limit {
		rawImages = rawImages[:limit]
		last := rawImages[limit-1]
		deletedAt := last.DeletedAt.Time
		nextCursor = &ImageCursor{DeletedAt: &deletedAt, ID: last.ID}
	}

	items := make([]ImageItem, len(rawImages))
	for i, img := range rawImages {
		items[i] = ImageItem{Image: img, ThumbnailURL: u.thumbnailURL(ctx, img.ThumbnailPath)}
	}

	return &ListTrashedResult{Images: items, NextCursor: nextCursor}, nil
}

func (u *trashUsecase) Restore(ctx context.Context, id uuid.UUID, userID string) (*ImageItem, error) {
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
	return &ImageItem{Image: image, ThumbnailURL: u.thumbnailURL(ctx, image.ThumbnailPath)}, nil
}

func (u *trashUsecase) PurgeExpiredTrash(ctx context.Context, threshold time.Duration) error {
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

func (u *trashUsecase) DeleteFromTrash(ctx context.Context, id uuid.UUID, userID string) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.DeleteFromTrash")
	defer span.End()

	logger := observability.LoggerFromContext(ctx, u.tel.Logger)

	img, err := u.imageRepo.GetDeletedByID(ctx, id, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if err := u.store.DeleteObject(ctx, img.R2Path); err != nil {
		logger.Warn("failed to delete R2 object from trash",
			zap.String("event", "r2.trash.delete_failed"),
			zap.String("image_id", img.ID.String()),
			zap.String("r2_path", img.R2Path),
			zap.Error(err),
		)
	}
	if img.ThumbnailPath != nil {
		if err := u.store.DeleteObject(ctx, *img.ThumbnailPath); err != nil {
			logger.Warn("failed to delete thumbnail from trash",
				zap.String("event", "r2.trash.thumbnail_delete_failed"),
				zap.String("image_id", img.ID.String()),
				zap.String("thumbnail_path", *img.ThumbnailPath),
				zap.Error(err),
			)
		}
	}
	if err := u.imageRepo.HardDelete(ctx, img.ID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("hard delete image: %w", err)
	}

	logger.Info("image mutated",
		zap.String("event", "image.mutated"),
		zap.String("image_id", img.ID.String()),
		zap.String("user_id", userID),
		zap.String("operation", "deleted_from_trash"),
	)
	return nil
}

func (u *trashUsecase) EmptyTrash(ctx context.Context, userID string) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.EmptyTrash")
	defer span.End()

	logger := observability.LoggerFromContext(ctx, u.tel.Logger)

	images, err := u.imageRepo.ListAllTrashed(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("list all trashed: %w", err)
	}

	if len(images) == 0 {
		return nil
	}

	for _, img := range images {
		if err := u.imageRepo.HardDelete(ctx, img.ID, userID); err != nil {
			logger.Warn("failed to hard delete during empty trash",
				zap.String("event", "r2.trash.hard_delete_failed"),
				zap.String("image_id", img.ID.String()),
				zap.Error(err),
			)
		}
	}

	for _, img := range images {
		if err := u.enqueuer.Insert(ctx, R2DeleteArgs{R2Path: img.R2Path, ThumbnailPath: img.ThumbnailPath}); err != nil {
			logger.Warn("failed to enqueue R2 delete job during empty trash",
				zap.String("event", "r2.trash.enqueue_failed"),
				zap.String("image_id", img.ID.String()),
				zap.String("r2_path", img.R2Path),
				zap.Error(err),
			)
		}
	}

	logger.Info("trash emptied",
		zap.String("event", "r2.trash.emptied"),
		zap.String("user_id", userID),
		zap.Int("deleted", len(images)),
	)

	return nil
}

func (u *trashUsecase) ProcessR2Delete(ctx context.Context, r2Path string, thumbnailPath *string) error {
	if err := u.store.DeleteObject(ctx, r2Path); err != nil {
		return fmt.Errorf("delete r2 object: %w", err)
	}
	if thumbnailPath != nil {
		if err := u.store.DeleteObject(ctx, *thumbnailPath); err != nil {
			return fmt.Errorf("delete thumbnail: %w", err)
		}
	}
	return nil
}
