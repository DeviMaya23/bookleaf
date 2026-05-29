## Purpose

Defines the Tag domain model, database migration, and repository interface for user-owned labels that can be associated with images via a many-to-many relationship.

## Requirements

### Requirement: Tag GORM Struct

The system SHALL define a `Tag` GORM struct in `internal/domain/tag.go` representing a user-owned label.

Fields:
- `ID` — UUID primary key (`id`)
- `UserID` — FK to users table; Kinde user ID string, not null (`user_id`)
- `Name` — display name, not null (`name`)
- `CreatedAt`, `UpdatedAt` — GORM timestamps

Associations:
- `User User` — belongs-to, `foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT`
- `Images []Image` — many-to-many via `image_tags` join table, `foreignKey:TagID;joinForeignKey:TagID;References:ID;joinReferences:ImageID`

`BeforeCreate` SHALL assign a new UUID if `ID` is nil.

#### Scenario: Tag struct compiles with GORM tags

- **WHEN** the Go package is compiled
- **THEN** `Tag` has a `gorm:"primaryKey"` UUID field, a not-null `user_id` FK, and a not-null `name` column

#### Scenario: Tag auto-assigns UUID on create

- **WHEN** a `Tag` is created with a nil `ID`
- **THEN** `BeforeCreate` assigns a new non-nil UUID before the INSERT

### Requirement: Tags DB Migration

The system SHALL include a `golang-migrate` SQL migration `000009_create_tags` that creates the `tags` table and the `image_tags` junction table.

`tags` table columns:
- `id` UUID PRIMARY KEY
- `user_id` TEXT NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT
- `name` TEXT NOT NULL
- `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
- `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
- UNIQUE constraint on `(user_id, name)`

`image_tags` table columns:
- `image_id` UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE
- `tag_id` UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE
- PRIMARY KEY `(image_id, tag_id)`

#### Scenario: Migration creates tags table with unique constraint

- **WHEN** the up migration is applied to a fresh database
- **THEN** the `tags` table exists with a `UNIQUE(user_id, name)` constraint

#### Scenario: Migration creates image_tags junction table

- **WHEN** the up migration is applied
- **THEN** the `image_tags` table exists with a composite primary key on `(image_id, tag_id)` and CASCADE delete on both FK columns

#### Scenario: Migration is reversible

- **WHEN** the down migration is applied
- **THEN** both `image_tags` and `tags` tables are dropped without error

#### Scenario: Deleting a tag cascades to image_tags

- **GIVEN** a tag is associated with one or more images
- **WHEN** the tag row is deleted
- **THEN** all corresponding rows in `image_tags` are deleted automatically
- **AND** the images themselves are unaffected

#### Scenario: Hard-deleting an image cascades to image_tags

- **GIVEN** an image has one or more tag associations
- **WHEN** the image row is hard-deleted
- **THEN** all corresponding rows in `image_tags` are deleted automatically

### Requirement: TagRepository Interface

The system SHALL define a `TagRepository` interface in `internal/usecase/tag_repository.go`.

Methods:
- `Create(ctx context.Context, tag *domain.Tag) (*domain.Tag, error)`
- `ListByUserID(ctx context.Context, userID string) ([]*domain.Tag, error)` — returns all non-deleted tags for the user ordered by `name ASC`
- `GetByID(ctx context.Context, id uuid.UUID, userID string) (*domain.Tag, error)`
- `Update(ctx context.Context, id uuid.UUID, userID string, name string) (*domain.Tag, error)` — renames the tag; returns updated record
- `Delete(ctx context.Context, id uuid.UUID, userID string) error`
- `ReplaceImageTags(ctx context.Context, imageID uuid.UUID, tagIDs []uuid.UUID) error` — within a single transaction, deletes all existing rows in `image_tags` for the given `imageID`, then inserts rows for each `tagID` in `tagIDs`; a nil or empty slice clears all associations

#### Scenario: Repository interface is satisfied by SQL implementation

- **WHEN** the Go package is compiled
- **THEN** `tagRepository` in `internal/repository/` implements `usecase.TagRepository` without compilation errors

#### Scenario: ReplaceImageTags clears associations when given empty slice

- **WHEN** `ReplaceImageTags` is called with an empty `tagIDs` slice
- **THEN** all rows in `image_tags` for that `imageID` are deleted
- **AND** no new rows are inserted

#### Scenario: ReplaceImageTags is atomic

- **WHEN** `ReplaceImageTags` is called with valid tag IDs
- **THEN** the delete and insert happen within a single database transaction
- **AND** if the insert fails, the delete is rolled back
