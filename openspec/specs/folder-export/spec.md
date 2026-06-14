## Purpose

Folder export allows a user to download all direct (non-recursive) images in a folder as a single zip archive, streamed from the backend.

## Requirements

### Requirement: FolderImageRepository Interface

The `usecase` package SHALL define a `FolderImageRepository` interface, replacing the narrower `ImageCounter` interface currently used by `folderUsecase`. It SHALL include:

```go
type FolderImageRepository interface {
    CountByFolderID(ctx context.Context, folderID uuid.UUID) (int64, error)
    ListByFolder(ctx context.Context, userID string, folderID uuid.UUID, sortField *string, direction *string) ([]*domain.Image, error)
}
```

The concrete `imageRepository` already implements both methods; only the interface definition and `NewFolderUsecase` construction (and its test doubles) change.

#### Scenario: Interface is satisfied by existing image repository

- **WHEN** the Go package is compiled
- **THEN** `imageRepository` implements `usecase.FolderImageRepository` without compilation errors

---

### Requirement: ExportFolder Usecase Method

`FolderUsecase` SHALL define an `ExportFolder` method that streams a zip archive of a folder's direct (non-recursive) images to a provided writer.

Signature:
```go
ExportFolder(ctx context.Context, folderID uuid.UUID, userID string, w io.Writer) error
```

Behavior:
1. Call `imageRepo.ListByFolder(ctx, userID, folderID, nil, nil)` to get the folder's non-deleted images, ordered by position. Return a wrapped error immediately if this fails — no bytes are written to `w` yet.
2. Create a `zip.Writer` wrapping `w`.
3. For each image, in order:
   - Sanitize `image.Title` by replacing path-separator characters (`/`, `\`) so the title cannot introduce nested paths inside the archive.
   - Derive the entry filename as `<sanitized title>.<ext>`, where `ext` comes from the existing `downloadFileExtension(image.MIMEType)`.
   - If an entry with that filename has already been written in this export, append ` (1)`, ` (2)`, etc. (tracked via a counter map) until the name is unique.
   - Open the image via `store.GetObject(ctx, image.R2Path)`, create the zip entry, `io.Copy` the contents in, then close the reader.
   - If `GetObject` or the copy fails, return a wrapped error (bytes already written to `w` remain as-is — no cleanup).
4. Close the `zip.Writer`. A folder with zero images produces a valid, empty zip archive.

#### Scenario: Writes one entry per image with derived names

- **WHEN** `ExportFolder` is called for a folder containing images titled `"Sunset"` (image/jpeg) and `"Portrait"` (image/png)
- **THEN** the resulting zip contains entries `"Sunset.jpg"` and `"Portrait.png"`
- **AND** `ExportFolder` returns no error

#### Scenario: Deduplicates colliding entry names

- **WHEN** `ExportFolder` is called for a folder containing two images both titled `"Untitled"` with the same MIME type
- **THEN** the resulting zip contains entries `"Untitled.jpg"` and `"Untitled (1).jpg"`

#### Scenario: Sanitizes titles containing path separators

- **WHEN** `ExportFolder` is called for an image titled `"Trip/Day 1"`
- **THEN** the resulting zip entry name does not contain `/` or `\`
- **AND** the entry is a top-level file in the archive, not nested in a subdirectory

#### Scenario: Empty folder produces an empty valid zip

- **WHEN** `ExportFolder` is called for a folder with zero images
- **THEN** it returns no error
- **AND** the bytes written to `w` form a valid (empty) zip archive

#### Scenario: Returns wrapped error when listing images fails

- **WHEN** `imageRepo.ListByFolder` returns an error
- **THEN** `ExportFolder` returns a non-nil error wrapping it
- **AND** no bytes are written to `w`

#### Scenario: Returns wrapped error when fetching an image from storage fails

- **WHEN** `store.GetObject` returns an error for one of the folder's images
- **THEN** `ExportFolder` returns a non-nil error wrapping it

---

### Requirement: GET /folders/:id/export Handler

The system SHALL expose `GET /folders/:id/export` as an authenticated route. The handler SHALL:

1. Extract `userID` from the Kinde JWT context
2. Parse `:id` as a UUID; return `400 Bad Request` on parse failure
3. Call `folderUsecase.GetByID(ctx, id, userID)` to verify the folder exists and is owned by the user; return `404 Not Found` if not, before writing any response body
4. On success, derive the archive filename as `<sanitized folder name>.zip` (sanitize by replacing characters invalid in filenames; if sanitizing leaves an empty string, use `export` as the base name)
5. Set `Content-Type: application/zip` and `Content-Disposition: attachment; filename="<archive filename>"`, write `200 OK`
6. Call `folderUsecase.ExportFolder(ctx, id, userID, c.Response())` to stream the zip body
7. If `ExportFolder` returns an error after the response has started, log it server-side; the response is not modified further (the client observes a truncated download)

#### Scenario: Authenticated owner receives a streamed zip

- **WHEN** an authenticated `GET /folders/:id/export` request is made for a folder owned by the user
- **THEN** the response status is `200 OK`
- **AND** the response has `Content-Type: application/zip`
- **AND** the response has a `Content-Disposition: attachment; filename="<folder name>.zip"` header

#### Scenario: Folder name with invalid filename characters is sanitized

- **WHEN** an authenticated `GET /folders/:id/export` request is made for a folder named `"Trip / 2024"`
- **THEN** the `Content-Disposition` filename does not contain `/`

#### Scenario: Folder name that sanitizes to empty falls back to a default

- **WHEN** an authenticated `GET /folders/:id/export` request is made for a folder whose name consists entirely of characters invalid in filenames
- **THEN** the `Content-Disposition` filename is `"export.zip"`

#### Scenario: Invalid UUID returns 400

- **WHEN** `GET /folders/not-a-uuid/export` is requested
- **THEN** the response is `400 Bad Request`

#### Scenario: Folder not found or not owned returns 404

- **WHEN** an authenticated `GET /folders/:id/export` request is made for a folder that does not exist or belongs to another user
- **THEN** the response is `404 Not Found`
- **AND** no response body is written

#### Scenario: Unauthenticated request returns 401

- **WHEN** `GET /folders/:id/export` is called without a valid Bearer token
- **THEN** the response is `401 Unauthorized`

---

### Requirement: Folder Export Route Registration

The system SHALL register `GET /folders/:id/export` on the protected Echo group in `main.go`, alongside the other `/folders` routes.

#### Scenario: Export route is registered under auth middleware

- **WHEN** the server starts
- **THEN** `GET /folders/:id/export` requires a valid Kinde Bearer token
- **AND** unauthenticated requests return `401 Unauthorized`

---

### Requirement: ExportFolder Usecase Unit Tests

The system SHALL have unit tests for `folderUsecase.ExportFolder` using a fake `FolderImageRepository` and a value-return spy `StorageService`, covering:

- Entry naming derived from image title and MIME type
- Deduplication of colliding entry names
- Sanitization of titles containing path separators
- Empty folder producing a valid empty zip
- Error propagation when `ListByFolder` fails
- Error propagation when `GetObject` fails

#### Scenario: Unit test asserts zip contents for a multi-image folder

- **WHEN** the fake `FolderImageRepository` returns multiple images and the storage spy returns readable content for each
- **THEN** the test opens the resulting bytes as a zip archive and asserts the expected entry names and contents

#### Scenario: Unit test asserts error propagation from storage

- **WHEN** the storage spy's `GetObject` returns an error for one image
- **THEN** the test asserts `ExportFolder` returns a non-nil error

---

### Requirement: Folder Export Handler Unit Tests

The system SHALL have unit tests for `FolderHandler.ExportFolder` using a value-return spy `FolderUsecase`, covering:

- Success: status `200`, `Content-Type: application/zip`, and `Content-Disposition` with the expected filename
- `400` for an invalid UUID
- `404` when `GetByID` returns not-found/not-owned
- Filename sanitization and the empty-name fallback to `export.zip`

#### Scenario: Handler unit test asserts response headers on success

- **WHEN** the spy `FolderUsecase.GetByID` returns a folder named `"My Folder"` and `ExportFolder` succeeds
- **THEN** the test asserts the response status is `200`
- **AND** asserts `Content-Disposition` is `attachment; filename="My Folder.zip"`

#### Scenario: Handler unit test asserts 404 mapping

- **WHEN** the spy `FolderUsecase.GetByID` returns an error
- **THEN** the test asserts the response status is `404`
