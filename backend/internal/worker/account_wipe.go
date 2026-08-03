package worker

import (
	"context"

	"github.com/devi/bookleaf/internal/usecase"
	"github.com/riverqueue/river"
)

type accountWipeUsecase interface {
	WipeAccount(ctx context.Context, userID string) error
}

type AccountWipeWorker struct {
	river.WorkerDefaults[usecase.AccountWipeArgs]
	usecase accountWipeUsecase
}

func NewAccountWipeWorker(uc accountWipeUsecase) *AccountWipeWorker {
	return &AccountWipeWorker{usecase: uc}
}

func (w *AccountWipeWorker) Work(ctx context.Context, job *river.Job[usecase.AccountWipeArgs]) error {
	return w.usecase.WipeAccount(ctx, job.Args.UserID)
}
