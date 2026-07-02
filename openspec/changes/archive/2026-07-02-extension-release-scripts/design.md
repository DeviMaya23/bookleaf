## Context

Firefox extension releases currently require two manual steps after `npm run sign:firefox` completes: uploading the `.xpi` to R2, then uploading `bookleaf-extension-updates.json`. The order matters — an installed extension polling the update URL between the two uploads would see a hash pointing to a file that doesn't exist yet. Chrome has no scripted path at all: build, zip, and upload are all done by hand.

The R2 bucket used for extension distribution is separate from the app's storage bucket. Credentials for it belong alongside the AMO signing credentials in `extensions/.env.sign`, which is already the pattern for release-time secrets.

## Goals / Non-Goals

**Goals:**
- Automate Firefox release: sign + upload `.xpi` then `updates.json` in guaranteed order
- Automate Chrome release: build + zip + upload
- Keep both release commands as simple, single-step invocations (`npm run release:firefox`, `npm run release:chrome`)
- No changes to existing signing or manifest-generation scripts

**Non-Goals:**
- Version pinning or artifact archiving in R2 (latest-only, overwrite on each release)
- CI/CD integration (scripts are run locally by the developer)
- A combined "release both" command

## Decisions

### Use `wrangler r2 object put` for R2 uploads

`wrangler` is Cloudflare's own CLI and is already present in the project (used by the frontend). It uploads via the Cloudflare API using a `CLOUDFLARE_API_TOKEN`, which is a single credential rather than the three required by the S3-compatible API (`R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`).

Alternatives considered:
- **`@aws-sdk/client-s3`** — requires a JS wrapper script and three env vars; more portable but more moving parts for a dev-only release tool
- **`aws-cli`** — not installed on the development machine; would add an external system requirement
- **`rclone`** — not installed; same issue

### Add `wrangler` as a dev dependency in `extensions/`

Rather than relying on the frontend's `wrangler` installation or a globally installed version, it is added as a dev dep to `extensions/package.json`. This makes `npx wrangler` reliable without assumptions about the developer's global environment.

### `release-firefox.sh` calls `sign-firefox.sh` as a subprocess

`sign-firefox.sh` is the single source of truth for building, signing, and generating the update manifest. `release-firefox.sh` shells out to it rather than duplicating any of that logic. Both scripts source `.env.sign` independently — double-sourcing the same file is harmless.

### Chrome zip is created inside `dist/` from within that directory

The zip command runs as `(cd dist && zip -r bookleaf-extension.zip chrome/)` so that archive paths are `chrome/<file>`, not `dist/chrome/<file>`. This matches the existing `dist/bookleaf-extension.zip` artifact that already appears there from prior manual releases.

### R2 credentials live in `extensions/.env.sign`

`.env.sign` is already the designated file for extension release-time secrets (AMO credentials). Adding R2 credentials there keeps all release prerequisites in one place. The variable names are `CLOUDFLARE_API_TOKEN` and `R2_BUCKET_NAME` — distinct from the app's `R2_*` variables in the root `.env` to avoid confusion between the two buckets.

## Risks / Trade-offs

- **wrangler auth scope**: The `CLOUDFLARE_API_TOKEN` must have `Workers R2 Storage:Edit` permission on the extension bucket specifically. Using an overly broad token would be a security risk. → Mitigation: document required token scope in `.env.sign.example`.
- **No upload confirmation beyond exit code**: `wrangler r2 object put` exits non-zero on failure; `set -euo pipefail` in the scripts ensures the process halts. There is no post-upload integrity check. → Acceptable for a developer-run local tool; the developer can verify via the Cloudflare dashboard if needed.
- **Firefox upload order enforced by script, not atomic**: The gap between `.xpi` upload completing and `updates.json` upload starting is small but non-zero. An extension polling in that window would get a 404 on the `.xpi`. → This is the same risk that exists today with manual uploads and is documented in `sign-firefox.sh`; it is not made worse by automation.
