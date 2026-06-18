## Context

The extension's only save path is `handleSave()` in `extensions/src/background/index.ts`, triggered by the "Save to Bookleaf" image context menu. It currently fetches `info.srcUrl` as-is via `fetchImageBlob()`. Twitter/X and Pinterest serve downscaled thumbnail URLs by default for the `<img>` elements on their pages; both sites expose their full-resolution originals at a different, predictable URL reachable via a simple string transform of the thumbnail URL. This design introduces a rule-table abstraction (new pattern, not previously present in the codebase) to resolve and validate a high-res candidate before falling back to today's behavior.

`manifest.json` already grants `host_permissions: ["<all_urls>"]`, and the fetch runs in the background service worker, so no new permissions or CORS handling are needed.

## Goals / Non-Goals

**Goals:**
- Automatically resolve a high-resolution candidate URL for Twitter/X and Pinterest media images at save time, with no user action required.
- Validate the candidate response heuristically (content-type + decoded pixel dimensions) and fall back to the original `srcUrl` if validation fails, so a broken or stale rule never causes the save itself to fail.
- Structure the rule table so a third site can be added later by adding one rule object, without touching the resolver, validator, or existing rules.

**Non-Goals:**
- A generic "find the best resolution" algorithm — rules are explicit, reverse-engineered, per-site string transforms.
- Dimension comparison against the original thumbnail's actual rendered size (would require a `contextmenu` DOM listener in the content script and a content-script↔background handoff; deferred to a future change if the heuristic-only approach proves insufficient).
- Async/network-based resolution rules (e.g. resolving via a lookup request or following a redirect) — out of scope until a site actually needs it (YAGNI). The rule `transform` signature stays synchronous.
- Twitter profile/banner images (`_normal`, `_bigger`, `_400x400`, etc.) — scoped to media/post images only.
- Sites beyond Twitter/X and Pinterest.
- Monitoring/alerting for when a site silently changes its CDN scheme and a rule stops working — fallback masks breakage by design (see Risks).

## Decisions

### Rule table shape and location
```ts
type HighResRule = {
  id: string;
  matches: (url: URL) => boolean;
  transform: (url: URL) => string | null;
};
```
- New module `extensions/src/lib/highResRules.ts`: defines `HighResRule`, the Twitter and Pinterest rule objects, and the exported `rules: HighResRule[]` array.
- New module `extensions/src/lib/highResFetch.ts`: exports `resolveHighResUrl(srcUrl): string | null` (iterates `rules`, returns the first match's `transform` result) and the fetch+validate+fallback orchestration consumed by `background/index.ts`.
- Kept synchronous per earlier decision — no site identified so far needs a network call to determine the high-res URL, and adding `Promise` to the signature later is not a breaking change to callers that already `await` it isn't needed; we accept revisiting the signature only if/when an async site appears.

### Twitter/X rule
- Matches `pbs.twimg.com` URLs with a `/media/` path segment (post images), a `format` query param of `jpg`, `png`, or `webp`, and a `name` query param present and not already `orig`.
- Excludes profile/banner URLs, which live under `/profile_images/` or `/profile_banners/` (not `/media/`) and won't match the path check.
- The `format` query param check is a second scoping signal alongside the path check, to avoid matching unrelated `pbs.twimg.com` paths that aren't post images.
- Transform: set the `name` query param to `orig`, leaving `format` and the rest of the URL unchanged.
- Note: an earlier version of this rule assumed a legacy path-suffix URL shape (`/media/XXX.jpg:large` → `/media/XXX.jpg:orig`). Twitter/X currently serves `/media/XXX?format=jpg&name=small` instead, with the size encoded in the `name` query param (`thumb`, `small`, `medium`, `large`, `360x360`, `orig`, etc.) rather than a path suffix. The rule was corrected to match the live query-param format.

### Pinterest rule
- Matches `i.pinimg.com` URLs with a size-segment path component (`236x`, `474x`, `564x`, `736x`, etc.).
- Transform: replace the size segment with `originals`.

### Validation (heuristic-only)
- Triggered only when `resolveHighResUrl` returns a non-null candidate (i.e., a rule matched).
- Checks, in order:
  1. Response `ok`.
  2. `Content-Type` is one of `image/jpeg`, `image/png`, `image/webp`.
  3. Decoded pixel dimensions (via `createImageBitmap`) are at least 100×100.
- No byte-size floor — decoded pixel dimensions are a more reliable signal than byte size, which varies with compression/format.
- If `createImageBitmap` is unavailable in the runtime, dimension validation is skipped and only the content-type check applies (mirrors the existing `typeof OffscreenCanvas !== "undefined"` guard already used for thumbnail generation — `createImageBitmap` and `OffscreenCanvas` availability are treated as coupled in this codebase).
- On any validation failure (or a fetch error/non-ok response on the candidate URL), fall back to fetching `srcUrl` directly and proceed exactly as the save flow does today. The fallback fetch is not itself re-validated — it's today's existing trusted path.
- Risk asymmetry: over-triggering fallback only loses the high-res upgrade (graceful degrade to current behavior); under-triggering risks silently saving the wrong content. Thresholds and checks are intentionally conservative, biased toward triggering fallback when uncertain.

### Decode-once
- The existing thumbnail flow already decodes the saved blob via `createImageBitmap` when generating the thumbnail. The new validation step reuses that same decode (bitmap + dimensions) rather than decoding twice — `fetchImageBlob` → decode → validate → (if valid) generate thumbnail from the already-decoded bitmap.

## Risks / Trade-offs

- **[Risk] Site changes its CDN/URL scheme, silently breaking a rule.** → Mitigation: validation + fallback means the save still succeeds with the original thumbnail; no save-breaking failure mode. Detecting *that* a rule has gone stale is out of scope for this change (no telemetry/alerting added).
- **[Risk] Heuristic validation has false positives/negatives.** → Mitigation: conservative thresholds (content-type whitelist + 100×100 floor), and the cost of a false positive (lose the upgrade) is much lower than a false negative (wrong image saved), so thresholds lean toward triggering fallback.
- **[Risk] Extra fetch + decode on every save where a rule matches, even when the thumbnail would have been fine.** → Mitigation: only sites with confirmed thumbnail-by-default behavior get a rule; decode is shared with thumbnail generation so it's not pure overhead; no rule match (most other sites) means zero added cost.
- **[Trade-off] No dimension comparison against the actual thumbnail's real size.** → Accepted for this change since it requires content-script DOM access not currently wired up; the heuristic floor is a coarser but zero-new-surface-area substitute.

## Migration Plan

No data migration. This is a self-contained extension code change — ships via the existing build/release pipeline (`extensions/dist/chrome`, `extensions/dist/firefox`). Rollback is a plain revert; no persisted state or schema is affected.

## Open Questions

None outstanding — minimum pixel floor confirmed at 100×100, and Twitter scoping confirmed to use both the `/media/` path check and an image-format/extension check.
