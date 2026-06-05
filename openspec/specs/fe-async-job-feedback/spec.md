## Purpose

After a successful single-file image upload, the frontend provides live feedback for async background jobs: vision-based folder suggestion delivery. Thumbnail refresh is handled by the gallery's self-polling mechanism.

---

## Requirements

### Requirement: Post-upload vision suggestion check

After a successful single-file upload via modal or drag-and-drop — if the authenticated user has `vision_enabled: true` (from `GET /me`) — the system SHALL attempt to retrieve `suggested_folder_name` from `GET /images/:id` using a two-attempt retry: first at 1 second, then again 2 seconds later if the first attempt returns null (3 seconds total from upload). Batch uploads are explicitly excluded from this check.

- If either attempt returns a non-null `suggested_folder_name`, the system SHALL show a suggestion toast.
- If both attempts return null, the system SHALL show a "Couldn't get folder suggestion" error toast.
- If the request throws on either attempt, the system SHALL show a "Couldn't get folder suggestion" error toast.
- If the user has `vision_enabled: false`, no check is made and no toast is shown.

#### Scenario: Vision suggestion available on first attempt

- **WHEN** the first check at 1 second returns a non-null `suggested_folder_name`
- **THEN** a suggestion toast is displayed with the folder name and Accept / Ignore actions
- **AND** no second check is made

#### Scenario: Vision suggestion available on second attempt

- **WHEN** the first check at 1 second returns null
- **AND** the second check at 3 seconds returns a non-null `suggested_folder_name`
- **THEN** a suggestion toast is displayed with the folder name and Accept / Ignore actions

#### Scenario: Vision suggestion not available after both attempts

- **WHEN** both the 1-second and 3-second checks return null `suggested_folder_name`
- **THEN** a "Couldn't get folder suggestion" error toast is displayed

#### Scenario: Vision check request fails

- **WHEN** either attempt to call `GET /images/:id` throws an error
- **THEN** a "Couldn't get folder suggestion" error toast is displayed

#### Scenario: Vision disabled — no check performed

- **WHEN** the user has `vision_enabled: false`
- **THEN** no vision check is scheduled after upload
- **AND** no vision-related toast is shown

#### Scenario: Batch upload — no check performed

- **WHEN** a batch upload completes successfully
- **THEN** no vision check is scheduled
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
