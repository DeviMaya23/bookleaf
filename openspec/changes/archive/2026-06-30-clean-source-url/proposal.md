## Why

When saving an image, the recorded `source_url` contains tracking and session query params that are meaningless to users and add token cost when the URL is used as LLM context in the suggestion folder feature. The biggest offenders are Google Image Search URLs (15–25 noise params), but Twitter, Instagram, Facebook, Pinterest, Reddit, and others exhibit the same problem. Rather than maintaining a hand-written list of params to strip, this change vendors a snapshot of the [ClearURLs](https://gitlab.com/ClearURLs/rules) rule database — a well-maintained, community-sourced list of tracking params per provider — and uses it to clean `source_url` before it is persisted.

## What Changes

- The full ClearURLs `data.min.json` is fetched into `extensions/vendor/clearurls-data.min.json` as a local reference artifact. This file is gitignored and never imported into the build.
- A curated 6-provider subset (Google, DuckDuckGo, Twitter, Instagram, Facebook, Reddit) is extracted from the full file and committed as `extensions/src/lib/clearUrlsProviders.json`. This file is imported by the extension and bundled. Pinterest and Imgur are excluded — ClearURLs does not have rules for them.
- A Makefile target (`update-clearurls`) re-fetches the full file and re-extracts the subset, so future provider additions or rule updates follow the same workflow.
- A `cleanUrl(rawUrl: string): string` function in `extensions/src/lib/urlCleaner.ts` parses the vendored subset at module load time (compiling regexes once), then strips matching query params from a given URL. All matching providers are applied.
- `handleSave()` in `background/index.ts` calls `cleanUrl(pageUrl)` before forwarding to `persistImage()`, covering all save paths (context menu, drag-and-drop, image picker).
- The cleaned URL remains fully navigable — only tracking/session params defined by ClearURLs are removed.

## Capabilities

### New Capabilities

- `extension-source-url-cleaning`: Vendored ClearURLs provider subset, the parser that compiles and applies those rules, and the `cleanUrl()` function.

### Modified Capabilities

- `extension-save-image`: The `source_url` persisted to the backend is now the ClearURLs-cleaned form of the resolved URL, not the raw URL.

## Impact

- **Extension only** — no backend or frontend changes.
- New files: `extensions/src/lib/clearUrlsProviders.json`, `extensions/src/lib/urlCleaner.ts`, `extensions/src/lib/urlCleaner.test.ts`. `extensions/vendor/clearurls-data.min.json` is generated locally but gitignored.
- Modified files: `extensions/src/background/index.ts` (one `cleanUrl()` call in `handleSave()`), `extensions/Makefile` (new `update-clearurls` target).
- No new npm dependencies.
