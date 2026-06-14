#!/usr/bin/env bash
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
