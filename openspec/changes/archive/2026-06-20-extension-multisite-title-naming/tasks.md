## 1. Imgur & Instagram alt-text resolver

- [x] 1.1 Create `extensions/src/lib/altTextResolveRule.ts` with `isImgurUrl(pageUrl)`, `isInstagramUrl(pageUrl)`, an exported `findImg(target: Element): HTMLImageElement | null` (returns `target` itself if it's an `<img>`, otherwise `target.closest("a")?.querySelector("img")` to find the image nested in a card-style wrapper, mirroring `resolveCardImageSrc`), and a shared `resolveAltText(target: Element): string | null` that resolves the image via `findImg` and reads its `alt` (return `null` if empty/absent), mirroring the shape of `tweetTextResolveRule.ts`.
- [x] 1.2 Add `extensions/src/lib/altTextResolveRule.test.ts` covering: Imgur URL match/non-match, Instagram URL match/non-match, `resolveAltText` returns the alt string when present, returns `null` when empty/absent, returns `null` when target is not an `<img>` and has no enclosing link, and resolves the nested `<img>`'s alt via `findImg`'s `closest("a")` traversal when the target is a non-`<img>` wrapper (e.g. a grid card).

## 2. Facebook alt/aria-label resolver

- [x] 2.1 Create `extensions/src/lib/facebookAltResolveRule.ts` with `isFacebookUrl(pageUrl)` and `resolveFacebookAltText(target: Element): string | null`: resolves the image via `altTextResolveRule.ts`'s `findImg` and checks its `alt` first; if empty, climbs ancestors from `target` via `parentElement` (bounded by the same `MAX_CLIMB_DEPTH = 40` limit used in `tweetTextResolveRule.ts`) checking each ancestor's own `aria-label` attribute (not a descendant query); returns `null` if nothing found within the bound.
- [x] 2.2 Add `extensions/src/lib/facebookAltResolveRule.test.ts` covering: Facebook URL match/non-match, resolves from the image's own `alt` when present, falls back to an ancestor's `aria-label` when `alt` is empty, returns `null` when neither is found within the climb bound, resolves the nested `<img>`'s alt via the closest enclosing link when the target wraps it, and falls back to ancestor `aria-label` when the wrapped target has no nested img alt.

## 3. Content script wiring

- [x] 3.1 In `extensions/src/content/index.ts`, import the new matchers/resolvers and add `isImgurUrl`/`isInstagramUrl` → `resolveAltText`, and `isFacebookUrl` → `resolveFacebookAltText` branches inside the `contextmenu` listener (alongside the existing Twitter branch), populating `resolved.title` the same way the Twitter branch does.
- [x] 3.2 Widen the early-return gate (currently `!shouldResolveCardDom(...) && !isTwitterUrl(...)`) to also account for the three new site checks, so `resolved` messages are sent for Imgur/Instagram/Facebook pages.

## 4. Background title resolution

- [x] 4.1 In `extensions/src/background/index.ts`, generalize `resolveTitle()`: keep the existing Twitter-handle branch unchanged; add an else-if branch that uses `resolved?.title` verbatim (no truncation/formatting) when present; keep `tab?.title ?? info.pageUrl ?? "Untitled"` as the final fallback.
- [x] 4.2 Update/extend `extensions/src/background/index.test.ts` `resolveTitle()` tests: Imgur resolved title used verbatim, Instagram resolved title used verbatim (both caption-text and "Photo by ... on ..." forms), Facebook resolved title used verbatim, and the no-`resolved.title` case for each still falls back to `tab.title`.

## 5. Spec verification

- [x] 5.1 Manually verify against `openspec/changes/extension-multisite-title-naming/specs/extension-save-image/spec.md` scenarios: save an Imgur image, an Instagram image (both an opened post and a grid/search thumbnail), and a Facebook image (one with its own `alt`, one relying on ancestor `aria-label`, and one with neither) — confirm resulting title matches expectations and `tab.title` fallback still works for non-matching sites.

## 6. Checks

- [x] 6.1 Run `npm run build` in `extensions/` and fix any errors.
- [x] 6.2 Run `npm run lint` in `extensions/` and fix any errors. (Skipped — no lint script/ESLint config exists in `extensions/`; `npm run build`, `type-check`, and `test` cover this package's checks.)
