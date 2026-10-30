# Spec Delta

## MODIFIED Requirements

### Requirement: StorageService Interface

The system SHALL define a `StorageService` interface in `internal/storage/` that abstracts R2 operations. The image usecase SHALL depend on this interface, not the concrete R2 implementation.

Methods:
- `GeneratePresignedPutURL(ctx, key, contentType string, contentLength int64, ttl time.Duration) (string, error)` — when `contentLength` is greater than zero, the generated presigned URL SHALL include `content-length` in its signed headers; R2 will reject any PUT request whose `Content-Length` header does not match the signed value
- `GeneratePresignedGetURL(ctx, key string, ttl time.Duration) (string, error)`
- `GenerateDeterministicPresignedGetURL(ctx context.Context, key string) (string, error)` — produces a stable presigned GET URL valid for a 48 h window; the signing time is truncated to the start of the current 24 h UTC window so every call within the same window returns an identical URL
- `GetObject(ctx, key string) (io.ReadCloser, error)`
- `PutObject(ctx, key string, body io.Reader, contentType, cacheControl string) error` — when `cacheControl` is non-empty, the `Cache-Control` header SHALL be set on the stored object
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

---

### Requirement: R2 Storage Implementation

The system SHALL implement `StorageService` using the AWS SDK v2 S3-compatible client pointed at Cloudflare R2.

- R2 endpoint: `https://{R2_ACCOUNT_ID}.r2.cloudflarestorage.com`
- `GeneratePresignedPutURL` SHALL produce a PUT presigned URL valid for the given TTL
- `GeneratePresignedGetURL` SHALL produce a GET presigned URL valid for the given TTL
- `GenerateDeterministicPresignedGetURL` SHALL sign with the start of the current UTC day (truncated to 24 h) and set expiry to 48 h, so URLs are identical for all calls within the same window
- `GetObject` SHALL fetch an object's body from R2
- `PutObject` SHALL upload a byte stream to R2 at the given key with the given content type; when `cacheControl` is non-empty it SHALL be set as the object's `Cache-Control` metadata
- `DeleteObject` SHALL delete the object at the given key from R2
- The implementation SHALL expose a `now func() time.Time` field (defaulting to `time.Now`) so the signing window can be controlled in tests

#### Scenario: Presigned PUT URL targets the correct key

- **WHEN** `GeneratePresignedPutURL` is called with a key and TTL
- **THEN** the returned URL contains the key path and an expiry consistent with the TTL

#### Scenario: Deterministic URL is identical within the same window

- **WHEN** `GenerateDeterministicPresignedGetURL` is called twice with the same key within the same 24 h UTC window
- **THEN** both calls return the identical URL

#### Scenario: Deterministic URL changes at window rollover

- **WHEN** `GenerateDeterministicPresignedGetURL` is called with a clock value just before midnight UTC and again just after midnight UTC
- **THEN** the two calls return different URLs

#### Scenario: Deterministic URL from end of window remains valid in next window

- **WHEN** a URL is generated at the last second of a 24 h window
- **THEN** the URL remains valid for at least 24 h (expiry is 48 h, window is 24 h)
