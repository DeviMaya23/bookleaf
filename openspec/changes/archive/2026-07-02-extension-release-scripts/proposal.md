## Why

The Firefox and Chrome extension release processes are partially scripted but still require manual R2 uploads after each build — the Firefox signing step already produces the correct artifacts, but uploading them (and building/zipping Chrome) is done by hand each time. This creates room for sequencing errors (e.g. uploading `updates.json` before the `.xpi` is live) and makes releases slower than they need to be.

## What Changes

- Add `extensions/scripts/release-firefox.sh` — calls the existing `sign-firefox.sh`, then uploads the signed `.xpi` to R2 first, followed by `bookleaf-extension-updates.json`
- Add `extensions/scripts/release-chrome.sh` — runs the Chrome production build, zips `dist/chrome/`, and uploads the zip to R2
- Add `npm run release:firefox` and `npm run release:chrome` scripts to `extensions/package.json`
- Add `ext-release-firefox` and `ext-release-chrome` Makefile targets
- Add R2 credentials (`CLOUDFLARE_API_TOKEN`, `R2_BUCKET_NAME`) to `extensions/.env.sign` and `.env.sign.example`
- Add `wrangler` as a dev dependency in `extensions/` for R2 uploads via `wrangler r2 object put`

## Capabilities

### New Capabilities

- `extension-release-scripts`: Automated release scripts for Firefox and Chrome extensions — building, packaging, and uploading artifacts to R2 in the correct order using wrangler

### Modified Capabilities

None.

## Impact

- `extensions/scripts/` — two new shell scripts
- `extensions/package.json` — two new npm scripts, one new dev dependency (`wrangler`)
- `extensions/.env.sign` and `.env.sign.example` — two new env vars
- `Makefile` — two new targets
- No changes to existing scripts (`sign-firefox.sh`, `generate-update-manifest.sh`)
- No backend, frontend, or API changes
