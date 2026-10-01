## ADDED Requirements

### Requirement: GetManyByIDs repository method

The `ImageRepository` interface SHALL define a `GetManyByIDs` method:

```go
GetManyByIDs(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) ([]*domain.Image, error)
```

The implementation SHALL fetch all non-deleted images whose ID is in `ids` AND whose `user_id` matches `userID` in a single query. Results SHALL include Tags and ImageFolders preloaded. IDs not found or not owned by `userID` are silently omitted — no error is returned for missing IDs. Result order is not guaranteed.

#### Scenario: Returns owned images only

- **WHEN** `GetManyByIDs` is called with a list of IDs that includes one owned image and one image belonging to another user
- **THEN** only the owned image is returned
- **AND** no error is returned

#### Scenario: Returns empty slice when no IDs match

- **WHEN** `GetManyByIDs` is called with IDs that do not exist or do not belong to the user
- **THEN** an empty slice is returned with no error

---

### Requirement: writeImagesToZip package-level helper

The `usecase` package SHALL define a package-level function:

```go
func writeImagesToZip(ctx context.Context, images []*domain.Image, store StorageService, w io.Writer) error
```

Behavior:
1. Create a `zip.Writer` wrapping `w`.
2. For each image in order: derive an entry filename from `image.Title` and `image.MIMEType` using the existing `exportEntryName` logic (sanitize path separators; deduplicate colliding names with ` (1)`, ` (2)` suffixes). Call `store.GetObject(ctx, image.R2Path)`, create the zip entry, `io.Copy` the contents, then close the reader. Return a wrapped error if `GetObject` or the copy fails.
3. Close the `zip.Writer`. A zero-image slice produces a valid empty archive.

`folderUsecase.ExportFolder` SHALL be refactored to delegate to this helper instead of containing the loop inline.

#### Scenario: Writes one entry per image with derived names

- **WHEN** `writeImagesToZip` is called with images titled `"Sunset"` (image/jpeg) and `"Portrait"` (image/png)
- **THEN** the resulting zip contains entries `"Sunset.jpg"` and `"Portrait.png"`
- **AND** no error is returned

#### Scenario: Deduplicates colliding entry names

- **WHEN** `writeImagesToZip` is called with two images both titled `"Untitled"` with the same MIME type
- **THEN** the resulting zip contains entries `"Untitled.jpg"` and `"Untitled (1).jpg"`

#### Scenario: Zero images produce a valid empty zip

- **WHEN** `writeImagesToZip` is called with an empty image slice
- **THEN** no error is returned
- **AND** the bytes written to `w` form a valid empty zip archive

#### Scenario: Returns wrapped error when GetObject fails

- **WHEN** `store.GetObject` returns an error for one of the images
- **THEN** `writeImagesToZip` returns a non-nil error wrapping it

---

### Requirement: BulkExportImages usecase method

`ImageUsecase` SHALL define a `BulkExportImages` method:

```go
BulkExportImages(ctx context.Context, userID uuid.UUID, imageIDs []uuid.UUID, w io.Writer) error
```

Behavior:
1. Call `imageRepo.GetManyByIDs(ctx, imageIDs, userID)` to resolve owned images. Return a wrapped error if the query fails.
2. Call `writeImagesToZip(ctx, images, store, w)` and return any error.

IDs not owned by the user are silently dropped (handled by `GetManyByIDs`). A request where all IDs are unowned produces a valid empty zip.

#### Scenario: Streams zip for owned image IDs

- **WHEN** `BulkExportImages` is called with IDs for two owned images
- **THEN** the resulting zip contains one entry per image
- **AND** no error is returned

#### Scenario: Unowned IDs are silently dropped

- **WHEN** `BulkExportImages` is called with a mix of owned and unowned IDs
- **THEN** only the owned images appear in the zip
- **AND** no error is returned

#### Scenario: Returns error when GetManyByIDs fails

- **WHEN** `imageRepo.GetManyByIDs` returns an error
- **THEN** `BulkExportImages` returns a non-nil error wrapping it
- **AND** no bytes are written to `w`

---

### Requirement: POST /images/bulk/export handler

The system SHALL expose `POST /images/bulk/export` on the authenticated route group.

Request body:
```json
{ "image_ids": ["uuid", ...] }
```

Handler behavior:
1. Parse the request body; return `400 Bad Request` if any entry in `image_ids` is not a valid UUID.
2. Extract `userID` from the JWT context.
3. Set `Content-Type: application/zip` and `Content-Disposition: attachment; filename="bookleaf-export.zip"`, write `200 OK`.
4. Call `imageUsecase.BulkExportImages(ctx, userID, imageIDs, c.Response())` to stream the zip body. If it returns an error after the response has started, log it server-side; the response is not modified further.

#### Scenario: Returns streamed zip for valid request

- **WHEN** an authenticated `POST /images/bulk/export` request is made with valid owned image IDs
- **THEN** the response status is `200 OK`
- **AND** `Content-Type` is `application/zip`
- **AND** `Content-Disposition` is `attachment; filename="bookleaf-export.zip"`

#### Scenario: Malformed UUID in image_ids returns 400

- **WHEN** `image_ids` contains a string that is not a valid UUID
- **THEN** the response is `400 Bad Request`
- **AND** no bytes are written to the response body

#### Scenario: Unauthenticated request returns 401

- **WHEN** `POST /images/bulk/export` is called without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

#### Scenario: Unowned IDs are silently excluded from the archive

- **WHEN** `image_ids` contains IDs belonging to another user
- **THEN** those images are absent from the zip
- **AND** the response is still `200 OK`

---

### Requirement: Bulk export route registration

The system SHALL register `POST /images/bulk/export` on the protected Echo group in `main.go`, alongside the other `/images/bulk/` routes.

#### Scenario: Route requires authentication

- **WHEN** the server starts
- **THEN** `POST /images/bulk/export` requires a valid Kinde Bearer token
- **AND** unauthenticated requests return `401 Unauthorized`

---

### Requirement: BulkExportImages usecase unit tests

The system SHALL have unit tests for `imageUsecase.BulkExportImages` using mock doubles, covering:

- Zip output contains entries for all owned images
- Unowned IDs are silently dropped (empty zip, no error)
- Error from `GetManyByIDs` is propagated

#### Scenario: Unit test asserts zip contents

- **WHEN** the mock `ImageRepository.GetManyByIDs` returns two images and the mock storage returns readable content for each
- **THEN** the test opens the resulting bytes as a zip archive and asserts two entries with the expected names

#### Scenario: Unit test asserts error propagation

- **WHEN** the mock `ImageRepository.GetManyByIDs` returns an error
- **THEN** the test asserts `BulkExportImages` returns a non-nil error

---

### Requirement: POST /images/bulk/export handler unit tests

The system SHALL have unit tests for the bulk export handler using a mock `ImageUsecase`, covering:

- `200` with correct headers on success
- `400` for a malformed UUID in `image_ids`

#### Scenario: Handler unit test asserts response headers on success

- **WHEN** the mock `ImageUsecase.BulkExportImages` succeeds
- **THEN** the test asserts status `200`, `Content-Type: application/zip`, and `Content-Disposition: attachment; filename="bookleaf-export.zip"`

#### Scenario: Handler unit test asserts 400 for malformed UUID

- **WHEN** `image_ids` contains `"not-a-uuid"`
- **THEN** the test asserts status `400`
