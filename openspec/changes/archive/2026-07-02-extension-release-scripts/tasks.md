## 1. Dependencies & Credentials

- [x] 1.1 Add `wrangler` as a dev dependency in `extensions/package.json`
- [x] 1.2 Add `CLOUDFLARE_API_TOKEN` and `R2_BUCKET_NAME` to `extensions/.env.sign.example` with comments noting the required token scope (`Workers R2 Storage:Edit`) and that this bucket is separate from the app's R2 bucket
- [x] 1.3 Add `CLOUDFLARE_API_TOKEN` and `R2_BUCKET_NAME` to `extensions/.env.sign`

## 2. Firefox Release Script

- [x] 2.1 Create `extensions/scripts/release-firefox.sh` — source `.env.sign`, call `sign-firefox.sh`, upload `.xpi` via `wrangler r2 object put` as `bookleaf-extension.xpi`, then upload `bookleaf-extension-updates.json`
- [x] 2.2 Add `"release:firefox": "bash scripts/release-firefox.sh"` to `extensions/package.json`

## 3. Chrome Release Script

- [x] 3.1 Create `extensions/scripts/release-chrome.sh` — source `.env.sign`, run `npm run build:chrome:prod`, zip `dist/chrome/` to `dist/bookleaf-extension.zip` (paths rooted at `chrome/`), upload via `wrangler r2 object put` as `bookleaf-extension.zip`
- [x] 3.2 Add `"release:chrome": "bash scripts/release-chrome.sh"` to `extensions/package.json`

## 4. Makefile

- [x] 4.1 Add `ext-release-firefox` target to `Makefile` (`cd extensions && npm run release:firefox`)
- [x] 4.2 Add `ext-release-chrome` target to `Makefile` (`cd extensions && npm run release:chrome`)

## 5. Lint

- [x] 5.1 Run `npm run build` and `npm run lint` (or type-check) in `extensions/` and fix any issues
