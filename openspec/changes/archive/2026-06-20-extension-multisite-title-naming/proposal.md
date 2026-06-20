## Why

Today only Twitter saves get a meaningful title (`@handle: tweet text...`); Imgur, Instagram, and Facebook saves fall back to the current tab's title, which is frequently generic or unrelated to the saved image (e.g. just "Instagram", or a feed page title), making saved images hard to find or identify later.

## What Changes

- Extend step 5 of the save flow's title resolution to cover Imgur, Instagram, and Facebook, each using the right-clicked `<img>` element's `alt` text (or an ancestor's `aria-label`, for Facebook/Instagram cases where `alt` is empty) as the title, falling back to `tab.title` when nothing usable is found — same fallback behavior as today.
- Imgur: read `alt` directly off the right-clicked `<img>` (no DOM traversal needed).
- Instagram: read `alt` directly off the right-clicked `<img>`. Content varies by view (full caption text in an opened post vs. `"Photo by <name> on <date>."` in grid/search thumbnails) — both are used as-is, with no parsing or reformatting.
- Facebook: read `alt` off the right-clicked `<img>`; if empty, search upward through ancestors (same bounded traversal pattern used for Twitter's tweet-text resolution) for the nearest `aria-label`. This text is Facebook's AI-generated image description (e.g. "May be an image of gelato and text"), not the human-written post caption — caption scraping was explored and explicitly ruled out as too fragile (no stable selector, Facebook's post-body markup churns frequently).
- No new abstraction is introduced: each site's resolution is added as its own branch in the existing per-site title logic, mirroring how Twitter's case is already handled today (not refactored into a shared rule registry).

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `extension-save-image`: title resolution (step 5 of the Authenticated save flow) gains additional branches for Imgur, Instagram, and Facebook saves, each described above. The existing Twitter behavior and the `tab.title` fallback for unmatched cases are unchanged.

## Impact

- `extensions/src/background/index.ts` — `resolveTitle()` gains new branches for Imgur/Instagram/Facebook resolved context.
- `extensions/src/content/index.ts` — context-menu handler gains per-site alt/aria-label resolution calls, sent to background via the existing `resolved.title` messaging shape.
- New site-specific resolver files mirroring `extensions/src/lib/tweetTextResolveRule.ts` (e.g. for Imgur/Instagram alt read, and Facebook's upward aria-label search).
- No backend, API, or database changes.
