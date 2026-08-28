package main

import (
	"context"

	"github.com/devi/bookleaf/internal/worker"
	"github.com/google/uuid"
	pgx "github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type riverEnqueuer struct {
	client *river.Client[pgx.Tx]
}

func (e *riverEnqueuer) EnqueueAccountWipe(ctx context.Context, idpSubject string) error {
	args := worker.AccountWipeArgs{IDPSubject: idpSubject}
	_, err := e.client.Insert(ctx, args, &river.InsertOpts{MaxAttempts: args.MaxAttempts()})
	return err
}

func (e *riverEnqueuer) EnqueueAccountWipeUnique(ctx context.Context, idpSubject string) error {
	args := worker.AccountWipeArgs{IDPSubject: idpSubject}
	_, err := e.client.Insert(ctx, args, &river.InsertOpts{
		MaxAttempts: args.MaxAttempts(),
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: []rivertype.JobState{rivertype.JobStateAvailable, rivertype.JobStateScheduled, rivertype.JobStatePending, rivertype.JobStateRunning, rivertype.JobStateRetryable, rivertype.JobStateDiscarded},
		},
	})
	return err
}

func (e *riverEnqueuer) EnqueueBookletUserDeletion(ctx context.Context, idpSubject string) error {
	args := worker.BookletUserDeletionArgs{IDPSubject: idpSubject}
	_, err := e.client.Insert(ctx, args, &river.InsertOpts{MaxAttempts: args.MaxAttempts()})
	return err
}

func (e *riverEnqueuer) EnqueueR2Delete(ctx context.Context, r2Path string, thumbnailPath *string) error {
	args := worker.R2DeleteArgs{R2Path: r2Path, ThumbnailPath: thumbnailPath}
	_, err := e.client.Insert(ctx, args, &river.InsertOpts{MaxAttempts: args.MaxAttempts()})
	return err
}

func (e *riverEnqueuer) EnqueueVision(ctx context.Context, imageID uuid.UUID, userID uuid.UUID) error {
	args := worker.VisionArgs{ImageID: imageID, UserID: userID}
	_, err := e.client.Insert(ctx, args, &river.InsertOpts{MaxAttempts: args.MaxAttempts()})
	return err
}

func (e *riverEnqueuer) EnqueueCategoriseImage(ctx context.Context, imageID uuid.UUID, userID uuid.UUID) error {
	args := worker.CategoriseImageArgs{ImageID: imageID, UserID: userID}
	_, err := e.client.Insert(ctx, args, &river.InsertOpts{MaxAttempts: args.MaxAttempts()})
	return err
}
