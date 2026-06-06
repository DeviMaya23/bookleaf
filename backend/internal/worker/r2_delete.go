package worker

import (
	"context"

	"github.com/devi/bookleaf/internal/usecase"
	"github.com/riverqueue/river"
)

type r2DeleteUsecase interface {
	ProcessR2Delete(ctx context.Context, r2Path string, thumbnailPath *string) error
}

type R2DeleteWorker struct {
	river.WorkerDefaults[usecase.R2DeleteArgs]
	usecase r2DeleteUsecase
}

func NewR2DeleteWorker(uc r2DeleteUsecase) *R2DeleteWorker {
	return &R2DeleteWorker{usecase: uc}
}

func (w *R2DeleteWorker) Work(ctx context.Context, job *river.Job[usecase.R2DeleteArgs]) error {
	return w.usecase.ProcessR2Delete(ctx, job.Args.R2Path, job.Args.ThumbnailPath)
}
