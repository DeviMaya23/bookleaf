#!/usr/bin/env bash
# Firefox release checklist:
#   1. Bump "version" in extensions/manifest.json.
#   2. Run this script (npm run sign:firefox) — builds, signs, and writes
#      web-ext-artifacts/updates.json.
#   3. Upload to R2: the .xpi FIRST, then updates.json (so a mid-upload poll
#      from an installed extension never sees a hash that doesn't match what's
#      live yet at the .xpi URL).
#
# Note: update_url (and therefore auto-update) is only present in the Firefox
# manifest from this release onward. Anyone with an older build installed
# still needs one manual reinstall to pick up update_url; every release after
# that auto-updates on its own.
set -euo pipefail
cd "$(dirname "$0")/.."

if [ ! -f .env.sign ]; then
  echo "Missing extensions/.env.sign (copy .env.sign.example and fill in your AMO API credentials)" >&2
  exit 1
fi

set -a
source .env.sign
set +a

npm run build:firefox:prod

npx web-ext sign \
  --source-dir=dist/firefox \
  --artifacts-dir=web-ext-artifacts \
  --api-key="$AMO_JWT_ISSUER" \
  --api-secret="$AMO_JWT_SECRET" \
  --channel=unlisted

bash scripts/generate-update-manifest.sh
