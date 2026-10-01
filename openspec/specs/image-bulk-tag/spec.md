## Purpose

Defines the `POST /images/bulk/tag` endpoint for appending a set of tags to many images in a single atomic transaction, and the corresponding `bulkAddTags` frontend lib function.

## Requirements

### Requirement: POST /images/bulk/tag — Bulk Append Tags to Images

The system SHALL expose a `POST /images/bulk/tag` endpoint on the protected route group that appends a set of tags to many images in a single atomic transaction.

Request body:
```json
{
  "image_ids": ["uuid", ...],
  "tag_ids": ["uuid", ...]
}
```

- Every entry in `image_ids` and `tag_ids` MUST be a well-formed UUID. If any entry is malformed, the entire request SHALL fail with `400 Bad Request` and no rows SHALL be inserted.
- Every entry in `tag_ids` MUST exist and belong to the authenticated user. If any tag does not exist or belongs to another user, the entire request SHALL fail with `404 Not Found` and no rows SHALL be inserted.
- For each well-formed image ID that does not exist or does not belong to the authenticated user: it SHALL be silently ignored. It SHALL NOT cause a failure.
- For each valid `(image_id, tag_id)` pair where the image already has that tag: the existing row SHALL be left unchanged (`ON CONFLICT DO NOTHING`). This is not an error and does not reduce `succeeded_count`.
- All `(image_id, tag_id)` insertions SHALL occur within a single database transaction. If the transaction fails, no rows SHALL be inserted.
- `succeeded_count` in the response SHALL equal the number of `image_ids` that were owned by the authenticated user and processed (including idempotent no-ops on already-tagged images).
- On completion, the system SHALL return `200 OK` with `{"succeeded_count": <n>}`.

#### Scenario: All images tagged successfully

- **WHEN** a user submits three owned image IDs and two owned tag IDs
- **THEN** up to six `image_tags` rows are inserted and the response is `200 OK` with `{"succeeded_count": 3}`

#### Scenario: Image already has one of the submitted tags

- **WHEN** a user submits an image that already has tag A, along with tag A and tag B
- **THEN** the existing `(image_id, tag_A_id)` row is left unchanged, `(image_id, tag_B_id)` is inserted, no error occurs, and the image is counted in `succeeded_count`

#### Scenario: One image ID does not belong to the authenticated user

- **WHEN** a user submits `image_ids` where one ID belongs to a different user
- **THEN** that image is ignored, the remaining valid images are tagged normally, and `succeeded_count` reflects only the valid images

#### Scenario: A tag ID does not belong to the authenticated user

- **WHEN** a user submits a `tag_ids` entry that belongs to another user
- **THEN** the system returns `404 Not Found` and no `image_tags` rows are inserted

#### Scenario: image_ids contains a malformed UUID

- **WHEN** a user submits an `image_ids` entry that is not a valid UUID
- **THEN** the system returns `400 Bad Request` and no rows are inserted

#### Scenario: tag_ids contains a malformed UUID

- **WHEN** a user submits a `tag_ids` entry that is not a valid UUID
- **THEN** the system returns `400 Bad Request` and no rows are inserted

#### Scenario: All submitted image IDs are unowned

- **WHEN** every entry in `image_ids` fails ownership/existence validation
- **THEN** the system returns `200 OK` with `{"succeeded_count": 0}`

---

### Requirement: bulkAddTags frontend lib function

`src/lib/tags.ts` SHALL export a function:

```ts
export async function bulkAddTags(
  getToken: GetToken,
  imageIds: string[],
  tagIds: string[],
): Promise<BulkActionResult>
```

It SHALL POST to `/images/bulk/tag` with `{ image_ids: imageIds, tag_ids: tagIds }` and return the parsed JSON. It SHALL throw an `Error` if the response is not ok.

#### Scenario: Returns BulkActionResult on success

- **WHEN** the server responds with `200 OK`
- **THEN** `bulkAddTags` resolves with `{ succeeded_count: n }`

#### Scenario: Throws on non-ok response

- **WHEN** the server responds with a non-2xx status
- **THEN** `bulkAddTags` throws an `Error`
