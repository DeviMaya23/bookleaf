## 1. Dependencies

- [x] 1.1 Add `github.com/google/uuid` to go.mod (`go get github.com/google/uuid`)
- [x] 1.2 Add `github.com/golang-migrate/migrate/v4` and the `postgres` driver (`go get github.com/golang-migrate/migrate/v4` + `go get github.com/golang-migrate/migrate/v4/database/postgres`)

## 2. Domain Structs

- [x] 2.1 Write `internal/domain/user.go` — `User` struct with `ID` (Clerk string ID, `TEXT` PK; no UUID field)
- [x] 2.2 Write `internal/domain/folder.go` — `Folder` struct with UUID PK, `UserID`, nullable `ParentID` (self-ref), `Name`
- [x] 2.3 Write `internal/domain/image.go` — `Image` struct with UUID PK, `UserID`, nullable `FolderID`, `Title`, `SourceURL`, `R2Path`, `ThumbnailPath`, `MIMEType`, `VisionLabels` (JSON)

## 3. Migrations

- [x] 3.1 Create `migration/` directory
- [x] 3.2 Write `migration/000001_create_users.up.sql` — `users` table (`id TEXT PRIMARY KEY` using Clerk's user ID, timestamps, `deleted_at`)
- [x] 3.3 Write `migration/000001_create_users.down.sql` — `DROP TABLE users`
- [x] 3.4 Write `migration/000002_create_folders.up.sql` — `folders` table with FK to users, nullable self-ref parent_id, `ON DELETE RESTRICT`
- [x] 3.5 Write `migration/000002_create_folders.down.sql` — `DROP TABLE folders`
- [x] 3.6 Write `migration/000003_create_images.up.sql` — `images` table with FK to users and nullable FK to folders, `ai_labels` as JSONB
- [x] 3.7 Write `migration/000003_create_images.down.sql` — `DROP TABLE images`

## 4. Verify

- [x] 4.1 `go build ./...` passes with no errors
