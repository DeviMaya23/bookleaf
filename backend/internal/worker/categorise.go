package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/devi/bookleaf/internal/domain"
	"github.com/devi/bookleaf/internal/usecase"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
)

type categorisationUsecase interface {
	CategoriseImage(ctx context.Context, userID string, imageID uuid.UUID) error
}

type categorisationBroadcaster interface {
	Publish(userID string, event domain.Event)
}

type CategorisationWorker struct {
	river.WorkerDefaults[usecase.CategoriseImageArgs]
	usecase     categorisationUsecase
	broadcaster categorisationBroadcaster
}

func NewCategorisationWorker(uc categorisationUsecase, broadcaster categorisationBroadcaster) *CategorisationWorker {
	return &CategorisationWorker{usecase: uc, broadcaster: broadcaster}
}

func (w *CategorisationWorker) NextRetry(job *river.Job[usecase.CategoriseImageArgs]) time.Time {
	return time.Now().Add(30 * time.Second)
}

func (w *CategorisationWorker) Work(ctx context.Context, job *river.Job[usecase.CategoriseImageArgs]) error {
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
