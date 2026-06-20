## Why

The "Save to Bookleaf" extension currently titles every saved image with the browser tab's title (`tab?.title`), regardless of site. For Twitter/X, this produces a generic, often-stale title (SPA navigation doesn't always update `document.title`) instead of anything tied to the actual tweet. The `extension-card-context-resolve` capability already anticipated this — its resolved-context payload is typed as `Partial<{ srcUrl: string; title: string }>` — but title extraction was explicitly deferred. This change implements it for Twitter.

## What Changes

- Add a new DOM-resolve rule for Twitter/X: on right-click, walk up from the click target to the nearest `[data-testid="tweetText"]` in scope and resolve it as `title` in the existing resolved-context payload.
- In the background script's title assembly (`background/index.ts`), add a Twitter-specific branch: when the URL matches the existing Twitter status pattern (`linkPermalinkRules.ts`), build the title as `@handle: <tweet text, truncated to 100 chars>...`, where `handle` is parsed from the URL.
- If no `tweetText` is resolved (image-only tweet, or a retweet with no local text), fall back to `@handle` alone.
- No special-casing for quote-tweet vs. retweet nesting — the DOM-walk naturally resolves to whichever `tweetText` is structurally nearest to the right-clicked image.
- Non-Twitter sites are unaffected; they continue to use `tab?.title`.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `extension-card-context-resolve`: the resolved-context payload's `title` field — previously typed but never populated — is now populated for Twitter/X via a new DOM-resolve rule.
- `extension-save-image`: the title used when saving an image now has a Twitter-specific derivation (`@handle: tweet text`) instead of always using `tab?.title`.

## Impact

- `extensions/src/lib/cardDomResolveRules.ts` (or a new sibling rule file): add Twitter tweetText resolve rule.
- `extensions/src/content/index.ts`: wire the new rule into the existing right-click DOM-walk flow so `resolved.title` gets populated for Twitter.
- `extensions/src/background/index.ts`: title assembly logic (currently line ~44, `tab?.title ?? pageUrl ?? "Untitled"`) gains a Twitter-specific branch using `resolved.title` + handle parsed via the existing regex in `linkPermalinkRules.ts`.
- No backend or frontend (web app) changes. No new permissions or dependencies.
