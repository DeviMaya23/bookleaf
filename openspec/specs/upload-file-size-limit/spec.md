# upload-file-size-limit Specification

## Purpose

Enforces a 50 MB per-image ceiling at the server and on every client upload surface, preventing oversized uploads from reaching R2 and giving the user immediate feedback before any network request is made.

## Requirements

### Requirement: POST /images requires and validates file_size

`POST /images` SHALL require a `file_size` field in the request body representing the byte length of the file to be uploaded. The server SHALL return `400 Bad Request` when the field is absent or zero, and `413 Request Entity Too Large` when the value exceeds 52,428,800 bytes (50 MiB). No presigned URL is issued on rejection.

#### Scenario: Missing file_size returns 400

- **WHEN** `POST /images` is called without a `file_size` field
- **THEN** the server returns `400 Bad Request`

#### Scenario: Zero file_size returns 400

- **WHEN** `POST /images` is called with `file_size: 0`
- **THEN** the server returns `400 Bad Request`

#### Scenario: file_size above 50 MB returns 413

- **WHEN** `POST /images` is called with `file_size` greater than 52,428,800
- **THEN** the server returns `413 Request Entity Too Large`

#### Scenario: file_size at exactly 50 MB is accepted

- **WHEN** `POST /images` is called with `file_size` equal to 52,428,800
- **THEN** the server returns `201 Created` with a presigned upload URL

#### Scenario: file_size below 50 MB is accepted

- **WHEN** `POST /images` is called with a valid `file_size` below 52,428,800
- **THEN** the server returns `201 Created` with a presigned upload URL

---

### Requirement: POST /images/:id/complete validates file_size

`POST /images/:id/complete` SHALL return `413 Request Entity Too Large` when the `file_size` field in the request body exceeds 52,428,800 bytes. This is a belt-and-suspenders check: a file that passed initiation but is reported at a different size at completion is rejected.

#### Scenario: Oversized file_size at complete returns 413

- **WHEN** `POST /images/:id/complete` is called with `file_size` greater than 52,428,800
- **THEN** the server returns `413 Request Entity Too Large`

#### Scenario: Valid file_size at complete is accepted

- **WHEN** `POST /images/:id/complete` is called with `file_size` at or below 52,428,800
- **THEN** the request proceeds normally

---

### Requirement: Single-file upload modal rejects oversized files

The `UploadModal` SHALL reject a file larger than 50 MB when it is selected, dropped onto the modal drop zone, or passed in as an initial file (e.g. from clipboard paste). An inline error SHALL be shown and the upload button SHALL remain disabled.

#### Scenario: Oversized file selected via picker shows inline error

- **WHEN** the user selects a file larger than 50 MB via the file picker in `UploadModal`
- **THEN** an inline error is shown and the upload button is disabled

#### Scenario: Oversized file dropped onto modal drop zone shows inline error

- **WHEN** the user drops a file larger than 50 MB onto the `UploadModal` drop zone
- **THEN** an inline error is shown and the upload button is disabled

#### Scenario: Oversized file from clipboard paste shows inline error

- **WHEN** the user pastes an image from the clipboard that is larger than 50 MB
- **THEN** `UploadModal` opens with an inline size error shown and the upload button disabled

#### Scenario: Valid file clears the size error

- **WHEN** the user selects or drops a file within the size limit after a size error was shown
- **THEN** the error is cleared and the file is accepted normally

---

### Requirement: Single-file auto-upload rejects oversized files with a toast

When a single file is dragged directly onto the main gallery surface (triggering the silent auto-upload path), the system SHALL show an error toast and not initiate any upload if the file exceeds 50 MB.

#### Scenario: Oversized file dropped on gallery area shows error toast

- **WHEN** the user drops a single file larger than 50 MB onto the main gallery area
- **THEN** an error toast is shown and no `POST /images` request is made
