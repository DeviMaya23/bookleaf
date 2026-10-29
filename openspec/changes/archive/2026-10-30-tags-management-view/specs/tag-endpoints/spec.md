# Spec Delta

## ADDED Requirements

### Requirement: GET /tags supports optional image count

The system SHALL accept an optional `include_count` query parameter on `GET /tags`. When `include_count=true` is passed, each tag in the response SHALL include an `image_count` field containing the number of images the tag is applied to for that user. When the parameter is absent or any other value, the response format is unchanged and no count query is performed.

#### Scenario: GET /tags without include_count returns no count

- **WHEN** the client calls `GET /tags` without any query parameters
- **THEN** the response is a JSON array where each object has `id`, `name`, `created_at`, and `updated_at` but no `image_count` field

#### Scenario: GET /tags with include_count=true returns image counts

- **WHEN** the client calls `GET /tags?include_count=true`
- **THEN** each tag object in the response includes an `image_count` field with the number of images that tag is applied to
