package worker

import (
	"context"

	"github.com/devi/bookleaf/internal/usecase"
	"github.com/riverqueue/river"
)

type bookletUserDeletionUsecase interface {
	ProcessBookletUserDeletion(ctx context.Context, userID string) error
}

type BookletUserDeletionWorker struct {
	river.WorkerDefaults[usecase.BookletUserDeletionArgs]
	usecase bookletUserDeletionUsecase
}

func NewBookletUserDeletionWorker(uc bookletUserDeletionUsecase) *BookletUserDeletionWorker {
	return &BookletUserDeletionWorker{usecase: uc}
}

func (w *BookletUserDeletionWorker) Work(ctx context.Context, job *river.Job[usecase.BookletUserDeletionArgs]) error {
	return w.usecase.ProcessBookletUserDeletion(ctx, job.Args.UserID)
}
