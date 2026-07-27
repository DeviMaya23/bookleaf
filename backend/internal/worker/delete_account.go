package worker

import (
	"context"

	"github.com/devi/bookleaf/internal/usecase"
	"github.com/riverqueue/river"
)

type deleteAccountUsecase interface {
	DeleteAccount(ctx context.Context, userID string) error
}

type DeleteAccountWorker struct {
	river.WorkerDefaults[usecase.DeleteAccountArgs]
	usecase deleteAccountUsecase
}

func NewDeleteAccountWorker(uc deleteAccountUsecase) *DeleteAccountWorker {
	return &DeleteAccountWorker{usecase: uc}
}

func (w *DeleteAccountWorker) Work(ctx context.Context, job *river.Job[usecase.DeleteAccountArgs]) error {
	return w.usecase.DeleteAccount(ctx, job.Args.UserID)
}
