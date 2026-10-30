# Spec Delta

## ADDED Requirements

### Requirement: Thumbnail GET URLs are deterministic

All endpoints that return a `thumbnail_url` field SHALL generate it via `GenerateDeterministicPresignedGetURL`. The URL SHALL be stable for the full 24 h UTC window so the browser can cache the image across repeated API calls within the same window.

#### Scenario: Same thumbnail URL is returned on consecutive calls within a window

- **WHEN** any endpoint that returns `thumbnail_url` is called twice for the same image within the same 24 h UTC window
- **THEN** both responses contain the identical `thumbnail_url` value

#### Scenario: Endpoints affected

- **WHEN** any of the following endpoints responds with image data that includes a thumbnail
- **THEN** the `thumbnail_url` is produced by `GenerateDeterministicPresignedGetURL`
- **AND** the affected endpoints are: list images, get image detail, list folder images, list trash items, get trash item, list shared folder images, get shared image detail
