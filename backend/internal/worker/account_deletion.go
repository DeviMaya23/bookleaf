package worker

import (
	"context"

	"github.com/riverqueue/river"
)

// AccountWipe

type accountWipeUsecase interface {
	WipeAccount(ctx context.Context, userID string) error
}

type AccountWipeArgs struct {
	UserID string `json:"user_id"`
}

func (AccountWipeArgs) Kind() string     { return "account_wipe" }
func (AccountWipeArgs) MaxAttempts() int { return 5 }

type AccountWipeWorker struct {
	river.WorkerDefaults[AccountWipeArgs]
	usecase accountWipeUsecase
}

func NewAccountWipeWorker(uc accountWipeUsecase) *AccountWipeWorker {
	return &AccountWipeWorker{usecase: uc}
}

func (w *AccountWipeWorker) Work(ctx context.Context, job *river.Job[AccountWipeArgs]) error {
	return w.usecase.WipeAccount(ctx, job.Args.UserID)
}

// BookletUserDeletion

type bookletUserDeletionUsecase interface {
	ProcessBookletUserDeletion(ctx context.Context, userID string) error
}

type BookletUserDeletionArgs struct {
	UserID string `json:"user_id"`
}

func (BookletUserDeletionArgs) Kind() string     { return "booklet_user_deletion" }
func (BookletUserDeletionArgs) MaxAttempts() int { return 10 }

type BookletUserDeletionWorker struct {
	river.WorkerDefaults[BookletUserDeletionArgs]
	usecase bookletUserDeletionUsecase
}

func NewBookletUserDeletionWorker(uc bookletUserDeletionUsecase) *BookletUserDeletionWorker {
	return &BookletUserDeletionWorker{usecase: uc}
}

func (w *BookletUserDeletionWorker) Work(ctx context.Context, job *river.Job[BookletUserDeletionArgs]) error {
	return w.usecase.ProcessBookletUserDeletion(ctx, job.Args.UserID)
}

// AccountWipeReconcile (periodic)

type accountWipeReconcileUsecase interface {
	ReconcilePendingDeletions(ctx context.Context) error
}

type AccountWipeReconcileArgs struct{}

func (AccountWipeReconcileArgs) Kind() string     { return "account_wipe_reconcile" }
func (AccountWipeReconcileArgs) MaxAttempts() int { return 3 }

type AccountWipeReconcileWorker struct {
	river.WorkerDefaults[AccountWipeReconcileArgs]
	usecase accountWipeReconcileUsecase
}

func NewAccountWipeReconcileWorker(uc accountWipeReconcileUsecase) *AccountWipeReconcileWorker {
	return &AccountWipeReconcileWorker{usecase: uc}
}

func (w *AccountWipeReconcileWorker) Work(ctx context.Context, _ *river.Job[AccountWipeReconcileArgs]) error {
	return w.usecase.ReconcilePendingDeletions(ctx)
}

// PurgedAccountSweep (periodic)

type purgedAccountSweepUsecase interface {
	SweepPurgedAccounts(ctx context.Context) error
}

type PurgedAccountSweepArgs struct{}

func (PurgedAccountSweepArgs) Kind() string     { return "purged_account_sweep" }
func (PurgedAccountSweepArgs) MaxAttempts() int { return 3 }

type PurgedAccountSweepWorker struct {
	river.WorkerDefaults[PurgedAccountSweepArgs]
	usecase purgedAccountSweepUsecase
}

func NewPurgedAccountSweepWorker(uc purgedAccountSweepUsecase) *PurgedAccountSweepWorker {
	return &PurgedAccountSweepWorker{usecase: uc}
}

func (w *PurgedAccountSweepWorker) Work(ctx context.Context, _ *river.Job[PurgedAccountSweepArgs]) error {
	return w.usecase.SweepPurgedAccounts(ctx)
}
