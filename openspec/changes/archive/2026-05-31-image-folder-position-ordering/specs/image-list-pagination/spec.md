## MODIFIED Requirements

### Requirement: GET /images Pagination Query Parameters

The `GET /images` handler SHALL accept:

| Parameter  | Type   | Default | Max | Description                                          |
|------------|--------|---------|-----|------------------------------------------------------|
| `limit`    | int    | 50      | 200 | Page size (silently clamped, not rejected)            |
| `cursor`   | string | —       | —   | Opaque cursor from a previous response               |
| `folder_id`| uuid   | —       | —   | When present, bypasses cursor/limit (returns all)    |

When `folder_id` is provided, `cursor` and `limit` are ignored entirely. The handler SHALL NOT attempt to parse a cursor and SHALL NOT apply any limit. All images in the folder are returned in a single response.

When `folder_id` is absent (all or unfiled views), cursor/limit behaviour is unchanged: an unparseable `cursor` value SHALL return `400 Bad Request`.

#### Scenario: Folder view ignores cursor and limit

- **WHEN** `GET /images?folder_id=<id>&cursor=<any>&limit=<any>` is called
- **THEN** all images in the folder are returned regardless of cursor or limit values
- **AND** `next_cursor` in the response is `null`

#### Scenario: Request with no pagination params uses defaults (non-folder view)

- **WHEN** `GET /images` is called with no `limit`, `cursor`, or `folder_id` params
- **THEN** up to 50 images are returned

#### Scenario: Request with explicit limit (non-folder view)

- **WHEN** `GET /images?limit=10` is called without `folder_id`
- **THEN** up to 10 images are returned

#### Scenario: Limit above 200 is silently clamped (non-folder view)

- **WHEN** `GET /images?limit=500` is called without `folder_id`
- **THEN** up to 200 images are returned and no error is returned

#### Scenario: Invalid cursor returns 400 (non-folder view)

- **WHEN** `GET /images?cursor=notvalidbase64!!!` is called without `folder_id`
- **THEN** the response is `400 Bad Request`

---

### Requirement: GET /images Response Envelope — folder view

When `folder_id` is provided, the `GET /images` response envelope SHALL always have `next_cursor: null`:

```json
{
  "images": [ /* all images in folder, ordered by position ASC */ ],
  "next_cursor": null
}
```

#### Scenario: Folder view always returns null next_cursor

- **WHEN** `GET /images?folder_id=<id>` is called regardless of how many images are in the folder
- **THEN** `next_cursor` is `null` in the response body
- **AND** `images` contains all non-deleted images in that folder ordered by `image_folders.position ASC`
