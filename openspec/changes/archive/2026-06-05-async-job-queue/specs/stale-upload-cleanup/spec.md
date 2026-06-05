## MODIFIED Requirements

### Requirement: Background periodic job runs cleanup every 10 minutes

The system SHALL run `CleanupStaleUploads` as a River periodic job firing every 10 minutes with a 30-minute stale threshold, replacing the ticker goroutine previously started in `main.go`.

The ticker goroutine, `startWorkers` call, `imageWorkerUsecase` interface, and `compositeImageWorker` struct SHALL be removed from `main.go`.

#### Scenario: Cleanup fires on schedule via River

- **WHEN** the server is running
- **THEN** `CleanupStaleUploads(ctx, 30*time.Minute)` is invoked approximately every 10 minutes by River's periodic job scheduler

#### Scenario: main.go contains no ticker goroutine for cleanup

- **WHEN** the Go package is compiled
- **THEN** `main.go` contains no `time.NewTicker` call for stale upload cleanup
