## ADDED Requirements

### Requirement: FindDuplicates repository method

The image repository SHALL provide a `FindDuplicates` method that returns all non-deleted images owned by a given user whose `phash` column is within a specified Hamming distance of a query hash, excluding a nominated image ID. The query SHALL be evaluated entirely in Postgres using `bit_count(phash # $hash::bit(64)) <= $threshold`.

#### Scenario: Returns images within Hamming distance threshold

- **WHEN** `FindDuplicates` is called with a phash and threshold of 10
- **THEN** all non-deleted user images with a non-null phash and a Hamming distance ≤ 10 from the query hash are returned
- **AND** the image with the excluded ID is not included even if its hash matches

#### Scenario: Excludes images with null phash

- **WHEN** `FindDuplicates` is called
- **THEN** images where `phash IS NULL` are not returned

#### Scenario: Excludes soft-deleted images

- **WHEN** `FindDuplicates` is called
- **THEN** images with a non-null `deleted_at` are not returned

#### Scenario: Scoped to the requesting user

- **WHEN** `FindDuplicates` is called for user A with a phash that matches an image belonging to user B
- **THEN** user B's image is not returned

#### Scenario: Returns empty slice when no duplicates exist

- **WHEN** `FindDuplicates` is called and no stored images are within the threshold
- **THEN** an empty slice is returned without error

---

### Requirement: CompleteUpload accepts optional phash

The `POST /images/:id/complete` endpoint SHALL accept an optional `phash` string field in the request body. The value SHALL be a 64-character binary string (e.g. `"0110...1010"`) representing the perceptual hash computed client-side. When absent or empty, `phash` SHALL be stored as `NULL` on the image record.

#### Scenario: Request with phash persists the value

- **WHEN** `POST /images/:id/complete` is called with a 64-character `phash` string
- **THEN** the committed `images` row has the `phash` column set to that value

#### Scenario: Request without phash stores null

- **WHEN** `POST /images/:id/complete` is called without a `phash` field
- **THEN** the committed `images` row has `phash = NULL`

---

### Requirement: CompleteUpload returns duplicate matches

After committing the new image, `POST /images/:id/complete` SHALL query for duplicates using a Hamming distance threshold of 10 and return any matches in the response. The response SHALL include a `duplicates` array. Each entry SHALL contain the matching image's `id`, `title`, and `thumbnail_path`. The array SHALL be empty when no duplicates are found or when no `phash` was provided.

#### Scenario: Response includes matches when duplicates exist

- **WHEN** `POST /images/:id/complete` is called with a phash that matches one or more existing images within threshold 10
- **THEN** the response body includes a non-empty `duplicates` array
- **AND** each entry contains the matching image's `id`, `title`, and `thumbnail_path`

#### Scenario: Response includes empty array when no duplicates

- **WHEN** `POST /images/:id/complete` is called with a phash that matches no existing images
- **THEN** the response body includes `"duplicates": []`

#### Scenario: Response includes empty array when phash is absent

- **WHEN** `POST /images/:id/complete` is called without a `phash` field
- **THEN** the response body includes `"duplicates": []`
- **AND** no duplicate query is executed

---

### Requirement: Periodic phash backfill

The system SHALL register a River periodic job (`BackfillPhashArgs`) that runs every 5 minutes. Each execution SHALL fetch up to 20 non-deleted images where `phash IS NULL`, download each image's thumbnail from R2 (falling back to the full `r2_path` if `thumbnail_path` is absent), compute a perceptual hash via `goimagehash.PerceptionHash`, and update the `phash` column. Images that cannot be fetched or decoded SHALL be skipped with a warning log; processing continues for the remaining batch.

#### Scenario: Backfill stores phash for un-hashed images

- **WHEN** the backfill job runs and images with `phash IS NULL` exist
- **THEN** each image in the batch receives a non-null `phash` value

#### Scenario: Backfill is a no-op when all images are hashed

- **WHEN** the backfill job runs and no images have `phash IS NULL`
- **THEN** no updates are made and the job completes without error

#### Scenario: Backfill skips images it cannot fetch

- **WHEN** an image's thumbnail cannot be retrieved from R2
- **THEN** that image's `phash` remains `NULL`
- **AND** a warning is logged
- **AND** remaining images in the batch are still processed
