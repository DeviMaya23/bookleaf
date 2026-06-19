## Why

Only Twitter/X and Pinterest currently get upgraded to a full-resolution image when saving via the extension; Imgur and Facebook images are saved at thumbnail/CDN-display resolution. Both sites' thumbnail URLs can be algorithmically rewritten to a full-resolution candidate URL without any DOM inspection, so they fit the existing `HighResRule` table used by Twitter/Pinterest.

## What Changes

- Add an Imgur high-res rule: detect `i.imgur.com` thumbnail URLs (suffixed image id + size letter, e.g. `xCbCj7a_d.webp`), strip the size suffix and extension, and re-request the bare id with a non-webp extension to get the unscaled original asset.
- Add a Facebook high-res rule: detect `fbcdn.net` image URLs carrying a `ctp` query param and strip only that param, leaving `stp`/`cstp`/the signature-bearing `_nc_*`/`oh`/`oe` params untouched.
- Both rules plug into the existing `resolveHighResUrl` table and existing candidate-validation pipeline (status, content-type allowlist, dimension check) — no changes to the save flow, validation logic, or title resolution.
- Imgur video posts (`.gifv`/`.mp4`) are explicitly excluded from the rule's `matches` predicate — they fall through to the unmodified original URL, same as any other unmatched site.

## Capabilities

### Modified Capabilities
- `extension-highres-image-resolve`: add Imgur and Facebook entries to the high-resolution URL rule table, alongside the existing Twitter and Pinterest rules.

## Impact

- `extensions/src/lib/highResRules.ts`: add `imgurRule` and `facebookRule` to the `rules` array.
- No changes to `extensions/src/lib/highResFetch.ts`, `extensions/src/background/index.ts`, or the candidate-validation logic in `extension-save-image` — both new rules consume the existing dispatch and validation mechanisms unchanged.
- No changes to `extensions/src/lib/linkPermalinkRules.ts` (the existing `facebookRule` there is for `source_url` permalink detection, a separate concern from high-res resolution).
