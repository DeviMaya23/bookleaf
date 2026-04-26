## 1. Dependencies

- [ ] 1.1 Add `github.com/google/uuid` to go.mod (`go get github.com/google/uuid`)
- [ ] 1.2 Add `github.com/golang-migrate/migrate/v4` and the `postgres` driver (`go get github.com/golang-migrate/migrate/v4` + `go get github.com/golang-migrate/migrate/v4/database/postgres`)

## 2. Domain Structs

- [ ] 2.1 Write `internal/domain/user.go` — `User` struct with `ID` (UUID) and `ClerkID` (string, unique index)
- [ ] 2.2 Write `internal/domain/folder.go` — `Folder` struct with UUID PK, `UserID`, nullable `ParentID` (self-ref), `Name`
- [ ] 2.3 Write `internal/domain/image.go` — `Image` struct with UUID PK, `UserID`, nullable `FolderID`, `Title`, `SourceURL`, `R2Path`, `ThumbnailPath`, `MIMEType`, `VisionLabels` (JSON)

## 3. Migrations

- [ ] 3.1 Create `migration/` directory
- [ ] 3.2 Write `migration/000001_create_users.up.sql` — `users` table (id UUID PK, clerk_id unique, timestamps)
- [ ] 3.3 Write `migration/000001_create_users.down.sql` — `DROP TABLE users`
- [ ] 3.4 Write `migration/000002_create_folders.up.sql` — `folders` table with FK to users, nullable self-ref parent_id, `ON DELETE RESTRICT`
- [ ] 3.5 Write `migration/000002_create_folders.down.sql` — `DROP TABLE folders`
- [ ] 3.6 Write `migration/000003_create_images.up.sql` — `images` table with FK to users and nullable FK to folders, `vision_labels` as JSONB
- [ ] 3.7 Write `migration/000003_create_images.down.sql` — `DROP TABLE images`

## 4. Verify

- [ ] 4.1 `go build ./...` passes with no errors
