package worker

import (
	"context"

	"github.com/devi/bookleaf/internal/usecase"
	"github.com/riverqueue/river"
)

type accountKindeDeletionUsecase interface {
	ProcessAccountKindeDeletion(ctx context.Context, userID string) error
}

type AccountKindeDeletionWorker struct {
	river.WorkerDefaults[usecase.AccountKindeDeletionArgs]
	usecase accountKindeDeletionUsecase
}

func NewAccountKindeDeletionWorker(uc accountKindeDeletionUsecase) *AccountKindeDeletionWorker {
	return &AccountKindeDeletionWorker{usecase: uc}
}

func (w *AccountKindeDeletionWorker) Work(ctx context.Context, job *river.Job[usecase.AccountKindeDeletionArgs]) error {
	return w.usecase.ProcessAccountKindeDeletion(ctx, job.Args.UserID)
}
