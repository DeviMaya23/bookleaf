## ADDED Requirements

### Requirement: Image PHash field

The `Image` GORM struct SHALL include a `PHash` nullable pointer field of type `*string`, mapped to a `phash` column of Postgres type `bit(64)`. The field represents the perceptual hash of the image content and is `NULL` for images that have not yet been hashed.

#### Scenario: Image struct compiles with PHash field

- **WHEN** the Go package is compiled
- **THEN** `Image` has a `PHash *string` field with GORM tags `column:phash` and `type:bit(64)`

---

### Requirement: phash DB migration

The system SHALL include a `golang-migrate` SQL migration that adds a nullable `phash bit(64)` column to the existing `images` table. Existing rows SHALL have `phash = NULL` after the migration.

#### Scenario: Migration adds the phash column

- **WHEN** the migration up is applied
- **THEN** the `images` table has a nullable `phash` column of type `bit(64)`
- **AND** all pre-existing rows have `phash = NULL`

#### Scenario: Migration is reversible

- **WHEN** the migration down is applied
- **THEN** the `phash` column is dropped from the `images` table without error
