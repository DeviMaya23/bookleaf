## Why

Saving an image to Bookleaf currently requires that image to already exist as a fetchable resource on the page (right-click on an `<img>`, or drag-and-drop). There's no way to save canvas-rendered content, video frames, CSS-background images, or any other on-screen pixels that aren't backed by a fetchable URL — for those, users fall back to the OS's native snipping tool and manually paste the result into Bookleaf. A hotkey-activated, in-extension snipping tool removes that manual paste step and extends saving to anything visible in the viewport.

## What Changes

- Add a new `commands` entry to the extension manifest for a hotkey (e.g. "Snip to Bookleaf"), remappable via the browser's native shortcut settings (`chrome://extensions/shortcuts` / `about:addons`) — no in-extension remapping UI in this change.
- On hotkey press, capture the current tab's visible viewport (`captureVisibleTab`), then show a dimmed, frozen-frame overlay in the existing top-frame content script.
- User drags one selection rectangle over the frozen frame; on mouseup, the selection is cropped via canvas and saved immediately — no preview, no resize handles, no confirmation step ("what you snip is what you get").
- Escape cancels the overlay at any point before mouseup, with no save attempted.
- `pageUrl` and `title` for a snip are taken directly from the active tab's `url`/`title` — no link/permalink/site-specific resolution (unlike the right-click flow).
- **Refactor (no behavior change to existing flows)**: extract the shared persist tail (auth check, thumbnail generation, upload, toast, recent-save bookkeeping) out of `handleSave` into a new `persistImage()` function, so the existing `handleSave` (resolves a `srcUrl` via fetch) and the new `handleCapture` (already has image bytes, no fetch) both call into it.

## Capabilities

### New Capabilities
- `extension-snip-capture`: hotkey-activated viewport snipping — manifest command registration, capture-and-freeze overlay UX in the content script, drag-to-select-and-crop interaction, and the `handleCapture` save path.

### Modified Capabilities
(none — `extension-save-image`'s observable behavior for the right-click flow is unchanged; the `persistImage` extraction is an internal implementation refactor, not a requirements change)

## Impact

- `extensions/src/background/index.ts`: extract `persistImage()` from `handleSave()`; add `handleCapture()`; add `commands.onCommand` listener; add `captureVisibleTab()` call.
- `extensions/src/content/index.ts`: add overlay UI (frozen-frame render, dimming, drag-selection, crosshair cursor, Escape-to-cancel), canvas crop logic, and the message that hands the cropped blob to the background worker.
- `extensions/manifest.json`: add `commands` key for the hotkey.
- No backend or frontend (web app) changes — this is extension-only.
