## MODIFIED Requirements

### Requirement: FE Cache Invalidation on categorisation_complete

Upon receiving a `categorisation_complete` event, the frontend SHALL invalidate the following React Query caches to reflect the updated folder assignment and usage count:

- `['folders']` — folder list (sidebar), as a new folder may have been created
- `['images']` — gallery image list, as the image's `folder_ids` changed
- `['image', imageId]` — the specific image detail, if the right panel is open
- `['folder']` (prefix) — all cached folder detail entries, as the assigned folder's `image_count` changed
- `['me']` — user profile, as `ai_categorisation_count_this_month` has incremented

#### Scenario: Caches invalidated on categorisation_complete

- **WHEN** the frontend receives a `categorisation_complete` event with `image_id` X
- **THEN** the `['folders']`, `['images']`, `['image', X]`, `['folder']`-prefixed, and `['me']` queries are invalidated
- **AND** React Query refetches any of those queries that are currently active
