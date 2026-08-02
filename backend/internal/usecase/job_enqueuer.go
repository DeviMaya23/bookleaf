package usecase

import "context"

// JobArgs is implemented by all async job argument types.
type JobArgs interface {
	Kind() string
	MaxAttempts() int
}

// JobEnqueuer inserts async jobs into the job queue.
type JobEnqueuer interface {
	Insert(ctx context.Context, args JobArgs) error
	// InsertUnique inserts a job that is deduplicated by (kind, args) while the
	// job is pending or running. A duplicate insert is treated as a no-op success.
	InsertUnique(ctx context.Context, args JobArgs) error
}
