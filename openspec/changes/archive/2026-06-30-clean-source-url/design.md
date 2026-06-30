## Context

The extension records `source_url` verbatim from the resolved page or link URL. For image-heavy browsing sites (Google Image Search, Twitter, Instagram, Facebook, Pinterest, Reddit, Imgur), this produces URLs with many tracking and session query params that mean nothing to users and inflate token cost when `source_url` is used as LLM context in the suggestion folder feature.

ClearURLs is a well-maintained, community-sourced browser extension whose rule database (`data.min.json`) maps URL patterns to lists of tracking param name regexes to strip. Vendoring a snapshot of this database — rather than maintaining a hand-written param list — means we inherit years of community curation and provide a clear update path for future additions.

All save paths in the extension funnel through `handleSave()` in `background/index.ts` before `pageUrl` is forwarded to `persistImage()` → `saveImage()` → `POST /images` as `source_url`. This is the single injection point for cleaning.

## Goals / Non-Goals

**Goals:**
- Strip tracking/session query params from `source_url` using ClearURLs rules before the URL is persisted.
- Keep the cleaned URL navigable — functional params are not in ClearURLs' rule lists.
- Commit the full ClearURLs `data.min.json` to the repo for developer reference without including it in the bundle.
- Provide a repeatable update path (`make update-clearurls`) for future rule refreshes or provider additions.

**Non-Goals:**
- Cleaning `srcUrl` (the image fetch URL) — it is not persisted.
- Runtime fetching of ClearURLs data — the snapshot is vendored at development time.
- Applying cleaning to `handleCapture()` paths (snip, video frame) — those record the active tab URL which is rarely a high-noise social/search URL.
- URL redirect unwrapping (`redirections` field in ClearURLs data) — out of scope.

## Decisions

### Decision: Gitignored vendor file + committed bundled subset

**Chosen:** `extensions/vendor/clearurls-data.min.json` is fetched locally by `make update-clearurls` but is gitignored — it is a build-time input, not source. Only `extensions/src/lib/clearUrlsProviders.json` (the extracted subset) is committed and imported into the bundle.

**Why gitignored:** The full file (~300KB, 150+ providers) is a fetch artifact. There is no value in committing it — it can always be re-fetched. Committing it would add churn to every rule update PR. The extracted subset is the authoritative artifact; it is small, human-reviewable, and is what actually ships.

**Providers in the subset:** Google, DuckDuckGo, Twitter, Instagram, Facebook, Reddit — the sites users most commonly browse for images that ClearURLs has rules for. Pinterest and Imgur are excluded because ClearURLs does not have provider entries for them.

### Decision: Parse at module load time, not per-call

**Chosen:** On import, compile all `urlPattern`, `exceptions[]`, and `rules[]` + `referralMarketing[]` strings into `RegExp` objects once. Store as a flat array of compiled cleaners. `cleanUrl()` iterates this precompiled array.

**Why:** `new RegExp()` on every `cleanUrl()` call would be wasteful, especially for the background service worker which may call it in rapid succession during picker-save (multiple images). Compiling once on load amortises the cost.

### Decision: Apply all matching providers, not just the first

**Chosen:** Iterate all compiled cleaners; apply every one whose `urlPattern` matches and whose `exceptions` don't veto it.

**Why:** ClearURLs itself applies all matching providers. A URL could match both a site-specific provider (e.g. `google`) and a generic cross-site provider if one exists in the subset. Stopping at the first match would silently miss rules.

### Decision: Apply all three ClearURLs rule fields

**Chosen:** All three relevant ClearURLs fields are applied per provider:
- `rules` and `referralMarketing` are concatenated into a single list of param name patterns to strip (no runtime distinction between them)
- `exceptions` are respected as a per-provider URL-level veto — if any exception regex matches the full URL, the provider is skipped entirely

**Why:** For `source_url` purposes there is no reason to treat referral marketing params differently from tracking params — both are noise. `exceptions` must be respected to avoid stripping params that are actually functional on specific sub-URLs. The `redirections` field (URL redirect unwrapping) is out of scope and not applied.

### Decision: Makefile target using curl + jq

**Chosen:** `make update-clearurls` uses `curl` to fetch `https://rules2.clearurls.xyz/data.min.json` into `vendor/`, then `jq` to extract the 8 providers into `src/lib/clearUrlsProviders.json`.

**Why:** The project already uses Makefile for dev tooling. `curl` and `jq` are standard and available in any Unix dev environment. A Node/TS script would add tooling overhead for a one-off maintenance task.

## Risks / Trade-offs

- **ClearURLs snapshot becomes stale** → New tracking params added by providers aren't stripped until someone runs `make update-clearurls` and commits the updated `clearUrlsProviders.json`. Mitigation: the Makefile target makes this a single command.
- **ClearURLs CDN availability** → `rules2.clearurls.xyz` could be unreliable. Mitigation: this is a development-time-only operation, not runtime. A failed fetch during `make update-clearurls` is obvious and retryable.
- **`completeProvider: true` providers** → If a provider in the subset sets `completeProvider: true`, all query params are stripped. None of the 6 chosen providers currently set this, but it should be handled correctly regardless to avoid a future surprise.
- **Regex correctness** → ClearURLs rule strings are intended as full-match regexes against param names (e.g. `"utm_.*"` matches any param starting with `utm_`). The parser wraps each rule as `^(rule)$` to enforce full-name matching and prevent partial matches.
