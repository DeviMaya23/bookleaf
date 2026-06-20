## Context

`cardDomResolveRules.ts` already solves this exact DOM shape for Pinterest: a `CardDomResolveRule` table gated by page URL (`shouldResolveCardDom`), a content-script `contextmenu` listener (capture phase) that calls `resolveCardImageSrc(target)` — which climbs to the nearest `<a>` ancestor and reads its first descendant `<img>`'s `src` — and a parallel `linkOnlyCardUrlPatterns` list that drives a second `contexts: ["link"]` context-menu item registered with `targetUrlPatterns` in `background/index.ts`. Confirmed (per proposal) that Instagram's grid/profile view has the identical shape: the right-click target is the post's `<a>` wrapper, with an `<img>` descendant, not the `<img>` itself.

## Goals / Non-Goals

**Goals:**
- Make "Save to Bookleaf" appear and work on Instagram grid/profile post thumbnails, using the existing card-resolution mechanism unchanged.
- Ensure the captured `srcUrl` is the `<img>`'s `src` attribute, not a `srcset`-resolved variant — already guaranteed by `resolveCardImageSrc` reading `.src`, not `.currentSrc`.

**Non-Goals:**
- No `highResRules.ts` changes — Instagram's `src` is already the largest available variant; no URL upgrade needed.
- No support for the main scrolling feed view, carousel (multi-image) posts, Reels, or video posts — only confirmed against single-image grid/profile thumbnails. These may or may not share the same `<a>`-wrapped-`<img>` shape; not claimed to work, not blocking.
- No changes to `resolveCardImageSrc`'s climb strategy (`closest("a")` then first descendant `img`) — reused as-is, same as Pinterest.

## Decisions

**Reuse `resolveCardImageSrc` and `shouldResolveCardDom` verbatim — add only a new `CardDomResolveRule` table entry.** The Pinterest implementation already generalizes the exact shape Instagram needs (link-wrapped image, no direct `<img>` click target). Alternative considered: writing an Instagram-specific resolver — rejected, no behavioral difference identified that would justify diverging from the existing generic climb.

**Add `*://*.instagram.com/p/*` to `linkOnlyCardUrlPatterns`.** This is Instagram's single-post permalink path shape (`instagram.com/p/{shortcode}/`), the equivalent of Pinterest's `/pin/*`. Reels (`/reel/*`) and profile/grid container URLs are not included — the pattern only needs to match the `<a>` `href` on the grid thumbnail itself (which always points to `/p/{shortcode}/` for image posts), not the page the user is currently on.

## Risks / Trade-offs

- **[Risk] Other Instagram surfaces (main feed, carousel posts, Reels) may not share the grid view's `<a>`-wrapped-`<img>` shape** → **Mitigation**: `shouldResolveCardDom` only activates card-DOM resolution when on an `instagram.com` page at all (page-level gate, not post-type-level), so on those other surfaces the existing native `contexts: ["image"]` menu item still applies wherever the click target genuinely is an `<img>` — this change only adds coverage for the grid case, it doesn't regress anything elsewhere. Unconfirmed surfaces are left as a known gap, not a failure mode.
- **[Risk] `instagram.com` page-level gate is coarse — `shouldResolveCardDom` returning true for any Instagram page could swallow a legitimate native-image right-click result if a future Instagram layout puts the click target back on a real `<img>`** → **Mitigation**: `resolveCardImageSrc` only sends a resolved context when it actually finds a wrapping `<a>` with a descendant `<img>`; if the click target is already an `<img>` outside that shape, the native context-menu `info.srcUrl` path still takes precedence per the existing `extension-save-image` flow (`info.srcUrl` is checked first, before any resolved-context fallback) — same protection Pinterest already relies on.
