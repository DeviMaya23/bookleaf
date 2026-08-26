package worker

import (
	"context"
	"time"

	"github.com/riverqueue/river"
)

// CleanupStaleUploads (periodic)

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

func (w *CleanupStaleUploadsWorker) Work(ctx context.Context, _ *river.Job[CleanupStaleUploadsArgs]) error {
	return w.usecase.CleanupStaleUploads(ctx, 30*time.Minute)
}

// TrashPurge (periodic)

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

func (w *TrashPurgeWorker) Work(ctx context.Context, _ *river.Job[TrashPurgeArgs]) error {
	return w.usecase.PurgeExpiredTrash(ctx, 30*24*time.Hour)
}

// R2Delete

type r2DeleteUsecase interface {
	ProcessR2Delete(ctx context.Context, r2Path string, thumbnailPath *string) error
}

type R2DeleteArgs struct {
	R2Path        string  `json:"r2_path"`
	ThumbnailPath *string `json:"thumbnail_path"`
}

func (R2DeleteArgs) Kind() string     { return "r2_delete" }
func (R2DeleteArgs) MaxAttempts() int { return 5 }

type R2DeleteWorker struct {
	river.WorkerDefaults[R2DeleteArgs]
	usecase r2DeleteUsecase
}

func NewR2DeleteWorker(uc r2DeleteUsecase) *R2DeleteWorker {
	return &R2DeleteWorker{usecase: uc}
}

func (w *R2DeleteWorker) Work(ctx context.Context, job *river.Job[R2DeleteArgs]) error {
	return w.usecase.ProcessR2Delete(ctx, job.Args.R2Path, job.Args.ThumbnailPath)
}
