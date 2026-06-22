## 1. Background: context menu and message handling

- [x] 1.1 Add `browser.contextMenus.create({ id: "save-video-frame-to-bookleaf", title: "Save video frame to Bookleaf", contexts: ["video"] })` alongside the existing two `contextMenus.create` calls in the `onInstalled` listener (`background/index.ts`).
- [x] 1.2 Add a `"save-video-frame-to-bookleaf"` branch in `handleContextMenuClick` that sends a capture-request message (e.g. `{ type: "capture-video-frame" }`) to the active tab via `browser.tabs.sendMessage`, instead of calling `handleSave` directly (no `srcUrl` is available from `info` for this menu item).
- [x] 1.3 Add a `VideoFrameCapturedMessage` type and `handleVideoFrameCapturedMessage` function (parallel to `SnipCapturedMessage`/`handleSnipCapturedMessage`) that resolves `pageUrl`/`title` from `sender.tab` and calls `handleCapture`.
- [x] 1.4 Add a `"video-frame-captured"` branch in the `runtime.onMessage` listener wired to `handleVideoFrameCapturedMessage`.

## 2. Content script: video tracking and capture

- [x] 2.1 Extend the existing `contextmenu` listener in `content/index.ts` to resolve a `<video>` element from `event.target` (direct match or first `<video>` descendant via `querySelector`), storing it in a module-level `lastRightClickedVideo` variable; clear it to `null` when no video resolves.
- [x] 2.2 Add a `runtime.onMessage` case for `"capture-video-frame"` that: aborts (sends a toast request) if `lastRightClickedVideo` is `null` or its `videoWidth` is `0`; otherwise draws the frame via `drawImage` onto a canvas sized to `videoWidth`/`videoHeight` and calls `toBlob`.
- [x] 2.3 On successful `toBlob`, `browser.runtime.sendMessage({ type: "video-frame-captured", blob, mimeType: "image/png" })`.
- [x] 2.4 On `toBlob` failure/rejection (tainted canvas) or the abort conditions in 2.2, trigger the existing toast mechanism with title "Bookleaf" and body "Can't capture this video." — no fallback capture attempt.

## 3. Unit tests (background handler layer)

- [x] 3.1 In `background/index.test.ts`, add tests for `handleContextMenuClick` covering the `"save-video-frame-to-bookleaf"` branch: asserts `browser.tabs.sendMessage` is called with the capture-request message for the active tab.
- [x] 3.2 Add tests for `handleVideoFrameCapturedMessage` mirroring the existing `handleSnipCapturedMessage` tests: asserts `handleCapture` is invoked with the blob, mime type, and the tab's `url`/`title` as `pageUrl`/`title`.

## 4. Verification

- [x] 4.1 Manually verify in a loaded unpacked extension: right-click a `<video>` on a same-origin/self-hosted test page shows the menu item and saves a frame without controls visible in the result.
- [x] 4.2 Manually verify the `videoWidth === 0` and tainted-canvas abort paths each show the "Can't capture this video." toast and produce no save.
- [x] 4.3 Run `npm run build` in `extensions/` and fix any errors.
- [x] 4.4 Run `npm run type-check` in `extensions/` and fix any issues (no `lint` script exists in this package; `type-check` is the equivalent static check) — passed clean.
