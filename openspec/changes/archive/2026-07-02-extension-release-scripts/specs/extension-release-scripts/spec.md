## ADDED Requirements

### Requirement: Firefox release script uploads artifacts to R2 in the correct order

`npm run release:firefox` SHALL execute the full Firefox release in a single command: it calls `scripts/sign-firefox.sh` (which builds, signs, and generates the update manifest), then uploads the signed `.xpi` to R2 as `bookleaf-extension.xpi`, and only after that upload completes uploads `bookleaf-extension-updates.json`. The script SHALL abort on any failure (`set -euo pipefail`). It SHALL source `extensions/.env.sign` for both AMO and R2 credentials.

#### Scenario: Firefox release succeeds end-to-end

- **WHEN** `npm run release:firefox` is run with valid `.env.sign` credentials
- **THEN** `sign-firefox.sh` is called and completes (build + sign + manifest)
- **AND** the signed `.xpi` is uploaded to R2 as `bookleaf-extension.xpi`
- **AND** `bookleaf-extension-updates.json` is uploaded to R2 only after the `.xpi` upload exits successfully

#### Scenario: Firefox release aborts if `.env.sign` is missing

- **WHEN** `npm run release:firefox` is run and `extensions/.env.sign` does not exist
- **THEN** the script exits with a non-zero status before attempting any build or upload
- **AND** an error message indicating the missing file is printed to stderr

#### Scenario: Firefox release aborts if signing fails

- **WHEN** `sign-firefox.sh` exits with a non-zero status (e.g. invalid AMO credentials)
- **THEN** `release-firefox.sh` exits immediately
- **AND** no R2 upload is attempted

#### Scenario: Firefox release aborts if the XPI upload fails

- **WHEN** the `wrangler r2 object put` call for the `.xpi` exits with a non-zero status
- **THEN** `release-firefox.sh` exits immediately
- **AND** `bookleaf-extension-updates.json` is NOT uploaded to R2

### Requirement: Chrome release script builds, zips, and uploads to R2

`npm run release:chrome` SHALL execute the full Chrome release in a single command: it runs `npm run build:chrome:prod`, zips the `dist/chrome/` directory to `dist/bookleaf-extension.zip` (paths inside the archive rooted at `chrome/`), and uploads the zip to R2 as `bookleaf-extension.zip`. The script SHALL abort on any failure. It SHALL source `extensions/.env.sign` for R2 credentials.

#### Scenario: Chrome release succeeds end-to-end

- **WHEN** `npm run release:chrome` is run with valid `.env.sign` credentials
- **THEN** `dist/chrome/` is populated by the production build
- **AND** `dist/bookleaf-extension.zip` is created with archive paths rooted at `chrome/`
- **AND** the zip is uploaded to R2 as `bookleaf-extension.zip`

#### Scenario: Chrome release aborts if `.env.sign` is missing

- **WHEN** `npm run release:chrome` is run and `extensions/.env.sign` does not exist
- **THEN** the script exits with a non-zero status before attempting any build or upload
- **AND** an error message indicating the missing file is printed to stderr

#### Scenario: Chrome release aborts if the build fails

- **WHEN** `npm run build:chrome:prod` exits with a non-zero status
- **THEN** `release-chrome.sh` exits immediately
- **AND** no zip is created and no R2 upload is attempted

### Requirement: R2 credentials and token scope are documented in `.env.sign.example`

`extensions/.env.sign.example` SHALL include `CLOUDFLARE_API_TOKEN` and `R2_BUCKET_NAME` with inline comments noting that the token requires `Workers R2 Storage:Edit` permission scoped to the extension distribution bucket, and that this bucket is separate from the app's R2 bucket.

#### Scenario: Developer sets up release credentials from the example file

- **WHEN** a developer copies `.env.sign.example` to `.env.sign` and fills in the values
- **THEN** both `npm run release:firefox` and `npm run release:chrome` have all credentials needed to complete a release
