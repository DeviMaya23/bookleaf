## 1. Manifest transform

- [x] 1.1 In `extensions/vite.config.ts`, add a hardcoded `FIREFOX_UPDATE_URL` constant for the hosted `updates.json` location, alongside the existing `geckoId` constant.
- [x] 1.2 In `transformManifest`, inject `browser_specific_settings.gecko.update_url` only when `mode === "firefox-production"` (not for `firefox` dev mode, not for chrome modes).
- [x] 1.3 Run `npm run build:firefox:prod` and confirm `dist/firefox/manifest.json` contains `browser_specific_settings.gecko.update_url` set to the constant value.
- [x] 1.4 Run `npm run build:firefox` and confirm `dist/firefox/manifest.json` does NOT contain `update_url`.
- [x] 1.5 Run `npm run build` (chrome) and confirm the output manifest still has no `browser_specific_settings` at all.

## 2. Update manifest generation script

- [x] 2.1 Add a script (e.g. `extensions/scripts/generate-update-manifest.sh`) that: reads `version` from `extensions/manifest.json`, locates the freshest `.xpi` in `web-ext-artifacts/`, computes its sha256 (prefer `sha256sum`, fall back to `shasum -a 256`), and writes `web-ext-artifacts/updates.json` in Mozilla's update-manifest format for gecko ID `bookleaf@evimay.me`, with `update_link` set to the existing fixed Firefox `.xpi` URL.
- [x] 2.2 Ensure the script always overwrites `updates.json` with a single `updates` entry (no appending/history).
- [x] 2.3 Call this script from the end of `extensions/scripts/sign-firefox.sh`, after the `web-ext sign` step succeeds.
- [x] 2.4 Manually verify: run `npm run sign:firefox` against test AMO credentials (or a local dry run of the new script against a manually-placed test `.xpi`) and confirm `web-ext-artifacts/updates.json` has the correct `version`, `update_link`, and a sha256 `update_hash` matching the actual `.xpi` file via `shasum -a 256`.
- [x] 2.5 Bump `extensions/manifest.json` version, re-run, and confirm the regenerated `updates.json` reflects only the new version (old entry is gone, not appended).

## 3. Documentation

- [x] 3.1 Update the release checklist (README or wherever the Firefox release process is documented) to state: bump version → `npm run sign:firefox` → upload both the `.xpi` and `web-ext-artifacts/updates.json` to R2, `.xpi` first then `updates.json`.
- [x] 3.2 Note in the same documentation that the first release under this change still requires existing installs to be manually reinstalled once; only releases after that auto-update.

## 4. Final checks

- [x] 4.1 Run `npm run build` and `npm run lint` in `extensions/` and fix any issues. (No `lint` script exists in `extensions/package.json`; ran `npm run build` and `npm run type-check` instead — both pass.)
