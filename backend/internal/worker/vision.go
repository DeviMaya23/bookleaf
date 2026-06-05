package worker

import (
	"context"
	"time"

	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
)

type visionUsecase interface {
	ProcessVisionLabelling(ctx context.Context, imageID uuid.UUID, userID string) error
}

type VisionWorker struct {
	river.WorkerDefaults[usecase.VisionArgs]
	usecase visionUsecase
}

func NewVisionWorker(uc visionUsecase) *VisionWorker {
	return &VisionWorker{usecase: uc}
}

func (w *VisionWorker) NextRetry(job *river.Job[usecase.VisionArgs]) time.Time {
	return time.Now().Add(10 * time.Second)
}

func (w *VisionWorker) Work(ctx context.Context, job *river.Job[usecase.VisionArgs]) error {
	return w.usecase.ProcessVisionLabelling(ctx, job.Args.ImageID, job.Args.UserID)
}
