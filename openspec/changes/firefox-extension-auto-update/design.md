## Context

The Firefox extension is signed via AMO's `unlisted` channel (`scripts/sign-firefox.sh`) and distributed from a fixed-filename `.xpi` on R2 (`EXTENSION_FIREFOX_URL` in `frontend/src/lib/downloads.ts`), linked from the extensions download page. There is no scripted upload to R2 — moving the signed `.xpi` (and, after this change, `updates.json`) onto R2 stays a manual step.

Firefox's self-hosted auto-update model is passive: the *already-installed* extension's manifest carries `browser_specific_settings.gecko.update_url`, and Firefox periodically fetches whatever JSON currently lives at that URL, comparing its own version against the `version` entry there. Nothing in this design can make Firefox check a URL that isn't already baked into a previously-installed build — see Migration Plan.

The Firefox/Chrome manifest split already exists in `vite.config.ts`'s `transformManifest`, gated on `isFirefox` / `isProduction`. This change extends that same gate; Chrome is untouched both now and when Chrome Web Store submission happens later (Web Store installs auto-update via Google's own infrastructure, not via manifest `update_url`).

## Goals / Non-Goals

**Goals:**
- Firefox production builds carry an `update_url` pointing at a hosted `updates.json`.
- A script reliably produces `updates.json` (Mozilla update-manifest format) matching the `.xpi` that was just signed — same version, correct sha256 hash — so the two artifacts can't drift apart.
- The release flow stays a single command (`npm run sign:firefox`) plus a manual upload, with no new persistent state to maintain.

**Non-Goals:**
- Automating the R2 upload itself (stays manual, per the existing flow).
- Multi-version `updates.json` history, `strict_min_version`/`strict_max_version` applicability rules, or rollback support — out of scope until there's a reason for them.
- Any Chrome-side changes — Chrome relies on Web Store auto-update once submitted, which is a separate, unrelated mechanism.
- Automating the `manifest.json` version bump — still a manual edit before running the release script.

## Decisions

**Single-entry `updates.json`, overwritten each release.** Mozilla's update-manifest format supports an array of historical version entries, but nothing in this project needs old versions to remain installable or for Firefox to choose between candidates. Each release simply overwrites the file with one `updates` entry for the current version. Simpler, and the only state involved is "what's the most recent release," which is already the version in `manifest.json`.

**Update-manifest generation lives inside `scripts/sign-firefox.sh`, not a separate npm script.** Folding hash + `updates.json` generation into the same script run that just produced the `.xpi` means the hash can never be computed against a stale or wrong artifact — there's no window where someone runs signing and update-manifest generation as two separate commands against two different builds.

**`update_link`/`update_url` host values are hardcoded constants, not env vars.** They point at the same fixed R2 host already hardcoded as `EXTENSION_FIREFOX_URL` in the frontend. Extensions and frontend are separate packages with no shared config layer today, so introducing one (an env var, a shared constants package) for two URL strings would be a new abstraction the proposal doesn't need — out of scope per the project's decision-boundary rule. The values are duplicated; if they need to change, both call sites are simple greps away (`bookleaf-files.evimay.me`).

**Hashing uses `shasum -a 256` with a `sha256sum` fallback.** `sha256sum` isn't installed in this macOS dev environment; `shasum -a 256` is. The script checks for `sha256sum` first (likely present on Linux CI runners, if this ever moves there) and falls back to `shasum -a 256` (macOS default), rather than assuming one or the other.

**The script locates the signed `.xpi` by freshest mtime in `web-ext-artifacts/` immediately after `web-ext sign` runs**, rather than trying to predict `web-ext`'s output filename (which embeds the extension name and version and isn't worth hardcoding/parsing).

## Risks / Trade-offs

- **[Risk]** Forgetting to bump `version` in `manifest.json` before a release → the script faithfully reproduces a same-version `updates.json`, and Firefox sees no update at all (silently, not an error). → **Mitigation**: none automated in this change (would require persisting "last published version" somewhere, a new piece of state not justified yet); flagged as an Open Question below.
- **[Risk]** Manual upload of `.xpi` and `updates.json` to R2 isn't atomic → a user could poll mid-upload and fetch a new `updates.json` pointing at a hash that doesn't match the `.xpi` still live at the old content. → **Mitigation**: document upload order in the release checklist — upload the `.xpi` first, `updates.json` second; worst case is a failed update attempt that retries on Firefox's next poll, not a broken install.
- **[Trade-off]** No applicability constraints (`strict_min_version`, etc.) in `updates.json` means there's no way to gate a release to only some Firefox versions if a future release needs that. Acceptable now; would need a follow-up if it becomes relevant.

## Migration Plan

`update_url` is read from the manifest of the *currently installed* extension — Firefox does not retroactively discover it. Concretely:

1. Anyone with a Firefox build installed today (no `update_url`) will **not** auto-update to the first release built under this change. They still need one manual reinstall.
2. From that reinstall onward, every subsequent release auto-updates, since the installed manifest now carries `update_url`.
3. No rollback mechanism is needed — this only adds a manifest field and a generated artifact; it doesn't change existing save/auth/runtime behavior, so the rollback story is "stop pointing `update_url` at a server" / "don't include the field," not a data migration.

## Open Questions

- Should a guard against re-publishing an unchanged `version` be added later (failing the script if `manifest.json`'s version matches what's already in the last-uploaded `updates.json`)? Deferred — would need to fetch or cache the previously published `updates.json`, which is new state/behavior beyond this proposal's scope.
