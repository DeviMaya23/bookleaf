## 1. Imgur rule

- [x] 1.1 Add `imgurRule: HighResRule` to `extensions/src/lib/highResRules.ts`: `matches` checks `url.hostname === "i.imgur.com"` and the path fits `{id}_{letter}.{ext}` (ext in webp/jpg/jpeg/png/gif); excludes `.mp4`/`.gifv` paths.
- [x] 1.2 Implement `imgurRule.transform`: strip the size-letter suffix and query string, return `https://i.imgur.com/{id}.jpg`.
- [x] 1.3 Add `imgurRule` to the exported `rules` array.

## 2. Facebook rule

- [x] 2.1 Add `facebookRule: HighResRule` to `extensions/src/lib/highResRules.ts`: `matches` checks `url.hostname` ends with `.fbcdn.net` (or equals `fbcdn.net`) and `url.searchParams.has("ctp")`.
- [x] 2.2 Implement `facebookRule.transform`: delete only the `ctp` param via `URLSearchParams.delete`, return the resulting URL string unchanged otherwise.
- [x] 2.3 Add `facebookRule` to the exported `rules` array.

## 3. Unit tests

- [x] 3.1 In `extensions/src/lib/highResRules.test.ts`, add `describe("imgur-thumbnail rule")` covering: matches a `{id}_{letter}.webp` thumbnail URL; does not match a bare id with no size suffix; does not match a `.mp4`/`.gifv` video URL; transform converts `xCbCj7a_d.webp?maxwidth=520&shape=thumb&fidelity=high` to `https://i.imgur.com/xCbCj7a.jpg`.
- [x] 3.2 In the same file, add `describe("facebook-ctp rule")` covering: matches an `fbcdn.net` URL containing `ctp`; does not match an `fbcdn.net` URL without `ctp`; transform removes only `ctp` and leaves `cstp`/`oh`/`oe`/`_nc_*` params byte-for-byte unchanged.

## 4. Verification

- [x] 4.1 Run `npm run build` in `extensions/` and fix any errors.
- [x] 4.2 Run `npm run lint` in `extensions/` and fix any errors. (No `lint` script exists in `extensions/`; ran `npm run type-check` instead — passed with no errors.)
- [x] 4.3 Run the extension test suite (`npm test` / vitest) in `extensions/` and confirm all tests pass.
