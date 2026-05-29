## Context

Bookleaf currently organises images through a folder hierarchy. Tags introduce a second, flat organisational axis. No tag concept exists anywhere in the codebase — this is a greenfield addition. The existing image layer (domain struct, repository, usecase, handler) is stable and well-tested; changes to it must be additive and non-breaking.

## Goals / Non-Goals

**Goals:**
- User-scoped tag vocabulary (each user owns their own tags)
- Many-to-many image–tag relationship via a junction table
- Tag CRUD API (create, list, rename, delete)
- Tags included in image list and detail responses (via preload)
- Image update endpoint accepts a tag replacement payload
- Image list endpoint accepts a `tag_id` filter

**Non-Goals:**
- Global/shared tags across users
- Tag search or autocomplete endpoint (list-all is sufficient at this scale)
- Filtering by multiple tags simultaneously (single `tag_id` filter for v1)
- Tagging during the initial upload flow (tags applied via PATCH after upload)
- Tag-based sorting

## Decisions

### Tag names are case-sensitive, unique per user at the DB level

A `UNIQUE(user_id, name)` constraint on the `tags` table enforces uniqueness. No normalisation is applied in the usecase. Rationale: simpler to implement, and users who want case-insensitive deduplication can manage that expectation on the client. Adding case-insensitivity later (e.g. `LOWER(name)` unique index) is a non-breaking migration.

**Alternative considered:** normalise to lowercase in the usecase before insert. Rejected because it silently transforms user input, which is surprising.

### Replace-all semantics for UpdateImage tags

When `PATCH /images/:id` includes a `tags` field, the image's entire tag set is replaced with the supplied list. Omitting `tags` leaves the current associations untouched (consistent with how other nullable fields like `folder_id` work).

The field is encoded as `json.RawMessage` so the handler can distinguish `null`/omitted from `[]` (clear all tags) vs `["uuid1"]` (set to this list) — the same pattern used for `folder_id` and `source_url` today.

**Alternative considered:** separate `tags_add` / `tags_remove` fields. Rejected for v1 — replace-all is simpler and the tag list per image is expected to be small.

### Tag–image association is managed by a TagRepository method, not the image Update path

`UpdateImage` already uses a `map[string]any` with GORM's `Updates()` for scalar field changes. GORM associations cannot go through that path. A dedicated `ReplaceImageTags(ctx, imageID, tagIDs []uuid.UUID) error` method on `TagRepository` handles the junction table atomically (delete existing rows, insert new ones in a single transaction). The image usecase calls this method when `UpdateImageParams.Tags` is non-nil.

**Alternative considered:** add the tag replacement to `ImageRepository`. Rejected to keep tag-related data operations co-located in `TagRepository`.

### Tags are preloaded via GORM Preload, not a JOIN

`Preload("Tags")` on `List` and `GetByID` fires one batched `IN (...)` query after the primary image fetch — always 2 queries regardless of page size. This avoids N+1 without restructuring the existing query logic.

**Alternative considered:** raw JOIN with aggregation. Rejected as more complex and harder to maintain alongside the existing cursor-based pagination.

### image_tags uses ON DELETE CASCADE on both FKs

- `image_id → images.id ON DELETE CASCADE`: when an image is hard-deleted (stale upload cleanup, trash purge), junction rows clean up automatically — no changes to the purge flows.
- `tag_id → tags.id ON DELETE CASCADE`: deleting a tag removes all its image associations; images are unaffected.

Soft-deleted images retain their tag associations in the junction table, so tags are restored alongside the image if it is un-trashed.

## Risks / Trade-offs

- **Preload adds a second DB round-trip to every image list call** → Acceptable at current scale. If it becomes a bottleneck, a JOIN-based approach can replace it without changing the interface.
- **Replace-all requires the client to know all current tags** → The client already has tags from the detail/list response, so this is safe in practice.
- **No migration for existing images** → Existing images simply have no tags; `image_tags` starts empty. No data migration needed.

## Migration Plan

1. Deploy migration `000009_create_tags` (creates `tags` and `image_tags` tables)
2. Deploy updated server binary
3. Rollback: run down migration — drops both tables; no data loss to existing images/folders
