## 1. Manifest

- [x] 1.1 Add a `commands` entry to `extensions/manifest.json` for the snip hotkey (e.g. `snip-capture`, with a sensible default key combination and description "Snip to Bookleaf")
- [x] 1.2 Verify the command appears in `chrome://extensions/shortcuts` (Chrome) and the Firefox shortcut editor under `about:addons`, and that it's remappable there

## 2. Background: extract `persistImage()`

- [x] 2.1 In `extensions/src/background/index.ts`, extract the auth check, thumbnail generation, upload, toast, and `addRecentSave` logic currently inline in `handleSave` into a new `persistImage({ blob, mimeType, title, pageUrl, tabId })` function
- [x] 2.2 Update `handleSave` to call `resolveImageBlob(srcUrl)` then delegate to `persistImage`, preserving identical behavior for the right-click and drag-drop flows
- [x] 2.3 Unit test `persistImage` directly: successful save (auth valid, upload succeeds → success toast + recent-save entry), unauthenticated save (no/expired token → login-required toast, no upload attempted), upload failure (any PUT/POST step fails → error toast, no partial save)

## 3. Background: capture trigger and `handleCapture`

- [x] 3.1 Add a `browser.commands.onCommand` listener that, on the snip command, calls `browser.tabs.captureVisibleTab()` for the active tab and sends the resulting image data URL to that tab's content script via `browser.tabs.sendMessage`
- [x] 3.2 Add a message handler for the cropped-blob message from the content script (sent on mouseup) that resolves `pageUrl`/`title` from the sender tab's `url`/`title` and calls the new `handleCapture({ blob, mimeType, pageUrl, title, tabId })`
- [x] 3.3 Implement `handleCapture` to delegate directly to `persistImage` (no fetch step)
- [x] 3.4 Unit test `handleCapture`: delegates to `persistImage` with the given blob/mimeType/title/pageUrl without attempting any fetch; pageUrl/title pass through tab.url/tab.title unmodified (no per-site resolution)

## 4. Content script: overlay UI

- [x] 4.1 In `extensions/src/content/index.ts`, add a message handler for the background's captured-frame message that renders a full-viewport overlay (in the existing shadow DOM) showing the captured image as a frozen, dimmed background
- [x] 4.2 Implement mousedown/mousemove/mouseup handling to draw a live selection rectangle over the frozen frame, with the dimmed mask cut out to reveal real pixels inside the current rectangle bounds
- [x] 4.3 On mouseup, crop the frozen frame to the finalized selection rectangle via an offscreen canvas, produce a blob, and send it to the background worker; then remove the overlay
- [x] 4.4 Add an Escape keydown handler (active any time the overlay is shown, including mid-drag) that removes the overlay immediately with no crop or message sent
- [x] 4.5 Ensure the overlay only renders in the top frame (rely on existing `all_frames` default/false behavior — no extra guard needed, but confirm no duplicate overlay appears on pages with iframes)

## 5. Verification

- [x] 5.1 Run `npm run build` in `extensions/` and fix any errors
- [x] 5.2 Run `npm run lint` in `extensions/` and fix any issues (no `lint` script exists in this project; ran `npm run type-check` instead — clean)
- [x] 5.3 Manually verify on a regular webpage: hotkey → frozen overlay appears → drag-select → release → image saved with success toast → appears in recent saves with correct title/source URL (tab title/URL)
- [x] 5.4 Manually verify Escape cancels both before and during a drag, with no save
- [x] 5.5 Manually verify the existing right-click and drag-drop save flows still work unchanged after the `persistImage` extraction
