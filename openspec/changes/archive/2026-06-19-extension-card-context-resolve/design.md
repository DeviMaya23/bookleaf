## Context

The save flow currently has two layers:
1. Browser-native context (`info.srcUrl`, `info.pageUrl`, `tab.title`) — page-level only, populated by `contextMenus.onClicked`.
2. `highResRules.ts` — rewrites a known `srcUrl` to a higher-res variant (Twitter `name=orig`, Pinterest size-segment → `originals`). This layer is unaffected by this change; it already does the right thing once handed a `srcUrl`.

Manual testing during exploration established the actual DOM shapes:
- **Twitter/X media-tab grid, Facebook posts**: thumbnail `<img>` is wrapped in `<a href=".../status/123/photo/1">`. Both "Open image" and "Open link" appear in the native menu, confirming Chrome's `OnClickData` populates `info.linkUrl` even though the registered menu item's `contexts` is `["image"]` — `contexts` only gates whether the item *shows*, not which fields `info` carries.
- **Pinterest grid/feed cards**: only "Open link" appears. The card is `<a href="/pin/...">` wrapping an overlay `<div>` with the `<img>` nested inside but not the click target — `contexts: ["image"]` never matches, so the menu item doesn't show, and `info.srcUrl` would be absent even if it did.

## Goals / Non-Goals

**Goals:**
- Twitter and Facebook saves from card/grid views get the actual post permalink as `source_url`, not the listing page.
- Pinterest grid cards get a working "Save to Bookleaf" menu item that saves the actual pin image (fed through the existing high-res rewrite).
- The mechanism added for Pinterest (content-script → background messaging) is shaped so that a future per-site `title` resolver can reuse it without a new channel or contract change.

**Non-Goals:**
- Title extraction (e.g. "@username — post text"). Only the message *shape* anticipates this; no extraction logic is implemented now.
- Sites beyond Twitter, Facebook, Pinterest.
- A menu item that shows on arbitrary links regardless of site (Option A from exploration) — rejected for its "shows up, sometimes does nothing" UX.
- Fetching/scraping a linked page's HTML (e.g. Pinterest's `og:image`) as the resolution mechanism — superseded by the DOM-walk approach below, which needs no network round-trip.

## Decisions

### Decision: Two resolution mechanisms, chosen per site by data availability, not a single "kind" enum
Twitter/Facebook need only a `source_url` override and already have `srcUrl` — no content script involvement needed at all; the background click handler can do this synchronously by reading `info.linkUrl`. Pinterest needs an `srcUrl` it doesn't have, which can only come from the DOM. Modeling these as one rule table with a `kind` discriminant was considered during exploration but rejected: the two paths have different inputs (`info` object vs. live DOM), different timing (synchronous in the click handler vs. async message arriving before the click), and different registration needs (no menu change vs. a second `contextMenus.create` item). Keeping them as two independent, separately-registered rule sets is simpler than a unified abstraction that would mostly branch on `kind` anyway. Revisit if a future site needs *both* DOM resolution and a distinct override in ways that don't fit either list cleanly.

### Decision: `linkUrl` override rule table (Twitter, Facebook) lives in `extensions/src/lib/`, parallel to `highResRules.ts`
Each entry: `{ id, matches(url: URL): boolean }`. In the existing `contextMenus.onClicked` handler, after the existing `srcUrl` logic, check `info.linkUrl` against this table; if a rule matches, use `info.linkUrl` as `source_url` instead of `info.pageUrl`. No menu changes — the existing `contexts: ["image"]` item already fires since these sites' thumbnails are real `<img>` elements.

### Decision: Pinterest resolution via content-script DOM-walk, not page fetch+scrape
Considered fetching `info.linkUrl` server-side and extracting `og:image`. Rejected: adds a network round-trip and a dependency on Pinterest's meta tags being present and unauthenticated-readable, when the actual `<img>` is already sitting in the DOM the user right-clicked into — no fetch needed if we can reach it. The content script approach also generalizes better to a future site whose DOM holds data (e.g. post text) that no page-level fetch could reconstruct anyway.

Mechanism: content script adds a `contextmenu` listener in capture phase (Pinterest's overlay may intercept bubble-phase handlers it doesn't expect to be observed). On a right-click, walk from `event.target`: `closest('a')` to find the card's link wrapper, then `querySelector('img')` within it to find the actual image. If found, send `{ tabId implicit via runtime.sendMessage, resolved: { srcUrl } }` to the background. The background stores the most recent resolved context keyed by `tabId` (from `browser.tabs.getCurrent()` context, or by deriving it from the sender in `runtime.onMessage`), with a short TTL/overwrite-on-next-event, since `contextmenu` always fires immediately before any subsequent `contextMenus.onClicked` for that tab.

The resolved payload type is `Partial<{ srcUrl: string; title: string }>` — only `srcUrl` is populated by this change. This is the forward-compatibility hook: a future Twitter title resolver populates `title` through the same message and same per-tab store, with no channel change.

### Decision: Second `contextMenus.create` item for link-only card sites, scoped by `targetUrlPatterns`
A new item with `contexts: ["link"]` and `targetUrlPatterns: ["*://*.pinterest.com/pin/*"]` (extendable per future site). This is a native WebExtensions feature (works on both Chrome and Firefox, consistent with `extension-firefox-compat`), so the menu only appears on recognized permalink shapes — not on every link, avoiding the rejected Option A UX. Both context-menu items share the same `onClicked` listener; the handler distinguishes them by `info.menuItemId`.

### Decision: Click handler merges content-script-resolved context with `info`, preferring resolved fields when present
When the link-context item is clicked, `info.srcUrl` is absent by construction. The handler reads the per-tab stored resolved context (set by the content script's `contextmenu` message) and uses `resolved.srcUrl` if present; if absent (DOM walk failed to find an `<img>`, e.g. site markup changed), the save fails gracefully with the existing "Couldn't save image." toast rather than attempting `info.linkUrl` as if it were an image.

### Decision: Pinterest high-res candidate fetch sends a `Referer` header
Discovered during manual verification (tasks.md 6.2): `i.pinimg.com` rejects `/originals/...` requests with no `Referer` header (403), even though the same requests succeed when made from an actual `pinterest.com` page. `HighResRule` gains an optional `referrer?: string` field, and a new `resolveHighResReferrer(srcUrl): string | undefined` (`extensions/src/lib/highResFetch.ts`) looks up the matching rule's referrer for a given `srcUrl`. `resolveImageBlob` passes it as `fetch(candidateUrl, { referrer })` when present. Only the Pinterest rule sets `referrer: "https://www.pinterest.com/"`; Twitter's rule leaves it unset, so its fetch behavior is unchanged. This is a small extension of the existing `highResRules.ts`/`highResFetch.ts` layer (`extension-highres-image-resolve`), not a new layer — added here because it was a precondition for Pinterest saves to actually succeed once a `srcUrl` could be resolved at all.

## Risks / Trade-offs

- [Pinterest's DOM structure changes, breaking the `closest('a') → querySelector('img')` walk] → Mitigation: failure degrades to the existing "Couldn't save image." error toast (already specified behavior for unfetchable images), not a crash or silent no-op; no worse than today's "doesn't appear at all."
- [Race between `contextmenu` event and `contextMenus.onClicked` if the per-tab store isn't updated before the click fires] → Mitigation: `contextmenu` fires synchronously on right-click, before the native menu even renders; `onClicked` cannot fire before that. Risk is negligible in practice but worth a defensive note: if the stored context is missing or stale, fail to the same error toast rather than reusing an unrelated previous resolution.
- [`info.linkUrl`-based `source_url` override could misfire if a future site wraps an unrelated link around an image (e.g. a lightbox link to the raw image file itself, not a post permalink)] → Mitigation: rule table is allow-listed per site (`matches(url)`), not a generic "any linkUrl present" heuristic — Twitter/Facebook only.
- [New content-script → background message channel is a genuinely new pattern in this codebase] → Mitigation: scoped narrowly (one message type, one direction beyond the existing toast channel), and flagged explicitly here per this repo's decision-boundary convention.

## Open Questions

None outstanding — the user confirmed the two-mechanism split and DOM-walk approach during exploration. Per-tab store implementation detail (in-memory `Map` vs. `chrome.storage.session`) is left to tasks/implementation as it doesn't affect the spec-level contract.
