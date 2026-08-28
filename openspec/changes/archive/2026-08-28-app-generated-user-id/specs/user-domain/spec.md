## MODIFIED Requirements

### Requirement: User GORM Struct

The system SHALL define a `User` GORM struct in `internal/domain/user.go` representing an authenticated user with an app-generated UUID primary key and a separate field for the IdP-issued subject.

Fields (all DB columns use snake_case):
- `ID` — app-generated UUID, `UUID` primary key (`id`); e.g. `550e8400-e29b-41d4-a716-446655440000`
- `IDPSubject` — IdP-issued subject string, `TEXT NOT NULL UNIQUE` (`idp_subject`); e.g. `kp_abc123`
- `VisionEnabled` — boolean flag indicating whether the user has opted into AI organising (`vision_enabled`); defaults to `false`
- `AICategorisationEnabled` — boolean flag indicating whether AI image categorisation is enabled for the user (`ai_categorisation_enabled`); defaults to `false`
- `AccountState` — account lifecycle state (`account_state`); defaults to `'active'`
- `PurgedAt` — nullable timestamp set when the account is purged (`purged_at`)
- `FolderIconsEnabled` — boolean flag indicating whether folder/system-entry icons are displayed in the sidebar (`folder_icons_enabled`); defaults to `true`
- `CreatedAt`, `UpdatedAt` — GORM timestamps (`created_at`, `updated_at`)
- `DeletedAt` — GORM soft-delete timestamp (nullable) (`deleted_at`)

#### Scenario: User struct uses app-generated UUID as primary key

- **WHEN** the Go package is compiled
- **THEN** `User` has a `uuid.UUID` `ID` field tagged `gorm:"type:uuid;primaryKey"`
- **AND** there is no string field holding the Kinde subject as the primary key

#### Scenario: User struct has IDPSubject field

- **WHEN** the Go package is compiled
- **THEN** `User` has a `string` `IDPSubject` field tagged `gorm:"column:idp_subject;not null;uniqueIndex"`
