## Purpose

Defines the `ai_categorisation_logs` table and `CategorisationLogRepository` — a persistent record of every AI categorisation run, capturing the agent's reasoning and outcome for future user-facing display.

---

## Requirements

### Requirement: ai_categorisation_logs Table

The system SHALL include a `golang-migrate` SQL migration (`000020_create_ai_categorisation_logs`) that creates the `ai_categorisation_logs` table.

```sql
CREATE TABLE ai_categorisation_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    image_id        UUID REFERENCES images(id) ON DELETE SET NULL,
    user_id         TEXT NOT NULL,
    reasoning       TEXT NOT NULL,
    folder_id       UUID,
    new_folder_name TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- `image_id` is nullable (via `ON DELETE SET NULL`) so log rows survive image deletion
- Exactly one of `folder_id` or `new_folder_name` will be set per row (the other will be NULL), reflecting which branch the agent took
- Down migration SHALL drop the table

#### Scenario: Migration creates table with correct schema

- **WHEN** the up migration is applied to a fresh database
- **THEN** the `ai_categorisation_logs` table exists with all columns as specified
- **AND** `image_id` is nullable and references `images(id)` with `ON DELETE SET NULL`

#### Scenario: Migration is reversible

- **WHEN** the down migration is applied
- **THEN** the `ai_categorisation_logs` table is dropped without error

#### Scenario: Log row survives image deletion

- **WHEN** an image referenced by a log row is deleted
- **THEN** the log row remains and `image_id` is set to NULL

---

### Requirement: CategorisationLog Domain Type

The system SHALL define a `CategorisationLog` struct in `internal/domain/`:

```go
type CategorisationLog struct {
    ID             uuid.UUID
    ImageID        *uuid.UUID
    UserID         string
    Reasoning      string
    FolderID       *uuid.UUID
    NewFolderName  *string
    CreatedAt      time.Time
}
```

#### Scenario: Domain type compiles

- **WHEN** the Go package is compiled
- **THEN** `CategorisationLog` is defined in `internal/domain/` without compilation errors

---

### Requirement: CategorisationLogRepository

The system SHALL define a `categorisationLogRepository` interface in `internal/usecase/categorisation_usecase.go` with two methods:

```go
type categorisationLogRepository interface {
    Create(ctx context.Context, log *domain.CategorisationLog) error
    GetByImageID(ctx context.Context, imageID uuid.UUID) (*domain.CategorisationLog, error)
}
```

A SQL implementation SHALL be added in `internal/repository/`. `GetByImageID` SHALL return the most recent log entry for the given image, or `nil, nil` if none exists.

#### Scenario: Log entry is persisted

- **WHEN** `CategorisationLogRepository.Create` is called with a valid `CategorisationLog`
- **THEN** a row is inserted into `ai_categorisation_logs` with the correct field values

#### Scenario: Existing log entry is found by image ID

- **WHEN** `GetByImageID` is called for an image that has a prior log entry
- **THEN** the most recent `CategorisationLog` for that image is returned

#### Scenario: No log entry returns nil

- **WHEN** `GetByImageID` is called for an image with no prior log entry
- **THEN** `nil, nil` is returned

#### Scenario: Repository error is propagated

- **WHEN** the database returns an error during insert or query
- **THEN** the method returns a non-nil error
