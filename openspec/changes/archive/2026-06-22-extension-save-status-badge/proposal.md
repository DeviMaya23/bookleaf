## Why

Bookleaf saves images at high resolution by default, so an upload (fetch the source image, generate a thumbnail, two presigned `PUT`s, a `complete` call) can take a few seconds. None of the three save entry points — drag-to-save, right-click context menu, snip capture — give any feedback until that whole chain finishes; the only signal today is the toast that fires at the end (`extensions/src/background/index.ts:330,332`). Between triggering a save and that toast, there's no way to tell whether the extension is working or has silently done nothing.

## What Changes

- Add a toolbar action badge that shows a single dot while at least one save is in flight, and clears when none are, across all three save flows (drag-drop, context menu, snip).
- The badge reflects in-flight presence only — no count, no success/failure state. Outcome reporting stays the toast's job (and "Recent saves" remains the durable record of what actually succeeded).
- Backed by a single in-memory counter in the background service worker, incremented when a save starts and decremented when it ends (success or failure), wrapped so the decrement always runs even if the save throws.
- The existing `DEV`-build badge text (`extensions/src/background/index.ts:11-14`) is removed, since it occupies the same badge slot this feature now uses.
- No new messaging, no new permissions, no new DOM/CSS — everything lives in the background script. The counter is intentionally not persisted: if the service worker dies mid-save (idle timeout or browser closing), the stale badge state dies with it and resets cleanly the next time the worker wakes for any reason.

## Capabilities

### New Capabilities
- `extension-save-status-badge`: the toolbar badge itself — in-flight counter, badge show/clear behavior, and its deliberate scope boundary (no count, no failure state, no persistence).

### Modified Capabilities
(none — this wraps the existing save entry points without changing their requirements)

## Impact

- `extensions/src/background/index.ts`: add the in-flight counter and badge update calls; wrap `handleSave`/`handleCapture` (or the shared `persistImage`) with start/end calls via `try/finally`; remove the `DEV` badge text block.
- No changes to `extensions/src/content/index.ts`, `extensions/manifest.json`, popup, or storage.
- No backend changes. Extension-only.
