### Requirement: Per-format accepted type list

The system SHALL define an expanded set of accepted MIME types for file uploads. Acceptance rules per format:

- `image/jpeg`, `image/png`, `image/gif` — accepted on all browsers
- `image/webp` — accepted on all browsers
- `image/avif` — accepted on all browsers
- `image/heic` — accepted on Safari only; detected via user agent at runtime

On non-Safari browsers, `image/heic` SHALL be treated as an unsupported type and rejected with an inline error message.

#### Scenario: WebP file is accepted on all browsers

- **WHEN** the user drops or selects a WebP file on any browser
- **THEN** no validation error is shown
- **AND** the file is staged for upload

#### Scenario: AVIF file is accepted on all browsers

- **WHEN** the user drops or selects an AVIF file on any browser
- **THEN** no validation error is shown
- **AND** the file is staged for upload

#### Scenario: HEIC file is accepted on Safari

- **WHEN** the user drops or selects a HEIC file on Safari
- **THEN** no validation error is shown
- **AND** the file is staged for upload

#### Scenario: HEIC file is rejected on non-Safari browsers

- **WHEN** the user drops or selects a HEIC file on Chrome or Firefox
- **THEN** a Safari-specific error is shown (distinct from the generic unsupported type error), informing the user that HEIC is only supported in Safari
- **AND** the file is not staged for upload

---

### Requirement: HEIC-to-JPEG conversion before upload

When a HEIC file is accepted (Safari only), the system SHALL convert it to JPEG at 93% quality using an HTML canvas before initiating the upload. The converted JPEG bytes SHALL be used for both the original upload and the thumbnail. The MIME type sent to the backend SHALL be `image/jpeg`. The original HEIC file SHALL NOT be uploaded.

Conversion steps:
1. Create an `HTMLImageElement` (or use `createImageBitmap`) to decode the HEIC blob
2. Draw the full-resolution image onto a canvas matching the image's natural dimensions
3. Export as `image/jpeg` at quality `0.93`
4. Use the resulting blob as the upload body

#### Scenario: HEIC file is uploaded as JPEG

- **WHEN** a HEIC file passes validation on Safari and the upload is submitted
- **THEN** the `POST /images` request body contains `mime_type: "image/jpeg"`
- **AND** the bytes PUT to R2 are JPEG, not HEIC

#### Scenario: HEIC conversion preserves full resolution

- **WHEN** a HEIC file is converted
- **THEN** the canvas dimensions match the image's natural width and height
- **AND** quality is 0.93

---

### Requirement: WebP and AVIF stored as original

WebP and AVIF files SHALL be uploaded to R2 as-is without conversion. The MIME type sent to the backend SHALL match the file's actual type (`image/webp` or `image/avif`).

#### Scenario: WebP file is stored without conversion

- **WHEN** a WebP file is uploaded
- **THEN** the `POST /images` request body contains `mime_type: "image/webp"`
- **AND** the bytes PUT to R2 are the original WebP bytes

#### Scenario: AVIF file is stored without conversion

- **WHEN** an AVIF file is uploaded
- **THEN** the `POST /images` request body contains `mime_type: "image/avif"`
- **AND** the bytes PUT to R2 are the original AVIF bytes

---

### Requirement: Client-side thumbnail generation

Before calling `POST /images/:id/complete`, the system SHALL generate a JPEG thumbnail from the upload file (or converted file, in the case of HEIC) and PUT it to the `thumbnail_upload_url` returned by `POST /images`.

Thumbnail generation steps:
1. Use `createImageBitmap` to decode the source blob
2. Draw onto a canvas, fitting within 600×600px while preserving aspect ratio
3. Export as `image/jpeg` at quality `0.9`

The thumbnail upload SHALL be performed in parallel with the original file upload. `POST /images/:id/complete` SHALL only be called after both PUTs succeed.

If thumbnail generation or upload fails, the full upload SHALL be aborted — `POST /images/:id/complete` SHALL NOT be called.

#### Scenario: Thumbnail is uploaded before complete is called

- **WHEN** both the original and thumbnail PUT requests succeed
- **THEN** `POST /images/:id/complete` is called
- **AND** the thumbnail is already present at the thumbnail R2 path

#### Scenario: Thumbnail upload failure aborts the upload

- **WHEN** the PUT to `thumbnail_upload_url` fails
- **THEN** `POST /images/:id/complete` is NOT called
- **AND** an error toast is shown

#### Scenario: Thumbnail fits within 600×600 preserving aspect ratio

- **WHEN** a 1200×800 image is thumbnailed
- **THEN** the thumbnail dimensions are 600×400
