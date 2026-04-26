## 1. Domain Model

- [ ] 1.1 Add `VisionEnabled bool` field to `User` struct in `internal/domain/user.go` with tag `gorm:"column:vision_enabled;default:false"`

## 2. Migration

- [ ] 2.1 Create `migration/000004_add_vision_enabled_to_users.up.sql` with `ALTER TABLE users ADD COLUMN vision_enabled BOOLEAN NOT NULL DEFAULT false`
- [ ] 2.2 Create `migration/000004_add_vision_enabled_to_users.down.sql` with `ALTER TABLE users DROP COLUMN vision_enabled`
