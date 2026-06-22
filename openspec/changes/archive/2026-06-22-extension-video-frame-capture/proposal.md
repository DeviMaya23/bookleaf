## Why

The extension can save images via right-click, drag-drop, and a manual snip tool, but has no way to save a frame from a `<video>` element. Users want to capture "what's currently playing" without downloading the underlying video file, and a manual workaround (screen-capturing via snip) reliably includes the native video controls bar in the saved frame — undesirable for the saved image.

## What Changes

- Add a new "Save video frame to Bookleaf" context menu item, scoped to the native browser `contexts: ["video"]` context type (parallel to the existing `contexts: ["image"]` item) — no DOM-level video detection needed, the browser resolves this natively.
- Content script records a reference to the right-clicked `<video>` element on `contextmenu` (parallel to the existing card-context tracking used by the link-only save flow), so the background can ask it to perform the capture once the menu item is clicked.
- On click, the content script captures the frame via `canvas.drawImage(videoEl, 0, 0, videoEl.videoWidth, videoEl.videoHeight)` and `toBlob()` — reading the decoded frame buffer directly, which never includes the native controls overlay (a separate browser-rendered UI layer).
- If the video's `videoWidth` is `0` (metadata never loaded) or `toBlob()` throws `SecurityError` (tainted canvas, e.g. a cross-origin video served without CORS headers), the flow aborts with an in-page toast ("Bookleaf" / "Can't capture this video.") — no fallback capture mechanism is attempted.
- A successful capture is passed into the existing `handleCapture` save pipeline (shared with snip-capture), using the active tab's `url` and `title` as metadata — identical metadata handling to snip-capture, since there's no per-site DOM resolution applicable to a video frame.

## Capabilities

### New Capabilities
- `extension-video-frame-capture`: Right-click-triggered capture of the current rendered frame of a `<video>` element, saved through the existing image save pipeline.

### Modified Capabilities
(none — this introduces a new, additive capability; no existing capability's requirements change)

## Impact

- `extensions/src/background/index.ts`: add a third `contextMenus.create` call (`contexts: ["video"]`), a new `menuItemId` branch in the `onClicked` handler, and a new case in the `runtime.onMessage` listener (parallel to the existing `snip-captured` case) to receive the captured blob and invoke `handleCapture` — the sole existing call site of `handleCapture` (line 61) gains a sibling, not a modification.
- `extensions/src/content/index.ts`: extend the existing `contextmenu` listener (used today for card-context resolution) to also record the right-clicked element when it is a `<video>`, and add a new `runtime.onMessage` case to perform the `drawImage`/`toBlob` capture and `sendMessage` the result back (parallel to the existing `snip-captured` message).
- No backend or web app changes — image upload/thumbnail/persistence pipeline is reused unchanged via `handleCapture`.
