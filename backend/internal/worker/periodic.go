package worker

import (
	"context"
	"time"

	"github.com/riverqueue/river"
)

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
