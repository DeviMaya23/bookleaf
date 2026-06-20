## Context

Today's title resolution (`extensions/src/background/index.ts:39-49`, `resolveTitle()`) has exactly one special case: if the clicked link matches the Twitter status permalink rule, it formats `@<handle>: <resolved tweet text>...` using text the content script extracted via `resolveTweetText()` (`extensions/src/lib/tweetTextResolveRule.ts`), which climbs up to 40 ancestor levels from the click target looking for `[data-testid="tweetText"]`. Everything else — including Imgur, Instagram, and Facebook today — falls through to `tab?.title ?? info.pageUrl ?? "Untitled"`.

The content script (`extensions/src/content/index.ts:134-155`) only sends a `resolved` message to the background at all if the page is a Twitter URL or matches a card-DOM site (`shouldResolveCardDom`); otherwise the `contextmenu` listener returns early (line 151) without messaging. This gate needs to widen to include the three new sites, or their resolved alt text never reaches the background.

Investigation during exploration (manual DOM inspection) found:
- **Imgur**: the right-clicked `<img>` itself carries a usable `alt` attribute directly in the cases inspected.
- **Instagram**: an opened post's right-click target is the `<img>` itself, with `alt` content varying by view (full caption text in an opened post vs. `"Photo by <name> on <date>."` in grid/search thumbnails) — both accepted as-is. However, grid/search thumbnails wrap the `<img>` in card markup (e.g. a `<ul>`/`<li>` structure inside an `<a>`), so the `contextmenu` event's actual target is a non-`<img>` wrapper, not the image — discovered when manual testing showed grid-thumbnail saves falling back to `tab.title` instead of resolving the alt text. Fixed by reusing the same `closest("a")` → `querySelector("img")` traversal `resolveCardImageSrc` (`cardDomResolveRules.ts`) already uses for `srcUrl` resolution on these same card sites.
- **Facebook**: inconsistent — sometimes the `<img>`'s own `alt` is populated, sometimes it's empty and the description instead lives on an *ancestor* `<a aria-label="...">` (observed up to 2 levels up). The text itself is Facebook's AI-generated image description ("May be an image of gelato and text"), not the human-written post caption. Scraping the real caption was explored and ruled out — no stable selector exists for it, and Facebook's post-body markup is known to churn.

## Goals / Non-Goals

**Goals:**
- Imgur, Instagram, and Facebook saves get a more useful title than `tab.title` when a usable `alt`/`aria-label` is available.
- Reuse the existing fallback contract: if nothing useful resolves, behavior is identical to today (`tab.title`).
- Keep the implementation shape consistent with how Twitter's case already works (per-site resolver file + URL matcher + a branch in `resolveTitle()`), per the project's existing pattern — no new shared rule-registry abstraction introduced.

**Non-Goals:**
- Scraping Facebook's actual human-written post caption (ruled out as too fragile; out of scope).
- Parsing/reformatting Instagram's `"Photo by <name> on <date>."` variant into a Twitter-style `@handle: text` shape — used verbatim.
- Any change to the high-res image resolution, source_url, or upload flow — this only touches the `title` value.

## Decisions

**1. Two new resolver files, not one shared "alt resolver" abstraction spanning all three sites.**
- `extensions/src/lib/altTextResolveRule.ts`: `isImgurUrl`, `isInstagramUrl`, and a single `resolveAltText(target: Element): string | null` shared by both — both sites need the identical "read the relevant `<img>`'s `alt`" logic, so one helper function covers both call sites. (Sharing here is a plain function reuse, not a new abstraction layer — same granularity as `tweetTextResolveRule.ts` having one function for one site.) `resolveAltText` delegates to an exported `findImg(target: Element): HTMLImageElement | null` helper: if `target` is already an `<img>`, use it directly; otherwise fall back to `target.closest("a")?.querySelector("img")` to find the actual image nested inside a card-style wrapper (the same traversal `resolveCardImageSrc` uses for `srcUrl`). This was added after manual testing showed Instagram grid/search thumbnails right-click on a non-`<img>` wrapper element.
- `extensions/src/lib/facebookAltResolveRule.ts`: `isFacebookUrl` and `resolveFacebookAltText(target: Element): string | null`, which calls the same `findImg` helper to check the resolved image's own `alt` first; if empty, climbs ancestors *from `target`* (bounded the same way as `resolveTweetText`, reusing `MAX_CLIMB_DEPTH = 40` as the limit) checking each ancestor's `aria-label` attribute directly (not a descendant `querySelector`, since the attribute sits on the ancestor itself, unlike Twitter's tweet-text container which is found by querying down into a descendant).
- Alternative considered: a generic `TitleResolveRule[]` registry (mirroring `HighResRule`'s shape) covering all four sites including Twitter. Rejected for this change — flagged to the user as a real fork in the road; decided to keep matching Twitter's existing ad-hoc per-site-branch shape rather than introduce a new pattern, per the "stop and propose before introducing a new abstraction" rule.

**2. Widen the content script's send-gate (`content/index.ts:151`) to include the three new site checks**, so a `resolved.title` is actually sent to the background for these sites. Without this, the new resolver functions would have no path to reach `resolveTitle()`.

**3. Generalize `resolveTitle()`'s fallback branch, rather than add three near-identical `if` blocks.**
Today: Twitter handle present → format with `@handle`; else → `tab.title`. New: Twitter handle present → format with `@handle` (unchanged); else if `resolved?.title` is present (set by content script only for matched sites) → use it verbatim; else → `tab.title` (unchanged). This works because the content script's site-gating already ensures `resolved.title` is only populated by the site whose URL matched — `resolveTitle()` doesn't need to re-derive which site it came from.

## Risks / Trade-offs

- **[Risk] Facebook's `aria-label` text is an AI-generated visual description, not what the user actually posted** → Mitigation: explicitly scoped as accepted/non-goal in this design; still strictly better than `tab.title`, and documented as a known limitation rather than presented as caption parity with Twitter/Instagram.
- **[Risk] Imgur/Instagram images with empty or unhelpful `alt` (e.g., default-generated, no caption set)** → Mitigation: existing fallback path (`tab.title`) already handles "nothing resolved," same as today's behavior for every other site — no new failure mode introduced.
- **[Risk] Facebook's DOM structure for where `aria-label` lives could shift over time (the 2-ancestor-level observation was from manual inspection of current markup, not guaranteed stable)** → Mitigation: bounded climb (reuses `MAX_CLIMB_DEPTH`) degrades to the existing `tab.title` fallback if the attribute isn't found within range — failure is silent and non-breaking, consistent with how a missing tweet-text container already degrades today.
- **[Trade-off] Not introducing a shared `TitleResolveRule` registry now means a 4th similarly-shaped site in the future would mean a 4th near-duplicate resolver file and another `resolveTitle()` branch** → Accepted for this change; flagged in proposal as a deliberate "don't introduce new abstraction without asking" decision, revisit if a future site makes the duplication cost clearly worth it.

## Migration Plan

No data migration, no backend/API changes, no feature flag. This is a pure behavior widening in the extension: previously-unhandled sites gain a resolution path, and the existing fallback to `tab.title` is preserved as the worst case for all three. Ship directly; rollback is a plain revert if needed.

## Open Questions

None outstanding — all unresolved questions from the design conversation (scope split, Facebook caption feasibility, Instagram format variance) were settled before this document was written.
