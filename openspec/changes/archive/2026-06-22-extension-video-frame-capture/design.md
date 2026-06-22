## Context

The extension already has three save entry points (right-click image, drag-drop, manual snip), all converging on shared persistence logic. Right-click image works via `browser.contextMenus.create({ contexts: ["image"] })`: the browser supplies `info.srcUrl` directly in the `onClicked` payload, so the background script never needs a live reference to the clicked DOM element.

Video frame capture breaks that pattern: there is no `info.srcUrl` equivalent for "the currently decoded frame." Capturing it requires `canvas.drawImage()` against the actual `<video>` element, which only exists in the content script's DOM — the background service worker has no way to hold or receive a DOM node reference across the messaging boundary (`browser.runtime.sendMessage` payloads are structured-clone-serializable; DOM elements are not).

## Goals / Non-Goals

**Goals:**
- Capture the currently rendered frame of a right-clicked `<video>`, free of native controls overlay.
- Reuse the existing `handleCapture` persistence pipeline unchanged.
- Keep the change additive — no modification to the existing image/link/drag/snip flows.

**Non-Goals:**
- Position/geometry-based capture (no `getBoundingClientRect`, no screenshot fallback) — explicitly ruled out during exploration due to `object-fit`/transform/scroll fragility.
- Handling tainted-canvas (cross-origin, non-CORS) video — surfaces a toast and aborts, no fallback capture path.
- Drag-zone or full-screen video capture — out of scope per the proposal.

## Decisions

**1. Track the right-clicked `<video>` element reference in the content script, not in the background.**
The content script's existing `contextmenu` listener (`content/index.ts:333`) already runs synchronously on every right-click. It will be extended to check whether `event.target` is an `HTMLVideoElement` (or contains one — checking `event.target.closest("video")` to match the native browser context-detection behavior) and, if so, store that element in a module-level variable (e.g. `lastRightClickedVideo`). This reference never leaves the content script. The background's `contextMenus.onClicked` listener (which fires natively for `contexts: ["video"]`, no extra wiring needed to detect the click target) sends a `"capture-video-frame"` message to the tab; the content script's `onMessage` handler reads `lastRightClickedVideo` to perform the capture.
- *Alternative considered*: store a serializable "locator" (e.g. an index or selector) in the background's existing `resolvedContextByTab`-style map, and re-query the DOM in the content script on click. Rejected — re-querying risks resolving a different element if the page's DOM changed between right-click and menu-click (e.g. a re-rendered video list), whereas holding the actual reference is exact and matches how short-lived the gap between right-click and click is in practice.

**2. New context menu item is independent of the existing two, sharing only the `onClicked` listener via a new `menuItemId` branch.**
Mirrors today's `save-to-bookleaf` / `save-to-bookleaf-link` pattern: `browser.contextMenus.create({ id: "save-video-frame-to-bookleaf", contexts: ["video"] })`, with a new `if (info.menuItemId === "save-video-frame-to-bookleaf")` branch in `handleContextMenuClick`.

**3. Content script performs the capture and sends the resulting blob back via a new message type, parallel to `snip-captured`.**
`canvas.drawImage(video, 0, 0, video.videoWidth, video.videoHeight)` → `canvas.toBlob()` → `browser.runtime.sendMessage({ type: "video-frame-captured", blob, mimeType })`. Background's existing `onMessage` listener gets a new branch parallel to the existing `snip-captured` one, calling `handleCapture` with `pageUrl`/`title` from `sender.tab` — identical metadata resolution to `handleSnipCapturedMessage`.

**4. Guard checks run in the content script, before any message is sent back.**
- If `lastRightClickedVideo` is null (menu click outside any plausible video gap) or `videoWidth === 0`, skip capture and request a toast directly (reusing the existing toast message type) instead of attempting a 0×0 `drawImage`.
- If `toBlob()` throws/rejects (tainted canvas), catch and request the same toast. No retry, no fallback capture path.

**5. No new shared module.** The video-frame logic lives inline in `content/index.ts` and `background/index.ts` alongside the equivalent snip-capture code, rather than extracting a shared "capture pipeline" abstraction — there's only one other capture flow (snip) and it uses a different capture technique (`captureVisibleTab` vs. `drawImage`), so the only real sharing point remains `handleCapture`, which is already shared.

## Risks / Trade-offs

- **[Risk] Race between right-click and menu-click on a fast-changing page** (e.g. an infinite-scroll video feed where the element is removed from the DOM before the user clicks the menu item) → the stored element reference becomes stale/detached. *Mitigation*: a detached `HTMLVideoElement` still has its last-decoded frame in memory and `drawImage` will still succeed against it in practice; if it doesn't (browser-dependent), this falls into the same "capture failed" toast path as any other failure — no special-casing needed.
- **[Risk] Ambiguous "contains a video" detection** when `event.target` is a wrapper div with a nested `<video>` (common in custom players) → `closest("video")` only matches the video itself or its ancestors, not descendants. *Mitigation*: use `event.target instanceof HTMLVideoElement ? event.target : event.target.querySelector("video")` to also catch the wrapper-click case, consistent with how the native `contexts: ["video"]` menu item itself becomes visible (the browser does its own descendant resolution to decide when to show the item, so the content script's resolution should match that behavior as closely as possible).
- **[Trade-off] No fallback capture for tainted canvas** means some cross-origin video embeds will never be capturable by this feature. Accepted per explicit user decision — simplicity over coverage.

