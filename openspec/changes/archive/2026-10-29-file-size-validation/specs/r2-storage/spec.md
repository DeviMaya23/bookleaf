# Spec Delta

## MODIFIED Requirements

### Requirement: StorageService Interface

The system SHALL define a `StorageService` interface in `internal/storage/` that abstracts R2 operations. The image usecase SHALL depend on this interface, not the concrete R2 implementation.

Methods:
- `GeneratePresignedPutURL(ctx, key, contentType string, contentLength int64, ttl time.Duration) (string, error)` — when `contentLength` is greater than zero, the generated presigned URL SHALL include `content-length` in its signed headers; R2 will reject any PUT request whose `Content-Length` header does not match the signed value
- `GeneratePresignedGetURL(ctx, key string, ttl time.Duration) (string, error)`
- `GetObject(ctx, key string) (io.ReadCloser, error)`
- `PutObject(ctx, key string, body io.Reader, contentType string) error`
- `DeleteObject(ctx, key string) error`
- `Ping(ctx context.Context) error` — verifies R2 connectivity and credential validity; implemented as a `HeadBucket` call; returns `nil` on success or a wrapped error on failure

`CDNUrl` is removed. The bucket is private; all read access goes through presigned GET URLs.

#### Scenario: Interface is satisfied by R2 implementation

- **WHEN** the Go package is compiled
- **THEN** `r2Storage` in `internal/storage/` implements `StorageService` without compilation errors

#### Scenario: Ping succeeds when bucket is reachable

- **WHEN** `Ping` is called and R2 is reachable with valid credentials
- **THEN** it returns `nil`

#### Scenario: Ping fails when R2 is unreachable or credentials are invalid

- **WHEN** `Ping` is called and R2 returns an error (network failure, 403, etc.)
- **THEN** it returns a non-nil error describing the failure

#### Scenario: Presigned PUT URL with contentLength signs the Content-Length header

- **WHEN** `GeneratePresignedPutURL` is called with a non-zero `contentLength`
- **THEN** the returned URL contains `content-length` in its `X-Amz-SignedHeaders` parameter

#### Scenario: Presigned PUT URL without contentLength does not sign Content-Length

- **WHEN** `GeneratePresignedPutURL` is called with `contentLength` equal to zero
- **THEN** the returned URL does not include `content-length` in its `X-Amz-SignedHeaders` parameter
