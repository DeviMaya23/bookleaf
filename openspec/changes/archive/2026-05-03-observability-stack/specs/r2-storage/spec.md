## MODIFIED Requirements

### Requirement: StorageService Interface

ADD the following method to the `StorageService` interface in `internal/storage/`:

- `Ping(ctx context.Context) error` — verifies R2 connectivity and credential validity; implemented as a `HeadBucket` call; returns `nil` on success or a wrapped error on failure

This method is used by the health handler to probe R2 liveness without reading or writing any objects.

#### Scenario: Ping succeeds when bucket is reachable

- **WHEN** `Ping` is called and R2 is reachable with valid credentials
- **THEN** it returns `nil`

#### Scenario: Ping fails when R2 is unreachable or credentials are invalid

- **WHEN** `Ping` is called and R2 returns an error (network failure, 403, etc.)
- **THEN** it returns a non-nil error describing the failure
