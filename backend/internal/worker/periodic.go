package worker

import (
	"context"
	"time"

	"github.com/riverqueue/river"
)

// BackfillPhash periodic job

type backfillPhashUsecase interface {
	BackfillPhash(ctx context.Context, batchSize int) error
}

type BackfillPhashArgs struct{}

func (BackfillPhashArgs) Kind() string { return "backfill_phash" }

type BackfillPhashWorker struct {
	river.WorkerDefaults[BackfillPhashArgs]
	usecase backfillPhashUsecase
}

func NewBackfillPhashWorker(uc backfillPhashUsecase) *BackfillPhashWorker {
	return &BackfillPhashWorker{usecase: uc}
}

func (w *BackfillPhashWorker) Work(ctx context.Context, _ *river.Job[BackfillPhashArgs]) error {
	return w.usecase.BackfillPhash(ctx, 20)
}

// CleanupStaleUploads periodic job

type cleanupUsecase interface {
	CleanupStaleUploads(ctx context.Context, threshold time.Duration) error
}

type CleanupStaleUploadsArgs struct{}

func (CleanupStaleUploadsArgs) Kind() string { return "cleanup_stale_uploads" }

type CleanupStaleUploadsWorker struct {
	river.WorkerDefaults[CleanupStaleUploadsArgs]
	usecase cleanupUsecase
}

func NewCleanupStaleUploadsWorker(uc cleanupUsecase) *CleanupStaleUploadsWorker {
	return &CleanupStaleUploadsWorker{usecase: uc}
}

func (w *CleanupStaleUploadsWorker) Work(ctx context.Context, job *river.Job[CleanupStaleUploadsArgs]) error {
	return w.usecase.CleanupStaleUploads(ctx, 30*time.Minute)
}

// TrashPurge periodic job

type trashUsecase interface {
	PurgeExpiredTrash(ctx context.Context, threshold time.Duration) error
}

type TrashPurgeArgs struct{}

func (TrashPurgeArgs) Kind() string { return "trash_purge" }

type TrashPurgeWorker struct {
	river.WorkerDefaults[TrashPurgeArgs]
	usecase trashUsecase
}

func NewTrashPurgeWorker(uc trashUsecase) *TrashPurgeWorker {
	return &TrashPurgeWorker{usecase: uc}
}

func (w *TrashPurgeWorker) Work(ctx context.Context, job *river.Job[TrashPurgeArgs]) error {
	return w.usecase.PurgeExpiredTrash(ctx, 30*24*time.Hour)
}

// AccountKindeDeletionReconcile periodic job

type accountKindeDeletionReconcileUsecase interface {
	ReconcilePendingKindeDeletions(ctx context.Context) error
}

type AccountKindeDeletionReconcileArgs struct{}

func (AccountKindeDeletionReconcileArgs) Kind() string { return "account_kinde_deletion_reconcile" }

type AccountKindeDeletionReconcileWorker struct {
	river.WorkerDefaults[AccountKindeDeletionReconcileArgs]
	usecase accountKindeDeletionReconcileUsecase
}

func NewAccountKindeDeletionReconcileWorker(uc accountKindeDeletionReconcileUsecase) *AccountKindeDeletionReconcileWorker {
	return &AccountKindeDeletionReconcileWorker{usecase: uc}
}

func (w *AccountKindeDeletionReconcileWorker) Work(ctx context.Context, _ *river.Job[AccountKindeDeletionReconcileArgs]) error {
	return w.usecase.ReconcilePendingKindeDeletions(ctx)
}
