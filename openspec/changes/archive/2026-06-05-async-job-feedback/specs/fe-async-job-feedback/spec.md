## ADDED Requirements

### Requirement: Post-upload thumbnail polling

After `POST /images/:id/complete` succeeds — regardless of upload entry point (modal or drag-and-drop) — the system SHALL poll `GET /images/:id` every 2 seconds for up to 30 seconds until `thumbnail_url` is non-null. When `thumbnail_url` becomes non-null, the system SHALL invalidate the active React Query images cache (triggering a refetch) so the card updates live without a manual reload. Polling SHALL stop as soon as `thumbnail_url` is resolved or 30 seconds have elapsed, whichever comes first.

#### Scenario: Thumbnail resolves before timeout

- **WHEN** `GET /images/:id` returns a non-null `thumbnail_url` during the polling window
- **THEN** the React Query images cache is invalidated, triggering a refetch
- **AND** the image card transitions from placeholder to the real thumbnail without a page reload
- **AND** polling stops

#### Scenario: Thumbnail does not resolve within 30 seconds

- **WHEN** 30 seconds elapse without `thumbnail_url` becoming non-null
- **THEN** polling stops silently
- **AND** no error is shown to the user

---

### Requirement: Post-upload vision suggestion check

After `POST /images/:id/complete` succeeds — regardless of upload entry point (modal or drag-and-drop) — if the authenticated user has `vision_enabled: true` (from `GET /me`), the system SHALL schedule a single call to `GET /images/:id` after a 2–3 second delay. The check SHALL be made only once.

- If the response contains a non-null `suggested_folder_name`, the system SHALL show a suggestion toast.
- If `suggested_folder_name` is null (job not yet complete or no labels found), the system SHALL show a "Couldn't get folder suggestion" error toast.
- If the user has `vision_enabled: false`, no check is made and no toast is shown.

#### Scenario: Vision suggestion available at check time

- **WHEN** the single check fires and `GET /images/:id` returns a non-null `suggested_folder_name`
- **THEN** a suggestion toast is displayed with the folder name and Accept / Ignore actions

#### Scenario: Vision suggestion not available at check time

- **WHEN** the single check fires and `GET /images/:id` returns a null `suggested_folder_name`
- **THEN** a "Couldn't get folder suggestion" error toast is displayed

#### Scenario: Vision disabled — no check performed

- **WHEN** the user has `vision_enabled: false`
- **THEN** no vision check is scheduled after upload
- **AND** no vision-related toast is shown

---

### Requirement: Vision suggestion toast

When a `suggested_folder_name` is returned by the vision check, the system SHALL display a Sonner toast with:
- A message showing the suggested folder name
- An **Accept** action button that calls `POST /images/:id/accept-suggestion` with the suggested folder name and invalidates the `images` and `folders` query caches on success
- An **Ignore** action button that dismisses the toast without making any API call

#### Scenario: Accepting the suggestion assigns the folder

- **WHEN** the user clicks Accept on the suggestion toast
- **THEN** `POST /images/:id/accept-suggestion` is called with the suggested folder name
- **AND** the `images` and `folders` query caches are invalidated on success

#### Scenario: Ignoring the suggestion dismisses the toast

- **WHEN** the user clicks Ignore on the suggestion toast
- **THEN** the toast is dismissed
- **AND** no API call is made
