package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
)

// Vision

type visionUsecase interface {
	ProcessVisionLabelling(ctx context.Context, imageID uuid.UUID, userID string) error
}

type VisionArgs struct {
	ImageID uuid.UUID `json:"image_id"`
	UserID  string    `json:"user_id"`
}

func (VisionArgs) Kind() string     { return "vision_labelling" }
func (VisionArgs) MaxAttempts() int { return 3 }

type VisionWorker struct {
	river.WorkerDefaults[VisionArgs]
	usecase visionUsecase
}

func NewVisionWorker(uc visionUsecase) *VisionWorker {
	return &VisionWorker{usecase: uc}
}

func (w *VisionWorker) NextRetry(_ *river.Job[VisionArgs]) time.Time {
	return time.Now().Add(10 * time.Second)
}

func (w *VisionWorker) Work(ctx context.Context, job *river.Job[VisionArgs]) error {
	return w.usecase.ProcessVisionLabelling(ctx, job.Args.ImageID, job.Args.UserID)
}

// CategoriseImage

type categorisationUsecase interface {
	CategoriseImage(ctx context.Context, userID string, imageID uuid.UUID) error
}

type categorisationBroadcaster interface {
	Publish(userID string, event domain.Event)
}

type CategoriseImageArgs struct {
	ImageID uuid.UUID `json:"image_id"`
	UserID  string    `json:"user_id"`
}

func (CategoriseImageArgs) Kind() string     { return "categorise_image" }
func (CategoriseImageArgs) MaxAttempts() int { return 3 }

type CategorisationWorker struct {
	river.WorkerDefaults[CategoriseImageArgs]
	usecase     categorisationUsecase
	broadcaster categorisationBroadcaster
}

func NewCategorisationWorker(uc categorisationUsecase, broadcaster categorisationBroadcaster) *CategorisationWorker {
	return &CategorisationWorker{usecase: uc, broadcaster: broadcaster}
}

func (w *CategorisationWorker) NextRetry(_ *river.Job[CategoriseImageArgs]) time.Time {
	return time.Now().Add(30 * time.Second)
}

func (w *CategorisationWorker) Work(ctx context.Context, job *river.Job[CategoriseImageArgs]) error {
	if err := w.usecase.CategoriseImage(ctx, job.Args.UserID, job.Args.ImageID); err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{"image_id": job.Args.ImageID.String()})
	w.broadcaster.Publish(job.Args.UserID, domain.Event{
		Type:    "categorisation_complete",
		Payload: payload,
	})

	return nil
}

// BackfillPhash (periodic)

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
