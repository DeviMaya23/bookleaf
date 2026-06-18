## Why

Several popular sites (Twitter/X, Pinterest) serve downscaled thumbnail URLs by default for the `<img>` elements users right-click to save. Today the extension saves exactly the URL it finds (`info.srcUrl`), so users end up with a low-resolution copy in their Bookleaf library even though a full-resolution original is available at a predictable, site-specific URL. Resolving this automatically removes a manual "find the original" step from the save flow.

## What Changes

- Add a site-specific URL rewrite rule table (`HighResRule[]`) that maps a known thumbnail URL pattern to its high-resolution equivalent for a given site.
- Add a `resolveHighResUrl(srcUrl)` resolver that checks the rule table and returns a candidate high-res URL, or `null` if no rule matches.
- Add Twitter/X rule: media/post images at `pbs.twimg.com` with a `name` query param other than `orig` are rewritten to `name=orig`. Scoped to media images only — profile/banner image URL patterns (under `/profile_images/`, `/profile_banners/`, etc.) are not matched.
- Add Pinterest rule: images at `i.pinimg.com` with a size-segment path (`236x`, `474x`, `564x`, `736x`, etc.) are rewritten to use the `originals` segment.
- Change the background save flow to attempt the resolved high-res URL first (when a rule matches), validate the response (content-type whitelist + minimum decoded pixel dimensions), and fall back to the original `srcUrl` if validation fails or the request errors — the save never fails solely because a high-res rewrite didn't pan out.
- Combine the existing thumbnail-generation decode (`createImageBitmap`) with the new dimension-validation decode into a single decode step to avoid decoding the image twice.

## Capabilities

### New Capabilities
- `extension-highres-image-resolve`: The site-specific URL rewrite rule table and resolver (`resolveHighResUrl`) used to find a high-resolution candidate URL for a known site's thumbnail URL.

### Modified Capabilities
- `extension-save-image`: The image-fetch step of the save flow changes from "fetch `info.srcUrl` directly" to "resolve a high-res candidate if a rule matches, fetch and validate it, falling back to `info.srcUrl` on any validation failure."

## Impact

- `extensions/src/background/index.ts`: `fetchImageBlob` / `handleSave` gain a URL-resolution and validation step before the existing fetch-and-upload sequence.
- New module(s) under `extensions/src/lib/` for the rule table, resolver, and validator (exact file layout decided in design.md).
- No new permissions required — `manifest.json` already grants `host_permissions: ["<all_urls>"]`, and the fetch happens in the background service worker, which is not subject to page-level CORS.
- No backend or frontend changes.
