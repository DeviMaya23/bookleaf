# image-label-search

## Purpose

Defines the `search_labels` query parameter on `GET /images`, its handler-to-usecase-to-repository flow, and the SQL expansion that widens a title search to also match AI-generated labels above a confidence threshold.

## Requirements

### Requirement: GET /images search_labels query parameter

The `GET /images` handler SHALL accept an optional boolean `search_labels` query parameter. When present and set to `"true"`, and when a non-empty `name` parameter is also present, the handler SHALL pass `SearchLabels: true` to `ListImagesParams`. When `search_labels` is absent, empty, or any value other than `"true"`, `SearchLabels` SHALL be `false`.

When `SearchLabels` is `true` and `name` is a non-nil, non-empty string, the repository `List` method SHALL widen the title filter to also match images that have at least one label in `image_labels` whose value contains the search term, case-insensitively, **and whose `score` is at least `0.75`**. The resulting filter is:

```sql
(images.title ILIKE '%<term>%'
 OR EXISTS (
   SELECT 1 FROM image_labels
   WHERE image_id = images.id
   AND label ILIKE '%<term>%'
   AND score >= 0.75
 ))
```

The `0.75` threshold matches the existing `VISION_LABEL_SCORE_THRESHOLD` constant used elsewhere in the system. The repository SHALL define a local unexported constant for this value rather than importing from the agent package.

This OR EXISTS condition replaces the plain `images.title ILIKE` condition used when `SearchLabels` is `false`. All other active filters (`unfiled`, `folderIDs`, `tagIDs`, `mimeTypes`, sort, cursor) continue to compose with it via AND.

When `SearchLabels` is `true` but `name` is nil or empty, the label search has no effect — no `image_labels` subquery is added.

#### Scenario: search_labels=true widens search to include AI label matches

- **WHEN** `GET /images?name=sunset&search_labels=true` is called
- **THEN** the response includes images whose title contains "sunset" case-insensitively
- **AND** the response also includes images that have no title match but have at least one `image_labels` row whose `label` contains "sunset" case-insensitively and whose `score >= 0.75`
- **AND** images not matching either condition are excluded

#### Scenario: search_labels=false (or absent) limits search to title only

- **WHEN** `GET /images?name=sunset` is called without `search_labels`
- **THEN** only images whose title contains "sunset" case-insensitively are returned
- **AND** images whose labels contain "sunset" but whose title does not are excluded

#### Scenario: search_labels=true with no name term is a no-op

- **WHEN** `GET /images?search_labels=true` is called without a `name` parameter
- **THEN** results are identical to a request with neither parameter — no label filter is applied
