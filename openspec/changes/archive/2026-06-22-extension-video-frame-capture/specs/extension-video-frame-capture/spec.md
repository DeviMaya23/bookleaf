## ADDED Requirements

### Requirement: Context menu registration for video elements

The background service worker SHALL register a "Save video frame to Bookleaf" context menu item that appears only when right-clicking a `<video>` element (`contexts: ["video"]`), independent of the existing image and link context menu items.

#### Scenario: Context menu item appears on video right-click

- **WHEN** the user right-clicks a `<video>` element on any webpage
- **THEN** a "Save video frame to Bookleaf" option appears in the browser context menu

#### Scenario: Context menu item does not appear on non-video right-click

- **WHEN** the user right-clicks an element that is not a `<video>` and has no native browser video context
- **THEN** "Save video frame to Bookleaf" does not appear in the context menu

### Requirement: Content script tracks the right-clicked video element

The content script SHALL extend its existing `contextmenu` listener to resolve the right-clicked `<video>` element — either `event.target` itself when it is an `HTMLVideoElement`, or the first `<video>` descendant found via `event.target.querySelector("video")` otherwise — and SHALL retain a reference to that element in memory for use by a subsequent capture request. This reference SHALL NOT be sent to the background script (DOM elements are not serializable across the extension messaging boundary).

#### Scenario: Right-clicking a bare video element is tracked

- **WHEN** the user right-clicks a `<video>` element directly
- **THEN** the content script retains a reference to that element

#### Scenario: Right-clicking a wrapper containing a video element is tracked

- **WHEN** the user right-clicks a non-video wrapper element that contains a `<video>` descendant
- **THEN** the content script retains a reference to that descendant `<video>` element

#### Scenario: Right-clicking an element with no associated video clears tracking

- **WHEN** the user right-clicks an element that is neither a `<video>` nor contains one
- **THEN** the content script SHALL NOT retain a stale reference from a previous right-click for a subsequent capture request

### Requirement: Frame capture via canvas drawImage

When the user clicks "Save video frame to Bookleaf", the background SHALL send a capture-request message to the active tab's content script, which SHALL draw the tracked `<video>` element's currently decoded frame onto an in-memory canvas sized to the element's native `videoWidth`/`videoHeight` via `drawImage`, then produce an image blob via `toBlob`. This SHALL NOT use tab screenshot capture (`captureVisibleTab`) or any position/bounding-rect-based cropping.

#### Scenario: Successful capture produces a blob matching the video's native resolution

- **WHEN** a capture request is received for a tracked `<video>` element with non-zero `videoWidth`/`videoHeight`
- **THEN** the content script SHALL produce an image blob whose dimensions match the video's `videoWidth` and `videoHeight`

#### Scenario: Captured frame excludes native playback controls

- **WHEN** a capture request is received while the video's native controls are visibly rendered on screen
- **THEN** the captured blob SHALL NOT include the controls overlay, since `drawImage` reads the decoded frame buffer directly rather than the rendered screen output

### Requirement: Capture aborts without a fallback on failure

If the tracked video element has `videoWidth` equal to `0` (metadata not loaded), or no video element is currently tracked, or `toBlob` throws or rejects (e.g. a tainted canvas from a cross-origin video served without CORS headers), the content script SHALL abort the capture, SHALL NOT attempt any fallback capture mechanism, and SHALL trigger an in-page toast with title "Bookleaf" and body "Can't capture this video."

#### Scenario: Video with unloaded metadata aborts capture

- **WHEN** a capture request is received for a tracked `<video>` element whose `videoWidth` is `0`
- **THEN** no `drawImage` call is attempted
- **AND** an in-page toast with title "Bookleaf" and body "Can't capture this video." is shown

#### Scenario: No tracked video element aborts capture

- **WHEN** a capture request is received but no video element is currently tracked
- **THEN** no capture is attempted
- **AND** an in-page toast with title "Bookleaf" and body "Can't capture this video." is shown

#### Scenario: Tainted canvas aborts capture without a fallback

- **WHEN** `toBlob` throws or rejects due to a tainted canvas
- **THEN** no further capture attempt (e.g. screenshot-based) is made
- **AND** an in-page toast with title "Bookleaf" and body "Can't capture this video." is shown

### Requirement: Captured frame shares persistence with existing save flows

A successfully captured video frame blob SHALL be passed to the existing capture save flow (`handleCapture`), using the active tab's `url` as `pageUrl` and the active tab's `title` as `title`, performing the same authenticated persistence steps (auth/token validation, thumbnail generation, image and thumbnail upload, in-page toast notification, and recent-save bookkeeping) as the existing snip-capture and right-click image flows.

#### Scenario: Successful video frame save shows the same success toast as other flows

- **WHEN** a video frame is captured while the user is authenticated and the upload succeeds
- **THEN** an in-page toast with title "Saved to Bookleaf." and body "Added to Unsorted." is shown
- **AND** the saved image appears in recent saves

#### Scenario: Unauthenticated video frame save is rejected

- **WHEN** a video frame is captured and no valid token exists in storage
- **THEN** no upload is attempted
- **AND** an in-page toast with title "Bookleaf" and body "Please log in first." is shown

#### Scenario: Video frame save failure shows the same error toast as other flows

- **WHEN** a video frame is captured and any step of the upload sequence fails
- **THEN** an in-page toast with title "Couldn't save image." and body "Check your connection and try again." is shown
