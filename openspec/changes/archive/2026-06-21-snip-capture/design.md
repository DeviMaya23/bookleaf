## Context

The extension has two existing save flows, both converging on `handleSave()` in `extensions/src/background/index.ts`: right-click (`handleContextMenuClick`) and drag-and-drop (`handleDragSaveMessage`). Both supply a `srcUrl` string; `handleSave` fetches it (`resolveImageBlob`), generates a thumbnail, uploads, toasts, and records a recent save.

A snip has no `srcUrl` — the user selects a region of the already-rendered viewport, and the extension produces a cropped image blob directly via `captureVisibleTab()` + canvas. There's nothing to fetch. This is also the expected shape of a future "capture" category (e.g. saving a frame from a `<video>` element), so the design separates "resolve a reference into bytes" from "persist bytes" rather than bolting capture onto `handleSave`'s fetch-based contract.

The content script (`extensions/src/content/index.ts`) is already injected into every page (`<all_urls>`, top-frame only — `all_frames` is unset/`false`), and already renders one piece of injected UI (the drag-drop zone) in a shadow DOM. The overlay reuses this same injection point rather than adding a second content script.

## Goals / Non-Goals

**Goals:**
- Let the user snip any region of the current tab's viewport via a hotkey and save it to Bookleaf with one drag gesture.
- Share persistence logic (thumbnail, upload, toast, recent-save) between the existing URL-based flows and the new bytes-based flow, without duplicating it.
- Keep v1 minimal: one rectangle, no resize, no preview, auto-save on mouseup.

**Non-Goals:**
- Capturing anything outside the browser viewport (other windows, the desktop) — that requires `getDisplayMedia`/`desktopCapture` and OS-level permission prompts, a different feature.
- Capturing content that has scrolled out of view — viewport-only for v1, per explicit decision.
- In-extension hotkey remapping UI — the browser's native shortcut settings page covers this for v1.
- Video frame capture — anticipated as a natural next user of the same `handleCapture`/`persistImage` path, but not built in this change.
- Resize handles, multi-step preview/confirm before save.

## Decisions

### 1. Extract `persistImage()` from `handleSave()`

`handleSave({ srcUrl, pageUrl, title, tabId })` currently does: auth check → `resolveImageBlob(srcUrl)` (fetch) → thumbnail generation → upload → toast → recent-save bookkeeping. Everything after the fetch is identical to what a snip needs, since a snip already has its blob.

`persistImage({ blob, mimeType, title, pageUrl, tabId })` becomes the shared tail: auth check, `generateThumbnail`, `saveImage`, `sendToast`, `addRecentSave`. `handleSave` shrinks to fetch-then-delegate. `handleCapture({ blob, mimeType, pageUrl, title, tabId })` is the new entry point that delegates directly, with no fetch step.

Alternative considered: keep `handleSave` as the single entry point and have the snip flow synthesize a `data:` URL or `blob:` URL to pass through `resolveImageBlob`. Rejected — `resolveImageBlob` also runs high-res-candidate resolution (`resolveHighResUrl`/`resolveHighResReferrer`) that only makes sense for real page URLs; forcing a blob through that path means special-casing it anyway, with no benefit over a direct second entry point.

### 2. Freeze-first capture, not live-select

On hotkey: capture the visible tab immediately via `captureVisibleTab()`, then render the overlay with that captured image as a frozen background, dimmed except for the selection rectangle. The user drags over the frozen frame, not the live page.

Alternative considered: dim the live page first, let the user drag a rectangle, and only call `captureVisibleTab()` at mouseup. Rejected — this introduces a timing gap between what the user saw while dragging and what actually gets captured (page scroll, video frame advance, animation), and matches no existing OS snipping tool's mental model. Freeze-first guarantees the saved pixels are exactly what's behind the selection rectangle the user sees.

### 3. Hotkey via manifest `commands`, native remap only

Add a `commands` entry to `manifest.json`; rely on `chrome://extensions/shortcuts` / Firefox's `about:addons` shortcut editor for remapping. No custom settings UI in this change — the project's own extension settings panel is expected to host a custom remapping UI in the future, but that's out of scope here.

### 4. Capture is viewport-scoped, no scroll-stitching

`captureVisibleTab()` only returns what's currently rendered. No attempt is made to capture the full page or stitch multiple scroll positions into one image. The selection rectangle is bounded to the captured viewport image's dimensions.

### 5. `pageUrl`/`title` resolution: tab metadata only, no per-site logic

Unlike right-click's `resolveTitle` (which special-cases Twitter/Imgur/Instagram/Facebook), a snip's `pageUrl` is `tab.url` and `title` is `tab.title`, unconditionally. There's no DOM element being referenced, so none of the per-site alt-text/permalink resolution applies.

### 6. Overlay lives in the existing content script, top-frame only

The overlay is new UI rendered into the same shadow DOM already used for the drag-drop zone in `extensions/src/content/index.ts`. Because `all_frames` is unset (defaults to `false`), only the top frame ever runs the content script, so the background worker's command-triggered message reaches exactly one overlay instance per tab — no de-duplication logic needed.

## Risks / Trade-offs

- [Risk, confirmed during implementation] `captureVisibleTab()` is not exposed on `browser.tabs` in Firefox unless `activeTab` (or `tabs`) is declared under `permissions` — `host_permissions: ["<all_urls>"]` alone does not expose it. Added `"activeTab"` to the manifest `permissions` array to resolve.
- [Risk] Pages where the content script cannot run (`chrome://`, the Chrome Web Store, some PDF viewers) won't show an overlay even though the hotkey fires. → Mitigation: no special handling in v1; the hotkey simply does nothing visible on such pages, consistent with how the existing drag-drop zone also doesn't appear there.
- [Trade-off] Viewport-only scope means content above/below the fold isn't snippable without first scrolling it into view. Acceptable per explicit non-goal; revisit only if it becomes a real friction point.
- [Risk] Extracting `persistImage()` touches code shared with the existing right-click flow. → Mitigation: the refactor is required to preserve identical behavior for `handleSave`'s existing callers (verified by existing `extension-save-image` spec scenarios/tests); no requirement-level change is intended.

## Open Questions

None outstanding — viewport-only scope, auto-save-on-mouseup, no-resize, and Escape-to-cancel were all confirmed during exploration.
