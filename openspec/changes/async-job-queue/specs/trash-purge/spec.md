## MODIFIED Requirements

### Requirement: Background periodic job runs purge every 24 hours

The system SHALL run `PurgeExpiredTrash` as a River periodic job firing every 24 hours with a 30-day retention threshold, replacing the ticker goroutine previously started in `main.go`.

#### Scenario: Purge fires on schedule via River

- **WHEN** the server is running
- **THEN** `PurgeExpiredTrash(ctx, 30*24*time.Hour)` is invoked approximately every 24 hours by River's periodic job scheduler

#### Scenario: main.go contains no ticker goroutine for trash purge

- **WHEN** the Go package is compiled
- **THEN** `main.go` contains no `time.NewTicker` call for trash purge
