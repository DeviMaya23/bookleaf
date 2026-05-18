## ADDED Requirements

### Requirement: GeneratePresignedDownloadURL on StorageService

The `StorageService` interface SHALL define a `GeneratePresignedDownloadURL` method that produces a presigned GET URL with `response-content-disposition=attachment; filename="<filename>"` injected as a query parameter.

Signature:
```go
GeneratePresignedDownloadURL(ctx context.Context, key, filename string, ttl time.Duration) (string, error)
```

- `key` is the R2 object key
- `filename` is the value used in the `Content-Disposition` header (e.g., `"my-photo.jpg"`)
- `ttl` controls URL expiry
- The concrete R2 implementation SHALL set the `ResponseContentDisposition` field on the presign request to `attachment; filename="<filename>"`

#### Scenario: Interface is satisfied by R2 implementation

- **WHEN** the Go package is compiled
- **THEN** `r2Storage` implements `StorageService` including `GeneratePresignedDownloadURL` without compilation errors

#### Scenario: Generated URL forces browser download

- **WHEN** `GeneratePresignedDownloadURL` is called with a valid key, filename, and TTL
- **THEN** the returned URL contains a `response-content-disposition` query parameter with value `attachment; filename="<filename>"`

---

### Requirement: DownloadImage Usecase Method

The `ImageUsecase` interface SHALL define a `DownloadImage` method. The concrete implementation SHALL:

1. Fetch the image record via `imageRepo.GetByID(ctx, id, userID)` (ownership enforced)
2. Derive the download filename as `<image.Title>.<ext>` where `ext` is determined from `image.MIMEType` using the existing MIME-to-extension mapping (`image/jpeg` → `jpg`, `image/png` → `png`, `image/webp` → `webp`, `image/gif` → `gif`; unknown types default to `bin`)
3. Call `store.GeneratePresignedDownloadURL(ctx, image.R2Path, filename, 5*time.Minute)`
4. Return the presigned URL string

Signature:
```go
DownloadImage(ctx context.Context, id uuid.UUID, userID string) (string, error)
```

#### Scenario: Returns presigned download URL for owned image

- **WHEN** `DownloadImage` is called with a valid image ID owned by the user
- **THEN** it returns a non-empty presigned URL and no error

#### Scenario: Returns error when image not found or not owned

- **WHEN** `DownloadImage` is called with an image ID that does not exist or belongs to another user
- **THEN** it returns an empty string and a non-nil error

#### Scenario: Returns error when presigning fails

- **WHEN** `GeneratePresignedDownloadURL` returns an error
- **THEN** `DownloadImage` returns an empty string and a non-nil error

---

### Requirement: GET /images/:id/download Handler

The system SHALL expose `GET /images/:id/download` as an authenticated route. The handler SHALL:

1. Extract `userID` from the Kinde JWT context
2. Parse `:id` as a UUID; return `400 Bad Request` on parse failure
3. Call `imageUsecase.DownloadImage(ctx, id, userID)`
4. On success: return `200 OK` with JSON body `{ "download_url": "<presigned-url>" }`
5. On not-found / ownership error: return `404 Not Found`
6. On presigning error: return `500 Internal Server Error`

#### Scenario: Authenticated owner receives download URL

- **WHEN** an authenticated `GET /images/:id/download` request is made for an image owned by the user
- **THEN** the response is `200 OK`
- **AND** the body is `{ "download_url": "<non-empty string>" }`

#### Scenario: Invalid UUID returns 400

- **WHEN** `GET /images/not-a-uuid/download` is requested
- **THEN** the response is `400 Bad Request`

#### Scenario: Image not found or not owned returns 404

- **WHEN** `GET /images/:id/download` is requested for an image that does not exist or belongs to another user
- **THEN** the response is `404 Not Found`

#### Scenario: Unauthenticated request returns 401

- **WHEN** `GET /images/:id/download` is called without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: Image Download Route Registration

The system SHALL register the download route on the protected Echo group in `main.go`.

New route:
- `GET /images/:id/download`

#### Scenario: Download route is registered under auth middleware

- **WHEN** the server starts
- **THEN** `GET /images/:id/download` requires a valid Kinde Bearer token
- **AND** unauthenticated requests return `401 Unauthorized`
