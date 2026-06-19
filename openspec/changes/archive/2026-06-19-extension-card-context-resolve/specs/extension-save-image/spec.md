## MODIFIED Requirements

### Requirement: Context menu registration

The background service worker SHALL register two "Save to Bookleaf" context menu items on extension install and startup. The first item SHALL appear only when right-clicking an `<img>` element (`contexts: ["image"]`). The second item SHALL appear only when right-clicking a link whose URL matches a registered link-only card site pattern (`contexts: ["link"]`, `targetUrlPatterns` set to the union of registered patterns, e.g. Pinterest pin URLs). Both items SHALL share the same `onClicked` listener, which SHALL distinguish them by `menuItemId`.

#### Scenario: Context menu item appears on image right-click

- **WHEN** the user right-clicks an `<img>` element on any webpage
- **THEN** a "Save to Bookleaf" option appears in the browser context menu

#### Scenario: Context menu item does not appear on non-image, non-registered-link right-click

- **WHEN** the user right-clicks text or a non-image element that is not a link matching a registered card site pattern
- **THEN** "Save to Bookleaf" does not appear in the context menu

#### Scenario: Context menu item appears on a registered link-only card site

- **WHEN** the user right-clicks a link whose URL matches a registered link-only card site pattern (e.g. a Pinterest pin URL), even though no `<img>` is the direct click target
- **THEN** a "Save to Bookleaf" option appears in the browser context menu

### Requirement: Authenticated save flow

When the user clicks "Save to Bookleaf", the extension SHALL:
1. Read the stored auth token from `chrome.storage.local`
2. If no token exists or the token is expired, send a toast message to the active tab with title "Bookleaf" and body "Please log in first." and abort
3. Determine the effective `srcUrl` to fetch:
   a. If `info.srcUrl` is present (the "image" context menu item was clicked), use it as the base `srcUrl`.
   b. Otherwise (the link-context menu item was clicked), use the `srcUrl` from the content-script-resolved card context for the tab, if one was reported before the click. If no resolved `srcUrl` is available, the save SHALL fail per the Save failure notification requirement, without attempting to fetch `info.linkUrl` as an image.
4. Determine the effective `source_url`: if `info.linkUrl` is present and matches a registered link permalink rule, use `info.linkUrl` as `source_url`; otherwise use `info.pageUrl` as today.
5. Resolve the image to fetch:
   a. Call `resolveHighResUrl(srcUrl)` (the effective `srcUrl` from step 3). If it returns a candidate URL, fetch and validate that candidate per the High-resolution candidate validation requirement.
   b. If no rule matched, or the candidate failed validation, or fetching the candidate errored, fetch the effective `srcUrl` directly instead. This fallback fetch is not re-validated — it is treated the same as today's existing fetch path.
   c. The blob used for the remainder of the save SHALL be the valid high-res candidate's blob if one was obtained, otherwise the effective `srcUrl`'s blob.
6. Execute the 4-step upload sequence:
   a. `POST /images` → receives `{ upload_url, thumbnail_upload_url, id }`
   b. If `OffscreenCanvas` is available, decode the image via `createImageBitmap` and generate a thumbnail blob from the resulting bitmap (600px max, JPEG), capturing the bitmap's `width` and `height`. When step 5 already decoded the candidate to validate its dimensions, that same decoded bitmap SHALL be reused here rather than decoding again.
   c. `PUT` image blob to `upload_url` and (if thumbnail was generated) `PUT` thumbnail blob to `thumbnail_upload_url`, in parallel
   d. `POST /images/:id/complete` with a JSON body containing `file_size` (the byte length of the uploaded image blob), and `width`/`height` when they were captured in step (b)
7. Send a success toast to the active tab with title "Saved to Bookleaf." and body "Added to Unsorted."

The image title SHALL be set to the current tab's title (`tab.title`). The `source_url` SHALL be set per step 4 above. No `folder_id` SHALL be sent (image saves to root). The `tabId` from the context menu event SHALL be threaded into the save handler and used for all `browser.tabs.sendMessage` calls.

If thumbnail generation or the thumbnail PUT fails, the entire save SHALL fail (no partial save with a missing thumbnail).

If `OffscreenCanvas` is not available, the thumbnail PUT is skipped, no decode occurs, and the save proceeds with only the image PUT — the `complete` request body SHALL include `file_size` but omit `width`/`height`. The backend `HeadObject` fallback will enqueue the thumbnail worker in this case.

#### Scenario: Successful save shows in-page success toast

- **WHEN** the user clicks "Save to Bookleaf" while authenticated and the image is fetchable
- **THEN** the image and thumbnail are uploaded to Bookleaf with the page title as its title and the effective `source_url` as its `source_url`
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

- **WHEN** the user saves an image whose effective `srcUrl` matches a high-res rule (e.g. a Twitter `:large` media URL) and the resolved candidate passes validation
- **THEN** the candidate's blob is uploaded as the saved image instead of the original thumbnail's blob

#### Scenario: Invalid high-res candidate falls back to the original URL without failing the save

- **WHEN** the user saves an image whose effective `srcUrl` matches a high-res rule, but the resolved candidate fails validation (non-OK response, disallowed content-type, or undersized dimensions)
- **THEN** the extension fetches the effective `srcUrl` directly and proceeds with the save
- **AND** no error toast is shown solely because the high-res candidate was invalid

#### Scenario: No matching rule uses the original URL

- **WHEN** the user saves an image whose effective `srcUrl` does not match any high-res rule
- **THEN** the extension fetches the effective `srcUrl` directly, exactly as before this change

#### Scenario: Link-context save uses the content-script-resolved image

- **WHEN** the user right-clicks a link matching a registered link-only card site pattern, the content script reported a resolved `srcUrl` for that tab before the click, and the user clicks "Save to Bookleaf"
- **THEN** the extension uses the resolved `srcUrl` as the effective `srcUrl` for the remainder of the save flow
- **AND** `source_url` is set to `info.linkUrl`

#### Scenario: Link-context save with no resolved image fails gracefully

- **WHEN** the user right-clicks a link matching a registered link-only card site pattern, no resolved `srcUrl` was reported for that tab, and the user clicks "Save to Bookleaf"
- **THEN** no upload is attempted
- **AND** an in-page toast with title "Couldn't save image." and body "Check your connection and try again." is shown in the active tab

#### Scenario: Image-context save with a matching linkUrl rule overrides source_url

- **WHEN** the user right-clicks an `<img>` element whose enclosing link's URL matches a registered link permalink rule (e.g. a Twitter or Facebook post permalink), and clicks "Save to Bookleaf"
- **THEN** `source_url` is set to `info.linkUrl` instead of `info.pageUrl`
- **AND** the rest of the save flow (image fetch, high-res resolution, upload) is unchanged

#### Scenario: Image-context save with no matching linkUrl rule keeps today's behavior

- **WHEN** the user right-clicks an `<img>` element whose `info.linkUrl` is absent, or present but not matching any registered link permalink rule
- **THEN** `source_url` is set to `info.pageUrl`, exactly as before this change
