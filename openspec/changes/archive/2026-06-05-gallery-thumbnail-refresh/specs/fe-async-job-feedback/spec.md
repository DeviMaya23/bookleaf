## REMOVED Requirements

### Requirement: Post-upload thumbnail polling
**Reason**: Thumbnail refresh is now owned by the gallery via `refetchInterval` on `useInfiniteQuery`, covering all upload paths including batch. Upload-path-specific polling is no longer needed.
**Migration**: No migration required. The gallery self-polls automatically while any loaded image has `thumbnail_url === null`.

---

## MODIFIED Requirements

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
