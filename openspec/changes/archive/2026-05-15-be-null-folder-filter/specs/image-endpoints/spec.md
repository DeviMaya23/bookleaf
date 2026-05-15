## ADDED Requirements

### Requirement: GET /images unfiled query parameter
The `GET /images` handler SHALL accept an optional `unfiled` boolean query parameter.

| `unfiled` value | Behaviour |
|---|---|
| Absent or `false` | Existing behaviour — no unfoldered filter applied |
| `true` | Returns only images where `folder_id IS NULL`; `folder_id` param is ignored |

`ListImagesParams` SHALL include an `Unfiled bool` field. When `Unfiled = true`, the repository SHALL emit `WHERE folder_id IS NULL` and ignore `FolderID`.

#### Scenario: unfiled=true returns only unfoldered images
- **WHEN** `GET /images?unfiled=true` is called
- **THEN** only images where `folder_id IS NULL` are returned

#### Scenario: unfiled=true ignores folder_id param
- **WHEN** `GET /images?unfiled=true&folder_id=<valid-uuid>` is called
- **THEN** only images where `folder_id IS NULL` are returned
- **AND** the `folder_id` param is not applied as a filter

#### Scenario: unfiled absent or false preserves existing behaviour
- **WHEN** `GET /images` is called without `unfiled` or with `unfiled=false`
- **THEN** existing folder filtering behaviour applies unchanged
