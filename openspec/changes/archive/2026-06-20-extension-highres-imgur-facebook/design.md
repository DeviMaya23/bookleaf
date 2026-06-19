## Context

`extensions/src/lib/highResRules.ts` holds a `HighResRule[]` table consumed by `resolveHighResUrl()` (`highResFetch.ts`). Each rule has `id`, `matches(url)`, `transform(url)`, and an optional `referrer`. The existing Twitter and Pinterest rules are pure, synchronous string transforms — no network calls inside the rule itself. The candidate URL produced by `transform` is fetched and validated later by the existing pipeline (status `ok`, `Content-Type` in `image/jpeg|image/png|image/webp`, decoded dimensions ≥ 100×100), with automatic fallback to the original `srcUrl` on any validation failure (`extension-highres-image-resolve` / `extension-save-image` specs). Adding Imgur and Facebook rules means fitting both sites' empirically-confirmed URL behavior into this same synchronous transform shape.

Empirical findings from manual testing (see proposal):
- **Imgur**: thumbnail URLs are `i.imgur.com/{id}{sizeLetter}.webp?...` (e.g. `xCbCj7a_d.webp`). Stripping the size-letter suffix and extension yields the bare id. Re-requesting `i.imgur.com/{id}.jpg` returned byte-identical content to `.jpeg`/`.png`/`.gif` for the one sample tested (the CDN aliases these to the same master asset); only `.webp` triggers real transcoding to a smaller file. No extensionless or og:image shortcut exists — both were tested and are dead ends.
- **Facebook**: `fbcdn.net` URLs carry a signed/authenticated set of params (`_nc_ohc`, `_nc_oc`, `oh`, `oe`) plus presentation params `stp`, `cstp`, `ctp`. Stripping `ctp` alone yields a larger image without breaking the signature. Stripping `cstp` as well breaks it (mismatched/wrong result) — `cstp` is part of what the signature protects and is the real resolution ceiling reachable client-side.

## Goals / Non-Goals

**Goals:**
- Add `imgurRule` and `facebookRule` to the `rules` array in `highResRules.ts`, matching the existing `HighResRule` shape.
- Make the Imgur rule fail safe (return `null` / not match) for non-image Imgur URLs (video/gifv), so unmatched URLs fall through to the unmodified original exactly like any other unrecognized site.
- Rely on the existing content-type validation as the safety net for the unverified "does every Imgur format alias the same way" risk, rather than inventing new validation logic.

**Non-Goals:**
- No DOM inspection, no content-script changes, no new message contract fields — both rules are pure URL string transforms, same class as Twitter/Pinterest.
- No title/alt-text resolution changes — title remains `tab.title`, unrelated to this change.
- No handling of Facebook carousel/multi-photo or video posts beyond whatever the single-image-post URL shape already covers — not tested, not claimed to work, not blocking.
- No attempt to determine Imgur's true original format ahead of fetch (e.g. via a HEAD request inside `transform`) — `transform` stays synchronous per the existing contract; format correctness is left to post-fetch content-type validation, same as Twitter/Pinterest today.

## Decisions

**Imgur: pick a single fixed extension (`.jpg`) rather than probing.** `transform` is synchronous and the existing pipeline has no hook for an async pre-check before producing a candidate URL. Given the one tested sample shows `.jpg`/`.jpeg`/`.png`/`.gif` all alias to the same master bytes, requesting `.jpg` unconditionally is the simplest rule and costs nothing extra if wrong — the existing post-fetch validation (`Content-Type` allowlist) will reject it if Imgur ever serves something outside `image/jpeg|png|webp`, and the save falls back to the original thumbnail automatically. Alternative considered: keep the original extension/format (i.e. only strip the size suffix, keep `.webp`) — rejected because it forfeits the resolution gain that's the entire point of this change, and `image/webp` is already an allowed content-type so there's no extra validation risk in trying `.jpg` first.

**Imgur: exclude video explicitly via `matches`, not via a runtime check inside `transform`.** Mirrors how the Twitter rule's `matches` already excludes profile/banner images. `matches` SHALL return `false` for paths/hosts associated with Imgur video delivery (e.g. `i.imgur.com` URLs ending in `.mp4`, or `.gifv` pages), so those URLs are treated as "no rule matched" and pass through unchanged, consistent with how any other unsupported site is handled today.

**Imgur: id/suffix parsing via a single regex against the path, matching the Pinterest rule's style.** Imgur thumbnail suffixes are a single trailing letter before the extension on an otherwise fixed-length id (e.g. `aSjEe3C_d`); the rule extracts the id by matching `^/([a-zA-Z0-9]+)_[a-z]\.(webp|jpe?g|png|gif)$` (or similar) against `url.pathname` and SHALL NOT match if the path doesn't fit this exact shape (e.g. it's already a bare id with no suffix, or it's a multi-segment gallery path) — same conservative-match philosophy as the existing Pinterest size-segment regex.

**Facebook: strip `ctp` via `URLSearchParams.delete`, leave every other param untouched.** No allowlist/rewrite of other params is attempted, since `cstp` is confirmed signature-protected and any param not empirically tested is left alone rather than guessed at. Alternative considered: also try stripping `stp` — rejected, not tested, and `stp` looks like it controls format/codec hints (`dst-jpg_tt6`) rather than size, so touching it has unclear and untested risk for no confirmed benefit.

**Facebook: `matches` requires `ctp` to be present.** If a Facebook image URL has no `ctp` param at all (e.g. already at whatever its natural ceiling is), the rule returns `false` from `matches` and the original URL is used — avoids a no-op transform that would otherwise return an identical URL.

## Risks / Trade-offs

- **[Risk] Imgur's single-sample format-aliasing finding may not generalize to PNG/animated GIF sources** → **Mitigation**: existing content-type validation already rejects unexpected `Content-Type` responses and falls back to the original thumbnail automatically; worst case is a missed resolution upgrade for those formats, not a broken or corrupted save. Flagged as a known gap, not blocking.
- **[Risk] Facebook's `ctp`-strip behavior was only verified on a single-image post; carousel/video posts may render through a different URL shape or not carry `ctp` at all** → **Mitigation**: `matches` only fires when `ctp` is present, so untested post types simply don't match and fall through to the unmodified original — no new failure mode introduced for those cases, just no upgrade.
- **[Risk] Imgur CDN behavior (extension aliasing) is undocumented/unofficial and could change server-side** → **Mitigation**: same exposure Twitter's `name=orig` and Pinterest's `/originals/` rules already carry; validation + fallback contains the blast radius to "no upgrade," not a broken save.
