## Why

Right-clicking an Instagram grid/profile post does not show "Save to Bookleaf" today, because the click target is the post's enclosing `<a>` link rather than the `<img>` itself — the same DOM shape that required dedicated card-detection support for Pinterest. Without it, Instagram images can't be saved at all via the extension.

## What Changes

- Register Instagram in the existing card-resolution mechanism: add an `instagramRule` to `cardDomResolveRules.ts` (matching `instagram.com` and subdomains) and add Instagram's post-permalink pattern (e.g. `*://*.instagram.com/p/*`) to `linkOnlyCardUrlPatterns`, mirroring the existing Pinterest entries.
- This reuses `resolveCardImageSrc()` unchanged (`closest("a")` → `querySelector("img")` → `.src`) and the existing link-only context-menu item / content-script messaging path — no new resolution logic.
- No changes to `highResRules.ts`: Instagram's plain `<img src>` (as opposed to its `srcset`) is already the highest-resolution variant available in the DOM, confirmed by comparing the `stp` transform segment on `src` (no size constraint) against `srcset` entries (which carry an explicit size constraint like `p1080x1080`). `resolveCardImageSrc()` already reads the `src` attribute, not the browser-resolved `currentSrc`, so no extra extraction logic is needed to avoid picking up a downscaled `srcset` candidate.

## Capabilities

### Modified Capabilities
- `extension-card-context-resolve`: add Instagram to the card-level DOM resolution site rule table and to the link-only card URL patterns, alongside the existing Pinterest entry.

## Impact

- `extensions/src/lib/cardDomResolveRules.ts`: add `instagramRule` to `rules`, add an Instagram pattern to `linkOnlyCardUrlPatterns`.
- No changes to `extensions/src/background/index.ts` (menu registration already derives `targetUrlPatterns` from `linkOnlyCardUrlPatterns`), `extensions/src/content/index.ts` (already generic over `shouldResolveCardDom`/`resolveCardImageSrc`), or `highResRules.ts`.
