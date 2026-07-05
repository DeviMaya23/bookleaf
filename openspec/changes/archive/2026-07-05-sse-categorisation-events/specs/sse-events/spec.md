## ADDED Requirements

### Requirement: GET /events SSE Endpoint

The system SHALL expose a `GET /events` endpoint in the protected route group that establishes a long-lived Server-Sent Events stream. The endpoint SHALL keep the connection open and push events to the client as they occur. Each event SHALL be formatted as an SSE `data:` line containing a JSON object with at minimum a `type` string field.

The endpoint SHALL set the following response headers:
- `Content-Type: text/event-stream`
- `Cache-Control: no-cache`
- `Connection: keep-alive`

The connection SHALL be closed cleanly when the client disconnects.

#### Scenario: Authenticated client establishes SSE connection

- **WHEN** an authenticated request is made to `GET /events`
- **THEN** the response status is `200 OK`
- **AND** the `Content-Type` header is `text/event-stream`
- **AND** the connection remains open

#### Scenario: Unauthenticated request is rejected

- **WHEN** a request is made to `GET /events` without a valid Bearer token
- **THEN** the response is `401 Unauthorized`
- **AND** no SSE stream is established

### Requirement: categorisation_complete Event

When an AI categorisation job completes successfully for an image, the system SHALL publish a `categorisation_complete` event to all open SSE connections belonging to the affected user.

Event wire format:
```json
{ "type": "categorisation_complete", "payload": { "image_id": "<uuid>" } }
```

- `type` — the event type discriminator
- `payload.image_id` — the UUID of the image that was categorised

The event SHALL NOT be published if the categorisation job fails or is retried.

#### Scenario: Event delivered after successful categorisation

- **WHEN** the categorisation job completes successfully for image X belonging to user U
- **THEN** all open SSE connections for user U receive a `categorisation_complete` event with `image_id` set to X's UUID

#### Scenario: No event delivered on categorisation failure

- **WHEN** the categorisation job returns an error for image X
- **THEN** no `categorisation_complete` event is published for any connection

#### Scenario: Event fan-out across multiple tabs

- **WHEN** user U has two open SSE connections (two browser tabs)
- **AND** a categorisation job completes successfully for an image belonging to user U
- **THEN** both connections receive the `categorisation_complete` event

### Requirement: FE SSE Connection Lifecycle

The frontend SHALL open an SSE connection to `GET /events` using `@microsoft/fetch-event-source`, sending the authenticated user's Bearer token in the `Authorization` header. The connection SHALL be opened only when the authenticated user has `ai_categorisation_enabled: true`. The connection SHALL be closed when the component that owns it unmounts.

#### Scenario: Connection opened for user with ai_categorisation_enabled true

- **WHEN** the authenticated user has `ai_categorisation_enabled: true`
- **THEN** the frontend opens an SSE connection to `GET /events`

#### Scenario: Connection not opened for user with ai_categorisation_enabled false

- **WHEN** the authenticated user has `ai_categorisation_enabled: false`
- **THEN** the frontend does not open an SSE connection

### Requirement: FE Cache Invalidation on categorisation_complete

Upon receiving a `categorisation_complete` event, the frontend SHALL invalidate the following React Query caches to reflect the updated folder assignment:

- `['folders']` — folder list (sidebar), as a new folder may have been created
- `['images']` — gallery image list, as the image's `folder_ids` changed
- `['image', imageId]` — the specific image detail, if the right panel is open
- `['folder']` (prefix) — all cached folder detail entries, as the assigned folder's `image_count` changed

#### Scenario: Caches invalidated on categorisation_complete

- **WHEN** the frontend receives a `categorisation_complete` event with `image_id` X
- **THEN** the `['folders']`, `['images']`, `['image', X]`, and `['folder']`-prefixed queries are invalidated
- **AND** React Query refetches any of those queries that are currently active
