# Spec: Extension Save Image

## Purpose

Defines the requirements for the "Save to Bookleaf" browser extension feature, which allows users to save images from any webpage directly to their Bookleaf library via a right-click context menu.

## Requirements

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
4. Determine the effective `source_url`: if `info.linkUrl` is present and matches a registered link permalink rule, use `info.linkUrl` as the raw `source_url`; otherwise use `info.pageUrl` as the raw `source_url`. The raw `source_url` SHALL then be passed through `cleanUrl()` before being sent to the backend.
5. Determine the effective `title`:
   a. If `info.linkUrl` is present and matches the Twitter status permalink rule, and the content-script-resolved card context for the tab contains a `title` (resolved tweet text), set the title to `@<handle>: <tweet text, truncated to 100 characters>...`, where `<handle>` is parsed from `info.linkUrl`'s path segment preceding `/status/`.
   b. Otherwise, if `info.linkUrl` matches the Twitter status permalink rule but no resolved tweet text is available, set the title to `@<handle>` (no colon, no text).
   c. Otherwise, if the active tab's URL is an Imgur, Instagram, or Facebook URL and the content-script-resolved card context for the tab contains a `title` (resolved `alt`/`aria-label` text), set the title to that resolved text verbatim (no truncation, no formatting).
   d. Otherwise, set the title to the current tab's title (`tab.title`), as today.
6. Resolve the image to fetch:
   a. Call `resolveHighResUrl(srcUrl)` (the effective `srcUrl` from step 3). If it returns a candidate URL, fetch and validate that candidate per the High-resolution candidate validation requirement.
   b. If no rule matched, or the candidate failed validation, or fetching the candidate errored, fetch the effective `srcUrl` directly instead. This fallback fetch is not re-validated — it is treated the same as today's existing fetch path.
   c. The blob used for the remainder of the save SHALL be the valid high-res candidate's blob if one was obtained, otherwise the effective `srcUrl`'s blob.
7. Execute the 4-step upload sequence:
   a. `POST /images` → receives `{ upload_url, thumbnail_upload_url, id }`
   b. If `OffscreenCanvas` is available, decode the image via `createImageBitmap` and generate a thumbnail blob from the resulting bitmap (600px max, JPEG), capturing the bitmap's `width` and `height`. When step 6 already decoded the candidate to validate its dimensions, that same decoded bitmap SHALL be reused here rather than decoding again.
   c. `PUT` image blob to `upload_url` and (if thumbnail was generated) `PUT` thumbnail blob to `thumbnail_upload_url`, in parallel
   d. `POST /images/:id/complete` with a JSON body containing `file_size` (the byte length of the uploaded image blob), and `width`/`height` when they were captured in step (b)
8. Send a success toast to the active tab with title "Saved to Bookleaf." and body "Added to Unsorted."

The image title SHALL be set per step 5 above (the current tab's title, `tab.title`, for saves with no matching Twitter permalink and no resolved Imgur/Instagram/Facebook text). The `source_url` sent to the backend SHALL be the ClearURLs-cleaned URL produced by `cleanUrl()` per step 4 above. No `folder_id` SHALL be sent (image saves to root). The `tabId` from the context menu event SHALL be threaded into the save handler and used for all `browser.tabs.sendMessage` calls.

If thumbnail generation or the thumbnail PUT fails, the entire save SHALL fail (no partial save with a missing thumbnail).

If `OffscreenCanvas` is not available, the thumbnail PUT is skipped, no decode occurs, and the save proceeds with only the image PUT — the `complete` request body SHALL include `file_size` but omit `width`/`height`. The backend `HeadObject` fallback will enqueue the thumbnail worker in this case.

#### Scenario: Successful save shows in-page success toast

- **WHEN** the user clicks "Save to Bookleaf" while authenticated and the image is fetchable
- **THEN** the image and thumbnail are uploaded to Bookleaf with the effective title (per step 5) and the ClearURLs-cleaned `source_url` (per step 4) as its `source_url`
- **AND** an in-page toast with title "Saved to Bookleaf." and body "Added to Unsorted." is shown in the active tab

#### Scenario: source_url is cleaned before being sent to the backend

- **WHEN** the user saves an image from a page whose URL contains tracking params matched by a ClearURLs provider rule (e.g. a Google Image Search URL with `ei`, `ved`, `biw` params)
- **THEN** the `source_url` field in the `POST /images` body has those params stripped
- **AND** the URL remains navigable

#### Scenario: source_url from an unrecognised host is passed through unchanged

- **WHEN** the user saves an image from a page whose host does not match any ClearURLs provider in the vendored subset
- **THEN** the `source_url` field in the `POST /images` body is identical to the raw resolved URL

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
- **AND** `source_url` is set to `cleanUrl(info.linkUrl)`

#### Scenario: Link-context save with no resolved image fails gracefully

- **WHEN** the user right-clicks a link matching a registered link-only card site pattern, no resolved `srcUrl` was reported for that tab, and the user clicks "Save to Bookleaf"
- **THEN** no upload is attempted
- **AND** an in-page toast with title "Couldn't save image." and body "Check your connection and try again." is shown in the active tab

#### Scenario: Image-context save with a matching linkUrl rule overrides source_url

- **WHEN** the user right-clicks an `<img>` element whose enclosing link's URL matches a registered link permalink rule (e.g. a Twitter or Facebook post permalink), and clicks "Save to Bookleaf"
- **THEN** `source_url` is set to `cleanUrl(info.linkUrl)` instead of `cleanUrl(info.pageUrl)`
- **AND** the rest of the save flow (image fetch, high-res resolution, upload) is unchanged

#### Scenario: Image-context save with no matching linkUrl rule keeps today's behavior

- **WHEN** the user right-clicks an `<img>` element whose `info.linkUrl` is absent, or present but not matching any registered link permalink rule, and the tab is not an Imgur, Instagram, or Facebook URL with resolved text
- **THEN** `source_url` is set to `cleanUrl(info.pageUrl)`
- **AND** title is set to `tab.title`, exactly as before this change

#### Scenario: Twitter save with resolved tweet text uses the handle-and-text title

- **WHEN** the user right-clicks an image inside a tweet (whose enclosing link matches the Twitter status permalink rule) and the content script resolved tweet text for that right-click
- **THEN** the title is set to `@<handle>: <tweet text truncated to 100 characters>...`, where `<handle>` is parsed from `info.linkUrl`

#### Scenario: Twitter save with no resolved tweet text falls back to handle only

- **WHEN** the user right-clicks an image inside a tweet (whose enclosing link matches the Twitter status permalink rule) and no tweet text was resolved for that right-click (e.g. an image-only tweet)
- **THEN** the title is set to `@<handle>` with no colon or trailing text

#### Scenario: Imgur save uses the right-clicked image's alt text

- **WHEN** the user right-clicks an `<img>` element on an Imgur page whose `alt` attribute is non-empty, and clicks "Save to Bookleaf"
- **THEN** the title is set to that `alt` text verbatim

#### Scenario: Imgur save with no usable alt text falls back to tab title

- **WHEN** the user right-clicks an `<img>` element on an Imgur page whose `alt` attribute is empty or absent, and clicks "Save to Bookleaf"
- **THEN** the title is set to `tab.title`, exactly as before this change

#### Scenario: Instagram save uses the right-clicked image's alt text

- **WHEN** the user right-clicks an `<img>` element on an Instagram page whose `alt` attribute is non-empty (whether it contains the post's full caption text or Instagram's generated `"Photo by <name> on <date>."` form), and clicks "Save to Bookleaf"
- **THEN** the title is set to that `alt` text verbatim, with no parsing or reformatting

#### Scenario: Instagram save with no usable alt text falls back to tab title

- **WHEN** the user right-clicks an `<img>` element on an Instagram page whose `alt` attribute is empty or absent, and clicks "Save to Bookleaf"
- **THEN** the title is set to `tab.title`, exactly as before this change

#### Scenario: Facebook save uses the right-clicked image's own alt text when present

- **WHEN** the user right-clicks an `<img>` element on a Facebook page whose own `alt` attribute is non-empty, and clicks "Save to Bookleaf"
- **THEN** the title is set to that `alt` text verbatim

#### Scenario: Facebook save falls back to an ancestor's aria-label when the image's own alt is empty

- **WHEN** the user right-clicks an `<img>` element on a Facebook page whose own `alt` attribute is empty, and a bounded ancestor search finds an `aria-label` attribute (e.g. "May be an image of gelato and text") within the search depth limit, and clicks "Save to Bookleaf"
- **THEN** the title is set to that `aria-label` text verbatim

#### Scenario: Facebook save with no usable alt or aria-label falls back to tab title

- **WHEN** the user right-clicks an `<img>` element on a Facebook page whose own `alt` attribute is empty and no ancestor within the search depth limit has a usable `aria-label`, and clicks "Save to Bookleaf"
- **THEN** the title is set to `tab.title`, exactly as before this change

### Requirement: Save failure notification

If any step in the save flow fails (image fetch error, API error, presigned PUT failure), the extension SHALL send an error toast to the active tab with title "Couldn't save image." and body "Check your connection and try again." No retry is attempted automatically. If `sendMessage` rejects (e.g., tab navigated away), the error SHALL be silently swallowed.

#### Scenario: Image fetch failure shows in-page error toast

- **WHEN** the background SW fetch of the image URL returns a non-OK response or throws
- **THEN** an in-page toast with title "Couldn't save image." and body "Check your connection and try again." is shown
- **AND** no upload is initiated

#### Scenario: Upload API failure shows in-page error toast

- **WHEN** any step of the 4-step upload sequence returns a non-2xx response
- **THEN** an in-page toast with title "Couldn't save image." and body "Check your connection and try again." is shown

#### Scenario: sendMessage rejection is silently ignored

- **WHEN** `browser.tabs.sendMessage` rejects because the tab navigated away
- **THEN** no unhandled error is thrown in the background service worker

### Requirement: Save failure notification covers fallback fetch failures

If fetching `info.srcUrl` (whether as the sole fetch when no rule matched, or as the fallback after an invalid/failed high-res candidate) returns a non-OK response or throws, the extension SHALL send an error toast to the active tab with title "Couldn't save image." and body "Check your connection and try again.", per the existing Save failure notification requirement.

#### Scenario: Fallback fetch failure shows in-page error toast

- **WHEN** a high-res candidate fails validation and the subsequent fallback fetch of `info.srcUrl` also returns a non-OK response or throws
- **THEN** an in-page toast with title "Couldn't save image." and body "Check your connection and try again." is shown
- **AND** no upload is initiated

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
