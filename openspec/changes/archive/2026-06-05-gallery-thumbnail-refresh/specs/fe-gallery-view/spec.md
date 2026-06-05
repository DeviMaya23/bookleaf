## ADDED Requirements

### Requirement: Gallery self-polls while any image has a pending thumbnail
The system SHALL set `refetchInterval` on the gallery's `useInfiniteQuery` to 1000ms while any loaded image has `thumbnail_url === null`. Polling SHALL stop automatically (interval returns `false`) once all loaded images have a non-null `thumbnail_url`. This covers all upload paths without any upload-specific wiring.

#### Scenario: Polling active while pending thumbnails exist
- **WHEN** the gallery's loaded image list contains at least one image with `thumbnail_url === null`
- **THEN** the gallery refetches `GET /images` every 1 second

#### Scenario: Polling stops once all thumbnails resolve
- **WHEN** all loaded images have a non-null `thumbnail_url`
- **THEN** the gallery stops polling and makes no further periodic refetch requests
