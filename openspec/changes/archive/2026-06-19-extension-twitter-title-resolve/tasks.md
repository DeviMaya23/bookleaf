## 1. Tweet text DOM-resolve rule

- [x] 1.1 Add a Twitter tweet-text resolve rule (new file alongside `cardDomResolveRules.ts`, e.g. `tweetTextResolveRule.ts`): climbs ancestors from a given element (bounded depth, e.g. 10) and returns the text content of the first `[data-testid="tweetText"]` descendant found at the nearest ancestor level, or `null` if none found within the bound.
- [x] 1.2 Add a Twitter/X hostname matcher (reuse `TWITTER_HOSTNAMES` shape from `linkPermalinkRules.ts`) to gate when the rule applies.
- [x] 1.3 Unit test: image directly inside a tweet resolves that tweet's `tweetText`.
- [x] 1.4 Unit test: image inside an embedded quoted-tweet card resolves the quoted tweet's `tweetText` (nearer ancestor), not the outer quoting tweet's text.
- [x] 1.5 Unit test: no `tweetText` descendant within climb-depth bound resolves `null`.
- [x] 1.6 Unit test: climb-depth bound is respected (a matching element further up than the bound is not found).

## 2. Content script: independent title resolution

- [x] 2.1 Restructure the `contextmenu` listener in `content/index.ts` so `srcUrl` and `title` are resolved independently (not gated on each other), and `resolved` is sent if either field is present.
- [x] 2.2 Wire the Twitter tweet-text rule into this flow, gated by the Twitter/X hostname matcher from 1.2.
- [~] 2.3 Skipped — would require adding `jsdom` as a new dev dependency (content script has DOM side effects on import; project test env is `node`). User decided to skip rather than add the dependency.
- [~] 2.4 Skipped, same reason as 2.3. Underlying logic remains covered by lib-level unit tests (`tweetTextResolveRule.test.ts`, `cardDomResolveRules.test.ts`).

## 3. Handle extraction

- [x] 3.1 Add a handle-extraction export alongside `linkPermalinkRules.ts` (e.g. `extractTwitterHandle(url: URL): string | null`) that parses the path segment preceding `/status/` from a URL already matched by the Twitter status permalink rule.
- [x] 3.2 Unit test: extracts handle from `https://x.com/username/status/123456789/photo/1`.
- [x] 3.3 Unit test: returns `null` for a non-Twitter or non-status URL.

## 4. Title assembly in background script

- [x] 4.1 In `handleContextMenuClick` (`background/index.ts`), for the `"save-to-bookleaf"` (image) branch, read `resolvedContextByTab` for the tab's `title` (currently only the `"save-to-bookleaf-link"` branch reads resolved context).
- [x] 4.2 When `info.linkUrl` matches the Twitter status permalink rule: build title as `@<handle>: <tweetText.slice(0,100)>...` if `resolved.title` is present, else `@<handle>` alone.
- [x] 4.3 When `info.linkUrl` does not match (or is absent), keep existing `tab?.title ?? pageUrl ?? "Untitled"` behavior.
- [x] 4.4 Unit test: Twitter image save with resolved tweet text produces `@handle: <truncated text>...`.
- [x] 4.5 Unit test: Twitter image save with no resolved tweet text produces `@handle` (no colon).
- [x] 4.6 Unit test: Twitter image save with tweet text under 100 chars still appends `...` (per truncation rule as specified) — confirms no special-casing was added.
- [x] 4.7 Unit test: non-Twitter image save is unaffected (`tab?.title` used, same as before this change).

## 5. Verification

- [x] 5.1 Run `npm run test` in `extensions/` and fix any failures.
- [x] 5.2 Run `npm run type-check` in `extensions/` and fix any type errors.
- [x] 5.3 Run `npm run build` (and `build:firefox`) in `extensions/` and fix any build errors.
- [x] 5.4 Manual smoke test in Firefox: confirmed on real tweets. Tweet-detail view resolves `@handle: text...` correctly. Photo/media lightbox view has no `tweetText` in its DOM at all (confirmed via devtools) — falls back to `@handle`, accepted as a known limitation. Found and fixed a stale-title bug along the way (see design.md Risks/Trade-offs).
