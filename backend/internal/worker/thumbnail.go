package worker

import (
	"context"

	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
)

type thumbnailUsecase interface {
	ProcessThumbnailUpload(ctx context.Context, imageID uuid.UUID, r2Path, thumbnailKey string) error
}

type ThumbnailUploadWorker struct {
	river.WorkerDefaults[usecase.ThumbnailUploadArgs]
	usecase thumbnailUsecase
}

func NewThumbnailUploadWorker(uc thumbnailUsecase) *ThumbnailUploadWorker {
	return &ThumbnailUploadWorker{usecase: uc}
}

func (w *ThumbnailUploadWorker) Work(ctx context.Context, job *river.Job[usecase.ThumbnailUploadArgs]) error {
	return w.usecase.ProcessThumbnailUpload(ctx, job.Args.ImageID, job.Args.R2Path, job.Args.ThumbnailKey)
}
