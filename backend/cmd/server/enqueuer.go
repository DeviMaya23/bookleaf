package main

import (
	"context"

	"github.com/devi/bookleaf/internal/worker"
	pgx "github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/google/uuid"
)

type riverEnqueuer struct {
	client *river.Client[pgx.Tx]
}

func (e *riverEnqueuer) EnqueueAccountWipe(ctx context.Context, userID string) error {
	args := worker.AccountWipeArgs{UserID: userID}
	_, err := e.client.Insert(ctx, args, &river.InsertOpts{MaxAttempts: args.MaxAttempts()})
	return err
}

func (e *riverEnqueuer) EnqueueAccountWipeUnique(ctx context.Context, userID string) error {
	args := worker.AccountWipeArgs{UserID: userID}
	_, err := e.client.Insert(ctx, args, &river.InsertOpts{
		MaxAttempts: args.MaxAttempts(),
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: []rivertype.JobState{rivertype.JobStateAvailable, rivertype.JobStateScheduled, rivertype.JobStatePending, rivertype.JobStateRunning, rivertype.JobStateRetryable, rivertype.JobStateDiscarded},
		},
	})
	return err
}

func (e *riverEnqueuer) EnqueueBookletUserDeletion(ctx context.Context, userID string) error {
	args := worker.BookletUserDeletionArgs{UserID: userID}
	_, err := e.client.Insert(ctx, args, &river.InsertOpts{MaxAttempts: args.MaxAttempts()})
	return err
}

func (e *riverEnqueuer) EnqueueR2Delete(ctx context.Context, r2Path string, thumbnailPath *string) error {
	args := worker.R2DeleteArgs{R2Path: r2Path, ThumbnailPath: thumbnailPath}
	_, err := e.client.Insert(ctx, args, &river.InsertOpts{MaxAttempts: args.MaxAttempts()})
	return err
}

func (e *riverEnqueuer) EnqueueVision(ctx context.Context, imageID uuid.UUID, userID string) error {
	args := worker.VisionArgs{ImageID: imageID, UserID: userID}
	_, err := e.client.Insert(ctx, args, &river.InsertOpts{MaxAttempts: args.MaxAttempts()})
	return err
}

func (e *riverEnqueuer) EnqueueCategoriseImage(ctx context.Context, imageID uuid.UUID, userID string) error {
	args := worker.CategoriseImageArgs{ImageID: imageID, UserID: userID}
	_, err := e.client.Insert(ctx, args, &river.InsertOpts{MaxAttempts: args.MaxAttempts()})
	return err
}
