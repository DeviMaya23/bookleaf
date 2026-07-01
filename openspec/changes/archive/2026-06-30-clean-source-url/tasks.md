## 1. Vendor ClearURLs Data

- [x] 1.1 Create `extensions/vendor/` directory and fetch `https://rules2.clearurls.xyz/data.min.json` into `extensions/vendor/clearurls-data.min.json`
- [x] 1.2 Extract the 8 target providers (`google`, `duckduckgo`, `twitter`, `instagram`, `facebook`, `pinterest`, `reddit`, `imgur`) from the full file and write the subset to `extensions/src/lib/clearUrlsProviders.json` with the same `{ "providers": { ... } }` shape
- [x] 1.3 Add a `update-clearurls` target to the root `Makefile` that re-fetches the full file into `extensions/vendor/clearurls-data.min.json` and re-extracts the subset into `extensions/src/lib/clearUrlsProviders.json` using `curl` and `jq`

## 2. URL Cleaner

- [x] 2.1 Create `extensions/src/lib/urlCleaner.ts` — at module load, import `clearUrlsProviders.json` and compile each provider's `urlPattern`, `exceptions`, `rules`, and `referralMarketing` strings into `RegExp` objects; export `cleanUrl(rawUrl: string): string` that iterates all compiled providers, applies all matching ones (skipping any whose exception matches the URL), strips params per `rules`+`referralMarketing`, handles `completeProvider: true`, and returns the cleaned URL string or the original on parse failure

## 3. Unit Tests

- [x] 3.1 Create `extensions/src/lib/urlCleaner.test.ts` — write unit tests covering: known tracking param stripped, `referralMarketing` param stripped, provider exception vetoes rule application, `completeProvider` removes all params, multiple providers applied when both match, unrecognised host returned unchanged, malformed input returned unchanged

## 4. Wire into Save Flow

- [x] 4.1 In `extensions/src/background/index.ts`, import `cleanUrl` from `../lib/urlCleaner` and apply `cleanUrl(pageUrl)` in `handleSave()` before passing `pageUrl` to `persistImage()`

## 5. Build & Lint

- [x] 5.1 Run `npm run build` from the `extensions/` directory and fix any errors
- [x] 5.2 Run `npm run lint` from the `extensions/` directory and fix any issues
