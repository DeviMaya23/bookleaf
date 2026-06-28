package worker

import (
	"context"
	"time"

	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
)

type categorisationUsecase interface {
	CategoriseImage(ctx context.Context, userID string, imageID uuid.UUID) error
}

type CategorisationWorker struct {
	river.WorkerDefaults[usecase.CategoriseImageArgs]
	usecase categorisationUsecase
}

func NewCategorisationWorker(uc categorisationUsecase) *CategorisationWorker {
	return &CategorisationWorker{usecase: uc}
}

func (w *CategorisationWorker) NextRetry(job *river.Job[usecase.CategoriseImageArgs]) time.Time {
	return time.Now().Add(30 * time.Second)
}

func (w *CategorisationWorker) Work(ctx context.Context, job *river.Job[usecase.CategoriseImageArgs]) error {
	return w.usecase.CategoriseImage(ctx, job.Args.UserID, job.Args.ImageID)
}
