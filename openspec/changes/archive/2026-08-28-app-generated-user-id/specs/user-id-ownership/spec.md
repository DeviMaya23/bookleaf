## ADDED Requirements

### Requirement: idp_subject column on users table

The `users` table SHALL have an `idp_subject TEXT NOT NULL UNIQUE` column storing the IdP-issued subject (Kinde subject). It is the bridge between Kinde's identity space and Bookleaf's internal identity space. It SHALL NOT be used as a foreign key by any other table.

#### Scenario: idp_subject is unique and non-null

- **WHEN** the migration is applied
- **THEN** the `users` table has an `idp_subject` column of type `TEXT NOT NULL`
- **AND** a unique constraint exists on `idp_subject`

#### Scenario: Existing users have idp_subject backfilled

- **WHEN** the migration runs on a database with existing users whose `id` was previously the Kinde subject
- **THEN** every existing user row has `idp_subject` set to the value that `id` held before the migration

---

### Requirement: UserRepository.GetByIDPSubject

`UserRepository` SHALL expose:

```go
GetByIDPSubject(ctx context.Context, idpSubject string) (*domain.User, error)
```

It SHALL query `users` by `idp_subject = $1` (excluding soft-deleted rows). If no row matches it SHALL return `ErrUserNotFound`. If found it SHALL return the full `*domain.User`.

#### Scenario: Returns user when idp_subject matches

- **WHEN** `GetByIDPSubject` is called with a Kinde subject that exists in the database
- **THEN** it returns the corresponding `*domain.User` and no error

#### Scenario: Returns ErrUserNotFound when no match

- **WHEN** `GetByIDPSubject` is called with a Kinde subject that does not exist in the database
- **THEN** it returns `nil` and `ErrUserNotFound`

---

### Requirement: GetOrCreate generates app-owned UUID

`UserRepository.GetOrCreate(ctx, idpSubject string)` SHALL look up by `idp_subject`. If no row exists it SHALL insert a new row with `id = gen_random_uuid()` and `idp_subject = $1` using `INSERT ... ON CONFLICT (idp_subject) DO NOTHING`. It SHALL return the resulting `*domain.User` (existing or newly created).

#### Scenario: New user gets an app-generated UUID

- **WHEN** `GetOrCreate` is called with an idp_subject not yet in the database
- **THEN** a new `users` row is created with a UUID `id` (not the idp_subject value) and `idp_subject` set to the given value

#### Scenario: Existing user is returned unchanged

- **WHEN** `GetOrCreate` is called with an idp_subject that already exists
- **THEN** the existing user row is returned with its original `id`
- **AND** no duplicate row is inserted

---

### Requirement: Data Migration

The system SHALL include a `golang-migrate` SQL migration that, within a single transaction:

1. Adds `idp_subject TEXT`, backfills it from `id`, adds `NOT NULL` and `UNIQUE` constraints
2. Adds a `new_id UUID DEFAULT gen_random_uuid()` column to `users`
3. Adds `new_user_id UUID` staging columns to `folders`, `images`, `tags`, `pending_uploads`, and `ai_categorisation_logs`
4. Backfills each staging column via `UPDATE ... FROM users WHERE fk_table.user_id = users.id`
5. Drops existing FK constraints referencing `users.id` on all FK tables
6. Drops old `user_id` columns, renames `new_user_id` → `user_id`, adds FK constraints back (now referencing the new UUID primary key)
7. Replaces `users.id`: drops the old TEXT primary key, renames `new_id` → `id`, adds UUID primary key constraint

The down migration is intentionally not reversible (a DB snapshot serves as the rollback path).

#### Scenario: Migration produces UUID primary key on users

- **WHEN** the up migration is applied
- **THEN** `users.id` is of type `UUID` and is the primary key
- **AND** `users.idp_subject` is `TEXT NOT NULL UNIQUE`

#### Scenario: FK tables use UUID user_id after migration

- **WHEN** the up migration is applied
- **THEN** `folders.user_id`, `images.user_id`, `tags.user_id`, `pending_uploads.user_id`, and `ai_categorisation_logs.user_id` are all of type `UUID`
- **AND** the FK constraints on `folders`, `images`, `tags`, and `pending_uploads` reference `users(id)` correctly
