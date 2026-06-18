## 1. Rule table and resolver

- [x] 1.1 Add `extensions/src/lib/highResRules.ts`: define `HighResRule` type, the Twitter/X rule (`/media/` path + `format`/`name` query params → `name=orig`), the Pinterest rule (size-segment path → `originals`), and the exported `rules` array. **(corrected post-implementation — the live Twitter URL shape is query-param-based, not the `:large`/`:small` path-suffix scheme originally assumed; see design.md note)**
- [x] 1.2 Add `resolveHighResUrl(srcUrl: string): string | null` in `extensions/src/lib/highResFetch.ts`, iterating `rules` and returning the first match's transform result.
- [ ] 1.3 Unit tests for `resolveHighResUrl` and each rule's `matches`/`transform`: Twitter `name=small` → `name=orig`, Twitter numeric `name` (e.g. `360x360`) → `name=orig`, Twitter profile/banner URL → no match, Pinterest sized URL → `originals`, unrecognized site → `null`. **(skipped — no test runner configured in `extensions/`; user decided to write code only)**

## 2. Candidate validation

- [x] 2.1 Add a `validateCandidate(response, blob)`-style helper in `extensions/src/lib/highResFetch.ts`: checks `response.ok`, content-type whitelist (`image/jpeg`, `image/png`, `image/webp`), and (when `createImageBitmap` is available) decoded dimensions ≥ 100×100.
- [x] 2.2 Ensure the validator reuses a single `createImageBitmap` decode rather than decoding twice when the same blob later needs thumbnail generation.
- [ ] 2.3 Unit tests for the validator: valid response passes; non-OK response fails; disallowed content-type fails; undersized decoded dimensions fail; `createImageBitmap` unavailable skips dimension check and validates on content-type alone. **(skipped — same reason as 1.3)**

## 3. Wire into the save flow

- [x] 3.1 In `extensions/src/background/index.ts`, update the fetch step of `handleSave`/`fetchImageBlob` to: call `resolveHighResUrl(srcUrl)`; if a candidate is returned, fetch and validate it; on success use its blob, on any failure (no match, fetch error, validation failure) fall back to fetching `srcUrl` directly exactly as today.
- [x] 3.2 Thread the decoded bitmap from a successful high-res validation into the existing thumbnail-generation step so it isn't decoded a second time.
- [x] 3.3 Confirm the existing error-toast path ("Couldn't save image.") only fires when the fallback fetch of `srcUrl` itself fails — not merely because the high-res candidate was invalid.

## 4. Verification

- [x] 4.1 Manually verify saving a Twitter/X post image upgrades to `name=orig` resolution. **(verified by user — both the post and media tab save high-res)**
- [x] 4.2 Manually verify saving a Twitter/X profile/banner image is unaffected (no rule match, original behavior). **(verified by user — header photo and profile picture both save low-res, as designed; profile picture context-menu detection is flaky but that's a separate, out-of-scope issue)**
- [ ] 4.3 Manually verify saving a Pinterest pin image upgrades to the `originals` resolution. **(blocked — extension doesn't currently detect Pinterest pin image cards as a right-click target; tracked as a follow-up proposal, not part of this change)**
- [x] 4.4 Manually verify a save still succeeds (using the original thumbnail) when a high-res candidate is deliberately broken (e.g. point a rule at a 404 or a tiny placeholder) to confirm fallback behavior. **(verified by user — temporarily broke the Twitter rule's `name` param, save succeeded with original low-res image and no error toast; change reverted)**
- [x] 4.5 Run `npm run build` in `extensions/` and fix any issues. (build + `tsc --noEmit` both pass)
- [ ] 4.6 Run `npm run lint` in `extensions/` and fix any issues. **(skipped — no lint script/config configured in `extensions/`)**
