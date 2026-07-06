package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/platform/observability"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

const (
	presignedGetTTL = 24 * time.Hour
	downloadURLTTL  = 5 * time.Minute
)

type UpdateImageParams struct {
	Title       *string
	FolderIDs   *[]uuid.UUID
	Description *string
	SourceURL   **string
	Tags        *[]uuid.UUID
}

type ImageDetail struct {
	Image               *domain.Image
	ImageURL            string
	ThumbnailURL        *string
	SuggestedFolderName *string
}

type ImageItem struct {
	Image          *domain.Image
	ThumbnailURL   *string
	FolderPosition *string
}

type imageUsecase struct {
	imageRepo  ImageRepository
	tagRepo    TagRepository
	folderRepo ImageFolderRepository
	store      StorageService
	tel        *observability.Telemetry
}

func NewImageUsecase(
	imageRepo ImageRepository,
	tagRepo TagRepository,
	folderRepo ImageFolderRepository,
	store StorageService,
	tel *observability.Telemetry,
) *imageUsecase {
	return &imageUsecase{
		imageRepo:  imageRepo,
		tagRepo:    tagRepo,
		folderRepo: folderRepo,
		store:      store,
		tel:        tel,
	}
}

func (u *imageUsecase) thumbnailURL(ctx context.Context, path *string) *string {
	if path == nil {
		return nil
	}
	url, err := u.store.GeneratePresignedGetURL(ctx, *path, presignedGetTTL)
	if err != nil {
		return nil
	}
	return &url
}

func (u *imageUsecase) ListImages(ctx context.Context, userID string, params ListImagesParams) (*ListImagesResult, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.ListImages")
	defer span.End()

	var name *string
	if params.Name != nil && *params.Name != "" {
		name = params.Name
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	} else if limit > 200 {
		limit = 200
	}

	rawImages, err := u.imageRepo.List(ctx, userID, params.Unfiled, params.FolderIDs, params.TagIDs, params.MIMETypes, name, params.SearchLabels, params.Sort, params.Direction, params.Cursor, limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var nextCursor *ImageCursor
	if len(rawImages) > limit {
		rawImages = rawImages[:limit]
		last := rawImages[limit-1]

		dispatch := ResolveSort(params.Sort, params.Direction)
		var title *string
		if dispatch.Column == "title" {
			title = &last.Title
		}

		nextCursor = &ImageCursor{CreatedAt: last.CreatedAt, Title: title, ID: last.ID}
	}

	items := make([]ImageItem, len(rawImages))
	for i, img := range rawImages {
		items[i] = ImageItem{Image: img, ThumbnailURL: u.thumbnailURL(ctx, img.ThumbnailPath)}
	}

	return &ListImagesResult{Images: items, NextCursor: nextCursor}, nil
}

func (u *imageUsecase) ListFolderImages(ctx context.Context, userID string, folderID uuid.UUID, sort *string, direction *string) ([]ImageItem, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.ListFolderImages")
	defer span.End()

	if _, err := u.folderRepo.GetByID(ctx, folderID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	rawImages, err := u.imageRepo.ListByFolder(ctx, userID, folderID, sort, direction)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	items := make([]ImageItem, len(rawImages))
	for i, img := range rawImages {
		var folderPos *string
		for _, f := range img.ImageFolders {
			if f.FolderID == folderID {
				p := f.Position
				folderPos = &p
				break
			}
		}
		items[i] = ImageItem{Image: img, ThumbnailURL: u.thumbnailURL(ctx, img.ThumbnailPath), FolderPosition: folderPos}
	}

	return items, nil
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

	return &ImageDetail{
		Image:               image,
		ImageURL:            imageURL,
		ThumbnailURL:        u.thumbnailURL(ctx, image.ThumbnailPath),
		SuggestedFolderName: suggestedFolderName(image.AILabels),
	}, nil
}

func suggestedFolderName(aiLabels json.RawMessage) *string {
	if len(aiLabels) == 0 {
		return nil
	}
	var labels []domain.Label
	if err := json.Unmarshal(aiLabels, &labels); err != nil || len(labels) == 0 {
		return nil
	}
	name := labels[0].Description
	return &name
}

func (u *imageUsecase) DownloadImage(ctx context.Context, id uuid.UUID, userID string) (string, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.DownloadImage")
	defer span.End()

	image, err := u.imageRepo.GetByID(ctx, id, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	filename := image.Title + "." + downloadFileExtension(image.MIMEType)
	downloadURL, err := u.store.GeneratePresignedDownloadURL(ctx, image.R2Path, filename, downloadURLTTL)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", fmt.Errorf("generate presigned download url: %w", err)
	}

	return downloadURL, nil
}

func (u *imageUsecase) UpdateImage(ctx context.Context, id uuid.UUID, userID string, params UpdateImageParams) (*ImageItem, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.UpdateImage")
	defer span.End()

	if _, err := u.imageRepo.GetByID(ctx, id, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	fields := make(map[string]any)
	if params.Title != nil {
		fields["title"] = *params.Title
	}
	if params.Description != nil {
		fields["description"] = *params.Description
	}
	if params.SourceURL != nil {
		fields["source_url"] = *params.SourceURL
	}

	updated, err := u.imageRepo.Update(ctx, id, userID, fields)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if params.Tags != nil {
		if err := u.tagRepo.ReplaceImageTags(ctx, id, *params.Tags); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, fmt.Errorf("replace image tags: %w", err)
		}
		updated, err = u.imageRepo.GetByID(ctx, id, userID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
	}

	if params.FolderIDs != nil {
		if err := u.imageRepo.SyncImageFolders(ctx, id, *params.FolderIDs); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		updated, err = u.imageRepo.GetByID(ctx, id, userID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}

		folderIDs := make([]string, 0, len(*params.FolderIDs))
		for _, folderID := range *params.FolderIDs {
			folderIDs = append(folderIDs, folderID.String())
		}
		observability.LoggerFromContext(ctx, u.tel.Logger).Info("image mutated",
			zap.String("event", "image.mutated"),
			zap.String("image_id", id.String()),
			zap.String("user_id", userID),
			zap.String("operation", "synced_folders"),
			zap.Strings("folder_ids", folderIDs),
		)
	}

	return &ImageItem{Image: updated, ThumbnailURL: u.thumbnailURL(ctx, updated.ThumbnailPath)}, nil
}

func (u *imageUsecase) MoveImageFolder(ctx context.Context, imageID uuid.UUID, userID string, fromFolderID *uuid.UUID, toFolderID *uuid.UUID) (*ImageItem, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.MoveImageFolder")
	defer span.End()

	if _, err := u.imageRepo.GetByID(ctx, imageID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if (fromFolderID == nil && toFolderID == nil) ||
		(fromFolderID != nil && toFolderID != nil && *fromFolderID == *toFolderID) {
		img, err := u.imageRepo.GetByID(ctx, imageID, userID)
		if err != nil {
			return nil, err
		}
		return &ImageItem{Image: img, ThumbnailURL: u.thumbnailURL(ctx, img.ThumbnailPath)}, nil
	}

	if err := u.imageRepo.MoveImageFolder(ctx, imageID, fromFolderID, toFolderID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	img, err := u.imageRepo.GetByID(ctx, imageID, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return &ImageItem{Image: img, ThumbnailURL: u.thumbnailURL(ctx, img.ThumbnailPath)}, nil
}

func (u *imageUsecase) UpdateImagePosition(ctx context.Context, imageID uuid.UUID, userID string, folderID uuid.UUID, position string) error {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.UpdateImagePosition")
	defer span.End()

	if _, err := u.imageRepo.GetByID(ctx, imageID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if err := u.imageRepo.UpdateImageFolderPosition(ctx, imageID, folderID, position); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func (u *imageUsecase) BulkAddToFolder(ctx context.Context, userID string, imageIDs []uuid.UUID, folderID uuid.UUID) (int, error) {
	ctx, span := u.tel.Tracer.Start(ctx, "usecase.BulkAddToFolder")
	defer span.End()

	logger := observability.LoggerFromContext(ctx, u.tel.Logger)

	if _, err := u.folderRepo.GetByID(ctx, folderID, userID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}

	ownedIDs, err := u.imageRepo.FilterOwnedImageIDs(ctx, imageIDs, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}

	ownedSet := make(map[uuid.UUID]struct{}, len(ownedIDs))
	for _, id := range ownedIDs {
		ownedSet[id] = struct{}{}
	}

	succeeded := 0
	for _, imageID := range imageIDs {
		if _, ok := ownedSet[imageID]; !ok {
			logger.Info("skipping unowned image in bulk add-to-folder",
				zap.String("event", "image.bulk_add_to_folder.skipped"),
				zap.String("image_id", imageID.String()),
				zap.String("user_id", userID),
			)
			continue
		}

		if err := u.imageRepo.AddImageToFolder(ctx, imageID, folderID); err != nil {
			logger.Warn("failed to add image to folder in bulk add-to-folder",
				zap.String("event", "image.bulk_add_to_folder.failed"),
				zap.String("image_id", imageID.String()),
				zap.String("user_id", userID),
				zap.Error(err),
			)
			continue
		}
		succeeded++
	}

	logger.Info("bulk add to folder complete",
		zap.String("event", "image.bulk_add_to_folder.complete"),
		zap.String("user_id", userID),
		zap.String("folder_id", folderID.String()),
		zap.Int("succeeded_count", succeeded),
	)

	return succeeded, nil
}

func downloadFileExtension(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	case "image/gif":
		return "gif"
	case "image/avif":
		return "avif"
	default:
		return "bin"
	}
}
