## ADDED Requirements

### Requirement: ImageLabel domain type

The system SHALL define an `ImageLabel` struct in `internal/domain/`:

```go
type ImageLabel struct {
    ID      uuid.UUID `gorm:"type:uuid;primaryKey"`
    ImageID uuid.UUID `gorm:"column:image_id;not null;index"`
    Label   string    `gorm:"column:label;not null"`
    Score   float32   `gorm:"column:score;not null"`
}
```

`ImageLabel` SHALL have a `BeforeCreate` hook that generates a UUID if `ID` is nil, consistent with other domain types.

#### Scenario: BeforeCreate generates ID when not set

- **WHEN** an `ImageLabel` is created without an explicit ID
- **THEN** `BeforeCreate` assigns a new UUID before the insert

---

### Requirement: image_labels migration

The system SHALL add a numbered SQL migration that:

1. Creates the `image_labels` table:

```sql
CREATE TABLE image_labels (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    image_id   UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    label      TEXT NOT NULL,
    score      FLOAT4 NOT NULL
);

CREATE INDEX ON image_labels (image_id);
```

2. Backfills existing data from `images.ai_labels`:

```sql
INSERT INTO image_labels (image_id, label, score)
SELECT images.id,
       elem->>'Description',
       (elem->>'Score')::float4
FROM images,
     LATERAL jsonb_array_elements(ai_labels) AS elem
WHERE ai_labels IS NOT NULL
  AND jsonb_typeof(ai_labels) = 'array';
```

The `jsonb_typeof` guard ensures rows where `ai_labels` is a non-array JSON value are skipped.

The down migration SHALL drop the `image_labels` table.

#### Scenario: Migration creates table and index

- **WHEN** the up migration is applied
- **THEN** `image_labels` table exists with a foreign key to `images` and an index on `image_id`

#### Scenario: Backfill populates rows from existing ai_labels

- **WHEN** the up migration is applied and some images have `ai_labels` populated
- **THEN** `image_labels` contains one row per label per image matching the JSONB data

#### Scenario: Images with null ai_labels are skipped

- **WHEN** the up migration is applied and some images have `ai_labels IS NULL`
- **THEN** no `image_labels` rows are created for those images

---

### Requirement: UpdateLabels repository method

The system SHALL add `UpdateLabels(ctx context.Context, id uuid.UUID, rawJSON json.RawMessage, labels []domain.ImageLabel) error` to the `UploadImageRepository` interface and implement it on `imageRepository`.

The method SHALL atomically (within a single DB transaction):
1. Update `images.ai_labels` to `rawJSON` for the given image ID
2. Delete any existing rows in `image_labels` for that `image_id`
3. Bulk-insert all rows in `labels` into `image_labels`

`UpdateAILabels` SHALL be removed from the `UploadImageRepository` interface; `UpdateLabels` replaces it.

#### Scenario: Both writes succeed atomically

- **WHEN** `UpdateLabels` is called with valid data
- **THEN** `images.ai_labels` is updated and `image_labels` rows are inserted in the same transaction

#### Scenario: Failure rolls back both writes

- **WHEN** an error occurs during either write inside the transaction
- **THEN** neither the `ai_labels` update nor any `image_labels` rows are committed

#### Scenario: Zero labels clears image_labels rows and sets empty array

- **WHEN** `UpdateLabels` is called with an empty `labels` slice and `rawJSON` of `[]`
- **THEN** `ai_labels` is set to `[]` and any existing `image_labels` rows for that image are deleted
