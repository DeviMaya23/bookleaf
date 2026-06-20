## Why

Right-click → "Save to Bookleaf" is the only way to save an image today. It works, but it's an extra context-menu hop for every single save, and the user has found it to be the highest-friction part of an otherwise-fast workflow (informed by using Eagle's drag-to-save as a faster alternative). Native HTML5 drag-and-drop is available on every site already tested (Pinterest, Instagram) without any site cooperation, and a quick manual spike confirmed it's structurally compatible with the existing card-resolution logic — `dragstart` lands on the same `<a>` card wrapper that `contextmenu` already lands on, so the same DOM-walking resolvers apply unchanged.

## What Changes

- Add a `dragstart` listener in the content script that, for any element under the drag, resolves `srcUrl` (via `resolveCardImageSrc` when on a known card site, falling through to the dragged `<img>` itself or, failing that, an `<img>` descendant), `title` (via the existing per-site alt/tweet-text resolvers), and the card's permalink (`closest("a")?.href`) — reusing the exact same functions the `contextmenu` listener already calls. No new DOM-resolution logic; no `dataTransfer` parsing.
- Render a transient, content-script-owned drop-zone (shadow DOM, same isolation approach as the existing toast) that appears near the pointer's `dragstart` position and is removed on `dragend` (and, as a safety net for sites that cancel native dragging, on `pointerup`/`mouseup`) or `drop`.
- On `drop` inside the zone, send one message to the background service worker with the values captured at `dragstart`. Background runs them through the unchanged `resolveHighResUrl`/`resolveHighResReferrer`/`validateCandidate` and `resolveLinkPermalink` pipeline and calls the existing `handleSave`, identically to the right-click path.
- Right-click "Save to Bookleaf" is unchanged and remains available; drag-and-drop is purely an additional trigger.
- Drag-and-drop save is disabled on Bookleaf's own app (origin match against `VITE_APP_URL`), since saving an image that's already in Bookleaf doesn't make sense.

## Capabilities

### New Capabilities
- `extension-drag-drop-save`: Defines the drag-and-drop save trigger — the `dragstart`-time DOM capture, the transient pointer-anchored drop-zone UI, and the `drop`-time handoff to the existing save pipeline.

### Modified Capabilities
(none — `extension-save-image`, `extension-card-context-resolve`, and `extension-highres-image-resolve` are invoked as-is from a new trigger; their requirements do not change.)

## Impact

- `extensions/src/lib/dragImageResolveRule.ts`: new file housing `resolveDragImageSrc(target, pageUrl)`, the 3-tier `srcUrl` resolution helper.
- `extensions/src/content/index.ts`: new `dragstart`/`dragend`/`pointerup`/`mouseup`/`drop` listeners and drop-zone rendering, alongside the existing `contextmenu` listener and toast host. Drag listeners are gated by an origin check against `VITE_APP_URL` to exclude Bookleaf's own app.
- `extensions/src/background/index.ts`: new message handler (`handleDragSaveMessage`) for the drag-drop save trigger, calling the existing `handleSave` (no changes to `handleSave`, `resolveTitle`, or `handleContextMenuClick`).
- No changes to `cardDomResolveRules.ts`, `altTextResolveRule.ts`, `tweetTextResolveRule.ts`, `facebookAltResolveRule.ts`, `linkPermalinkRules.ts`, or `highResFetch.ts` — all consumed unchanged.
- `manifest.json`: no new permissions expected (drag-and-drop and content script DOM access are already covered by existing `content_scripts` + `<all_urls>`).
