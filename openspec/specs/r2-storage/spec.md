## ADDED Requirements

### Requirement: R2 Config

The system SHALL load R2 connection settings from environment variables and expose them via the existing `Config` struct.

New fields in `config.go`:
- `R2_ACCOUNT_ID` — required; Cloudflare account ID
- `R2_ACCESS_KEY_ID` — required; R2 API access key ID
- `R2_SECRET_ACCESS_KEY` — required; R2 API secret access key
- `R2_BUCKET_NAME` — required; target bucket name
- `R2_PUBLIC_URL` — required; CDN base URL (e.g. `https://assets.bookleaf.app`)

All five are required. `Load()` SHALL return an error if any are missing.

#### Scenario: All R2 vars set — config loads successfully

- **WHEN** all five R2 environment variables are set
- **THEN** `Load()` returns a `Config` with a populated `R2` field

#### Scenario: Missing R2 var — config fails with descriptive error

- **WHEN** any one of the five R2 environment variables is missing or empty
- **THEN** `Load()` returns an error naming the missing variable

---

### Requirement: StorageService Interface

The system SHALL define a `StorageService` interface in `internal/storage/` that abstracts R2 operations. The image usecase SHALL depend on this interface, not the concrete R2 implementation.

Methods:
- `GeneratePresignedPutURL(ctx, key, contentType string, ttl time.Duration) (string, error)`
- `GeneratePresignedGetURL(ctx, key string, ttl time.Duration) (string, error)`
- `GetObject(ctx, key string) (io.ReadCloser, error)`
- `PutObject(ctx, key string, body io.Reader, contentType string) error`
- `CDNUrl(key string) string` — constructs the public CDN URL for a given key (no network call)
- `Ping(ctx context.Context) error` — verifies R2 connectivity and credential validity; implemented as a `HeadBucket` call; returns `nil` on success or a wrapped error on failure

#### Scenario: Interface is satisfied by R2 implementation

- **WHEN** the Go package is compiled
- **THEN** `r2Storage` in `internal/storage/` implements `StorageService` without compilation errors

#### Scenario: Ping succeeds when bucket is reachable

- **WHEN** `Ping` is called and R2 is reachable with valid credentials
- **THEN** it returns `nil`

#### Scenario: Ping fails when R2 is unreachable or credentials are invalid

- **WHEN** `Ping` is called and R2 returns an error (network failure, 403, etc.)
- **THEN** it returns a non-nil error describing the failure

---

### Requirement: R2 Storage Implementation

The system SHALL implement `StorageService` using the AWS SDK v2 S3-compatible client pointed at Cloudflare R2.

- R2 endpoint: `https://{R2_ACCOUNT_ID}.r2.cloudflarestorage.com`
- `GeneratePresignedPutURL` SHALL produce a PUT presigned URL valid for the given TTL
- `GeneratePresignedGetURL` SHALL produce a GET presigned URL valid for the given TTL
- `GetObject` SHALL fetch an object's body from R2
- `PutObject` SHALL upload a byte stream to R2 at the given key with the given content type
- `CDNUrl` SHALL return `{R2_PUBLIC_URL}/{key}`

#### Scenario: CDN URL is constructed without a network call

- **WHEN** `CDNUrl("users/kp_abc123/thumbnails/img-uuid.jpg")` is called
- **THEN** the return value is `{R2_PUBLIC_URL}/users/kp_abc123/thumbnails/img-uuid.jpg`

#### Scenario: Presigned PUT URL targets the correct key

- **WHEN** `GeneratePresignedPutURL` is called with a key and TTL
- **THEN** the returned URL contains the key path and an expiry consistent with the TTL

---

### Requirement: R2 Object Key Scheme

The system SHALL use the following key structure for all objects stored in R2:

- Full-resolution images: `users/{kindeID}/images/{imageID}.{ext}`
- Thumbnails: `users/{kindeID}/thumbnails/{imageID}.jpg`

Extension is derived from the image MIME type at record creation time (`image/jpeg` → `.jpg`, `image/png` → `.png`, `image/webp` → `.webp`, `image/gif` → `.gif`). Thumbnails are always stored as JPEG regardless of the original format.

#### Scenario: Key scheme is consistent between upload and thumbnail

- **WHEN** an image is created with ID `abc` and mime type `image/png` for user `kp_xyz`
- **THEN** the full-image key is `users/kp_xyz/images/abc.png`
- **AND** the thumbnail key is `users/kp_xyz/thumbnails/abc.jpg`
