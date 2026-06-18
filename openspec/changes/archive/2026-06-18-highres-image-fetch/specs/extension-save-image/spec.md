## MODIFIED Requirements

### Requirement: Authenticated save flow

When the user clicks "Save to Bookleaf", the extension SHALL:
1. Read the stored auth token from `chrome.storage.local`
2. If no token exists or the token is expired, send a toast message to the active tab with title "Bookleaf" and body "Please log in first." and abort
3. Resolve the image to fetch:
   a. Call `resolveHighResUrl(info.srcUrl)`. If it returns a candidate URL, fetch and validate that candidate per the High-resolution candidate validation requirement.
   b. If no rule matched, or the candidate failed validation, or fetching the candidate errored, fetch `info.srcUrl` directly instead. This fallback fetch is not re-validated — it is treated the same as today's existing fetch path.
   c. The blob used for the remainder of the save SHALL be the valid high-res candidate's blob if one was obtained, otherwise the `info.srcUrl` blob.
4. Execute the 4-step upload sequence:
   a. `POST /images` → receives `{ upload_url, thumbnail_upload_url, id }`
   b. If `OffscreenCanvas` is available, decode the image via `createImageBitmap` and generate a thumbnail blob from the resulting bitmap (600px max, JPEG), capturing the bitmap's `width` and `height`. When step 3 already decoded the candidate to validate its dimensions, that same decoded bitmap SHALL be reused here rather than decoding again.
   c. `PUT` image blob to `upload_url` and (if thumbnail was generated) `PUT` thumbnail blob to `thumbnail_upload_url`, in parallel
   d. `POST /images/:id/complete` with a JSON body containing `file_size` (the byte length of the uploaded image blob), and `width`/`height` when they were captured in step (b)
5. Send a success toast to the active tab with title "Saved to Bookleaf." and body "Added to Unsorted."

The image title SHALL be set to the current tab's title (`tab.title`). The `source_url` SHALL be set to `info.pageUrl`. No `folder_id` SHALL be sent (image saves to root). The `tabId` from the context menu event SHALL be threaded into the save handler and used for all `browser.tabs.sendMessage` calls.

If thumbnail generation or the thumbnail PUT fails, the entire save SHALL fail (no partial save with a missing thumbnail).

If `OffscreenCanvas` is not available, the thumbnail PUT is skipped, no decode occurs, and the save proceeds with only the image PUT — the `complete` request body SHALL include `file_size` but omit `width`/`height`. The backend `HeadObject` fallback will enqueue the thumbnail worker in this case.

#### Scenario: Successful save shows in-page success toast

- **WHEN** the user clicks "Save to Bookleaf" while authenticated and the image is fetchable
- **THEN** the image and thumbnail are uploaded to Bookleaf with the page title as its title and page URL as its source_url
- **AND** an in-page toast with title "Saved to Bookleaf." and body "Added to Unsorted." is shown in the active tab

#### Scenario: Dimensions are sent when OffscreenCanvas is available

- **WHEN** the user saves an image and `OffscreenCanvas` is available
- **THEN** the `POST /images/:id/complete` request body includes `width` and `height` matching the decoded `ImageBitmap` dimensions, and `file_size` matching the byte length of the uploaded image blob

#### Scenario: Dimensions are omitted when OffscreenCanvas is unavailable

- **WHEN** the user saves an image and `OffscreenCanvas` is not available
- **THEN** the `POST /images/:id/complete` request body includes `file_size` but does not include `width` or `height`

#### Scenario: Unauthenticated save is rejected

- **WHEN** the user clicks "Save to Bookleaf" and no valid token exists in `chrome.storage.local`
- **THEN** no upload is attempted
- **AND** an in-page toast with title "Bookleaf" and body "Please log in first." is shown in the active tab

#### Scenario: Expired token is rejected

- **WHEN** the user clicks "Save to Bookleaf" and the stored token's `expiresAt` is in the past
- **THEN** no upload is attempted
- **AND** an in-page toast with title "Bookleaf" and body "Please log in first." is shown in the active tab

#### Scenario: Matching site rule resolves a valid high-res candidate

- **WHEN** the user saves an image whose `srcUrl` matches a high-res rule (e.g. a Twitter `:large` media URL) and the resolved candidate passes validation
- **THEN** the candidate's blob is uploaded as the saved image instead of the original thumbnail's blob

#### Scenario: Invalid high-res candidate falls back to the original URL without failing the save

- **WHEN** the user saves an image whose `srcUrl` matches a high-res rule, but the resolved candidate fails validation (non-OK response, disallowed content-type, or undersized dimensions)
- **THEN** the extension fetches `info.srcUrl` directly and proceeds with the save
- **AND** no error toast is shown solely because the high-res candidate was invalid

#### Scenario: No matching rule uses the original URL

- **WHEN** the user saves an image whose `srcUrl` does not match any high-res rule
- **THEN** the extension fetches `info.srcUrl` directly, exactly as before this change

## ADDED Requirements

### Requirement: Save failure notification covers fallback fetch failures

If fetching `info.srcUrl` (whether as the sole fetch when no rule matched, or as the fallback after an invalid/failed high-res candidate) returns a non-OK response or throws, the extension SHALL send an error toast to the active tab with title "Couldn't save image." and body "Check your connection and try again.", per the existing Save failure notification requirement.

#### Scenario: Fallback fetch failure shows in-page error toast

- **WHEN** a high-res candidate fails validation and the subsequent fallback fetch of `info.srcUrl` also returns a non-OK response or throws
- **THEN** an in-page toast with title "Couldn't save image." and body "Check your connection and try again." is shown
- **AND** no upload is initiated
