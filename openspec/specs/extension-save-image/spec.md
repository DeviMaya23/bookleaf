# Spec: Extension Save Image

## Purpose

Defines the requirements for the "Save to Bookleaf" browser extension feature, which allows users to save images from any webpage directly to their Bookleaf library via a right-click context menu.

## Requirements

### Requirement: Context menu registration

The background service worker SHALL register a "Save to Bookleaf" context menu item on extension install and startup. The item SHALL appear only when right-clicking an `<img>` element (`contexts: ["image"]`).

#### Scenario: Context menu item appears on image right-click

- **WHEN** the user right-clicks an `<img>` element on any webpage
- **THEN** a "Save to Bookleaf" option appears in the browser context menu

#### Scenario: Context menu item does not appear on non-image right-click

- **WHEN** the user right-clicks text or a non-image element
- **THEN** "Save to Bookleaf" does not appear in the context menu

### Requirement: Authenticated save flow

When the user clicks "Save to Bookleaf", the extension SHALL:
1. Read the stored auth token from `chrome.storage.local`
2. If no token exists or the token is expired, show a "Please log in first" notification and abort
3. Fetch the image blob from `info.srcUrl` via the background service worker
4. Execute the 3-step upload sequence: `POST /images` → `PUT` blob to presigned R2 URL → `POST /images/:id/complete`
5. Show a "Saved to Bookleaf!" notification on success

The image title SHALL be set to the current tab's title (`tab.title`). The `source_url` SHALL be set to `info.pageUrl`. No `folder_id` SHALL be sent (image saves to root).

#### Scenario: Successful save shows success notification

- **WHEN** the user clicks "Save to Bookleaf" while authenticated and the image is fetchable
- **THEN** the image is uploaded to Bookleaf with the page title as its title and page URL as its source_url
- **AND** a Chrome notification with message "Saved to Bookleaf!" is shown

#### Scenario: Unauthenticated save is rejected

- **WHEN** the user clicks "Save to Bookleaf" and no valid token exists in `chrome.storage.local`
- **THEN** no upload is attempted
- **AND** a Chrome notification with message "Please log in first" is shown

#### Scenario: Expired token is rejected

- **WHEN** the user clicks "Save to Bookleaf" and the stored token's `expiresAt` is in the past
- **THEN** no upload is attempted
- **AND** a Chrome notification with message "Please log in first" is shown

### Requirement: Save failure notification

If any step in the save flow fails (image fetch error, API error, presigned PUT failure), the extension SHALL show a "Save failed. Please try again." notification. No retry is attempted automatically.

#### Scenario: Image fetch failure shows error notification

- **WHEN** the background SW fetch of the image URL returns a non-OK response or throws
- **THEN** a Chrome notification with message "Save failed. Please try again." is shown
- **AND** no upload is initiated

#### Scenario: Upload API failure shows error notification

- **WHEN** any step of the 3-step upload sequence returns a non-2xx response
- **THEN** a Chrome notification with message "Save failed. Please try again." is shown

### Requirement: API client helper

The extension SHALL expose an `apiFetch(path, options?)` function in `src/lib/api.ts` that:
- Reads the access token from `chrome.storage.local` via `getAuth()`
- Attaches `Authorization: Bearer <accessToken>` to every request
- Prepends `VITE_API_BASE_URL` to the path
- Returns the raw `Response`

#### Scenario: apiFetch attaches auth header

- **WHEN** `apiFetch('/images', { method: 'POST', body })` is called with a valid stored token
- **THEN** the outgoing request includes `Authorization: Bearer <accessToken>`
- **AND** the request is sent to `${VITE_API_BASE_URL}/images`
