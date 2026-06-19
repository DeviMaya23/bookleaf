## Why

The "Save to Bookleaf" right-click flow derives `srcUrl`, `source_url`, and `title` from page-level browser context (`info.srcUrl`, `info.pageUrl`, `tab.title`). This breaks on card-based feed/grid UIs where the right-clicked element represents one item among many: on Twitter/X's media-tab grid and on Facebook, the saved `source_url` is the generic listing page instead of the actual post permalink (even though the post permalink is available as `info.linkUrl` today, just unread). On Pinterest, the grid card's clickable surface is a `<div>`/`<a>` wrapper around the `<img>`, so the browser's native "image" context menu never fires at all — "Save to Bookleaf" doesn't appear, forcing users to open the pin's own page first just to save it.

## What Changes

- Read `info.linkUrl` in the existing context-menu click handler and override `source_url` with it when the active tab's site has a registered card-context rule (Twitter, Facebook) — no DOM access needed, since `srcUrl` is already populated for these sites.
- Add a content script `contextmenu` listener (capture phase) that, for sites with a registered DOM-resolution rule (Pinterest), walks from the right-clicked element to find the card's actual `<img>` and messages the resolved data to the background service worker. **BREAKING** (internal only): introduces a new content-script → background message channel; today messages only flow background → content (toasts).
- Add a second `contextMenus.create` item scoped to `contexts: ["link"]` with `targetUrlPatterns` matching recognized permalink shapes (e.g. Pinterest pin URLs), so "Save to Bookleaf" appears on these link-only cards without showing on arbitrary links.
- Background correlates the content script's per-tab resolved context with the subsequent `contextMenus.onClicked` event (the `contextmenu` DOM event fires before the menu item click), since there is no other way to join DOM-derived data with the click handler.
- The content-script message payload is a partial bag of card-level fields (only `srcUrl` populated by this change) rather than a single-purpose `resolvedSrcUrl`, so that a future `title` resolver (e.g. Twitter's "@username — post text", explicitly out of scope here) can reuse the same channel without renegotiating its shape.
- The existing `highResRules.ts` URL-rewrite layer (Twitter `name=orig`, Pinterest `236x`→`originals`) is unchanged — it already produces a high-res URL from whatever `srcUrl` this change resolves.

## Capabilities

### New Capabilities
- `extension-card-context-resolve`: site-specific rules and resolution mechanism (linkUrl override, content-script DOM-walk) that determine the effective `srcUrl` and `source_url` for a right-clicked card, before the existing high-res URL-rewrite and save flow run.

### Modified Capabilities
- `extension-save-image`: the context-menu click handler now reads `info.linkUrl` and (for Pinterest-shaped sites) a content-script-resolved `srcUrl`, instead of relying solely on `info.srcUrl` and `info.pageUrl`. A second context-menu item is registered for link-only card sites.

## Impact

- `extensions/src/background/index.ts`: context-menu registration (new second item), `contextMenus.onClicked` handler (read `linkUrl`, merge content-script-resolved context), new per-tab resolved-context store.
- `extensions/src/content/index.ts`: new `contextmenu` listener and outbound message, alongside existing inbound toast listener.
- `extensions/src/lib/` : new module for the card-context rule table (site `matches()` + resolution behavior), parallel to but separate from `highResRules.ts`.
- No backend or frontend (web app) changes — extension-only.
