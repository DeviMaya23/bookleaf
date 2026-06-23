## Why

The Firefox extension is self-distributed (signed via AMO's `unlisted` channel, downloaded manually from the existing download page) and has no auto-update mechanism — every release requires users to manually re-download and reinstall the `.xpi`. Before the user base grows further, the extension should auto-update the same way Chrome Web Store installs will once that submission lands.

## What Changes

- Inject `browser_specific_settings.gecko.update_url` into the Firefox **production** manifest only (via the existing `transformManifest` step in `vite.config.ts`, alongside the existing `gecko.id` injection), pointing at a hosted `updates.json` update manifest. Dev/sideloaded Firefox builds SHALL NOT receive this field, to avoid dev installs attempting to "update" to a production release.
- Add a script that, given a signed `.xpi`, generates an `updates.json` file in Mozilla's update-manifest format containing the current `version` (read from `manifest.json`), the `update_link` (the existing fixed `.xpi` URL), and an `update_hash` (sha256 of the signed `.xpi`).
- Wire this generation step into `scripts/sign-firefox.sh` (or document it as the next step after `npm run sign:firefox`) so a release produces both the signed `.xpi` and a matching `updates.json` ready to upload.
- Document the release checklist: bump `version` in `manifest.json` → `npm run sign:firefox` → upload the resulting `.xpi` and `updates.json` to R2 (upload itself stays manual; no R2 automation in scope).

## Capabilities

### New Capabilities
- `extension-firefox-update-manifest`: generation of a Mozilla-format `updates.json` (version, update_link, update_hash) from a signed Firefox `.xpi`, used to drive self-hosted auto-update checks.

### Modified Capabilities
- `extension-firefox-compat`: the Firefox manifest transform gains a production-only `update_url` injection alongside the existing `gecko.id` injection.

## Impact

- `extensions/vite.config.ts` — `transformManifest` firefox branch gains `update_url`, gated on `isProduction`.
- `extensions/scripts/sign-firefox.sh` — gains a step (or a new sibling script) to compute the `.xpi` hash and emit `updates.json`.
- `extensions/package.json` — possible new script entry for generating the update manifest.
- No backend, frontend, or R2 upload automation changes — uploading the `.xpi`/`updates.json` to `bookleaf-files.evimay.me` remains a manual step outside this repo.
