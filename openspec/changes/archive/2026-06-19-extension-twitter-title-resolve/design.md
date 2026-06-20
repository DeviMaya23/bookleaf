## Context

The "Save to Bookleaf" extension titles every saved image with `tab?.title` (`background/index.ts:44`), regardless of site. For Twitter/X this is generic and often stale (SPA navigation doesn't reliably update `document.title`). The `extension-card-context-resolve` capability already typed the resolved-context payload as `Partial<{ srcUrl: string; title: string }>` in anticipation of this, but `title` has never been populated.

Two existing mechanisms are relevant and need to be distinguished:
- **DOM-walk resolution** (`cardDomResolveRules.ts`, `content/index.ts:133-143`): fires on `contextmenu` capture, only for sites in the `rules` table (currently Pinterest only), and only sends `resolved` if a value was found.
- **Native context-menu fields** (`info.srcUrl`, `info.linkUrl`): for Twitter, right-clicking an image uses `contexts: ["image"]`, so the browser already supplies `info.srcUrl` natively — no DOM walk is needed to get the image source. `info.linkUrl` is also already populated and used today to override `source_url` (`background/index.ts:49-50`), because Twitter wraps tweet media in an `<a>` pointing at the status permalink. This is why `source_url` works correctly even when saving from a scrolling timeline, where `pageUrl` is just `x.com/home`.

Title resolution needs to combine both: tweet **text** must come from the DOM (no API/network access), while the **handle** should come from `info.linkUrl` (already reliable per-tweet), not from `pageUrl`/`tab.title` (only reliable on the tweet's own permalink page).

## Goals / Non-Goals

**Goals:**
- Resolve a Twitter-specific title (`@handle: <tweet text>...` or `@handle`) at right-click time, working from any Twitter surface (timeline, profile, permalink), not just the tweet's own page.
- Reuse the existing DOM-walk and link-permalink infrastructure rather than introducing a new mechanism.

**Non-Goals:**
- No change to Pinterest's existing srcUrl resolution behavior.
- No network/API calls (no oEmbed, no syndication API) — DOM-only.
- No distinct handling for quote-tweets vs. retweets — nearest `tweetText` to the click target is used uniformly.
- No change to non-Twitter title behavior.

## Decisions

**1. Decouple the DOM-walk "trigger" from "found a srcUrl".**
Today, `content/index.ts:138-140` only calls `runtime.sendMessage` if `resolveCardImageSrc` returns a value. For Twitter we only need `title`, never `srcUrl` (the browser already supplies `srcUrl` natively for `contexts:["image"]`). The content script's contextmenu handler is restructured to resolve each field independently and send `resolved` if *any* field was found, rather than gating the whole message on `srcUrl`.

**2. Add a Twitter DOM-walk rule for `title`, generalized as ancestor-climb instead of single-selector lookup.**
Pinterest's existing walk (`resolveCardImageSrc`) is a single `closest("a")` + `querySelector("img")` — appropriate because Pinterest cards have one clear container per image. Twitter's nesting (quote-tweets embed another tweet-like card) means a single fixed selector (e.g. `closest('article')`) doesn't reliably name "the nearest tweet" — quoted/embedded tweets aren't always marked with the same tag as the outer tweet.

Instead: climb from the click target one ancestor at a time, and at each step run `querySelector('[data-testid="tweetText"]')` scoped to that ancestor; return the first match. This naturally resolves to whichever `tweetText` is structurally nearest to the clicked image — the smallest enclosing container wins — which gives "original/quoted post text" when the image is inside a quoted-tweet embed, and the outer tweet's text otherwise, with no explicit quote/retweet branching required.

Bounded by a max-ancestor-depth (e.g. 10) to avoid walking to `<body>` on pages with no matching structure.

**3. Handle parsed from `info.linkUrl`, not `pageUrl`/`tab.title`.**
A new export alongside `linkPermalinkRules.ts` extracts the handle from a matched Twitter status URL's pathname (segment before `/status/`). This is computed in the background script at click time (`handleContextMenuClick`), using the same `info.linkUrl` already validated by `resolveLinkPermalink`. If `info.linkUrl` doesn't match the Twitter status pattern (e.g. right-click target has no enclosing permalink anchor), no Twitter-specific title is built — falls back to existing `tab?.title` behavior.

**4. Title assembly happens in the background script, not the content script.**
The content script only resolves and sends raw `tweetText`. The background script combines `resolved.title` (raw tweet text) with the handle (parsed from `info.linkUrl`) and applies the truncation/format rule. This keeps the DOM-walk rule single-purpose (resolve text, nothing else) and keeps formatting logic (truncation, `@handle:` prefix, fallback) in one place alongside the existing title-assembly line.

**5. Truncation: 100 chars of tweet text, hard cut + `...` suffix; fallback `@handle` with no colon.**
`title = tweetText ? `@${handle}: ${tweetText.slice(0, 100)}...` : `@${handle}``. No special-casing for tweets shorter than 100 chars needing no ellipsis — out of scope per "100 chars + `...`" as given; if this reads oddly for short tweets it can be revisited later.

## Risks / Trade-offs

- **[Risk]** Twitter's `data-testid` attributes are not a public API and can change without notice → **Mitigation**: same risk already accepted for `highResRules.ts`'s reliance on `pbs.twimg.com` URL shape and `linkPermalinkRules.ts`'s status-path regex; failure mode is graceful (falls back to `@handle` or `tab?.title`), not a crash.
- **[Risk]** Ancestor-climb with unbounded depth could be slow on deeply nested pages → **Mitigation**: cap climb depth (tuned to 40 ancestors after testing showed Twitter's actual wrapper-div nesting between an image and its enclosing tweet exceeds a depth of 10).
- **[Risk]** `info.linkUrl` may be absent if the right-clicked image isn't wrapped in a status-permalink anchor (e.g. some video/gif players) → **Mitigation**: falls back to `tab?.title`, same as today.
- **[Known limitation, accepted]** The photo/media lightbox view (`x.com/.../status/.../photo/1`) has no `[data-testid="tweetText"]` element in its DOM at all — confirmed via manual testing, not just an ancestor-climb miss. Saves from this view always resolve to `@handle` with no tweet text, which is accepted as-is (the `source_url` still correctly points at the tweet permalink).
- **[Bug fixed during testing]** The original content-script implementation only sent a `resolved` message when at least one field was found, which meant a right-click that resolved nothing (e.g. the photo lightbox) left the previous right-click's stale `resolved.title` in `resolvedContextByTab` for that tab. Fixed by always sending `resolved` (even empty) for any registered site, so each right-click overwrites the prior one per the "most-recent wins" requirement.

## Migration Plan

No migration — purely additive behavior change in the extension, ships in the next extension release. No data model or backend changes. Rollback is reverting the extension build.
