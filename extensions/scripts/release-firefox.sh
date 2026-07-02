#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

if [ ! -f .env.sign ]; then
  echo "Missing extensions/.env.sign (copy .env.sign.example and fill in your credentials)" >&2
  exit 1
fi

set -a
source .env.sign
set +a

bash scripts/sign-firefox.sh

XPI_PATH=$(ls -t web-ext-artifacts/*.xpi | head -n 1)
UPDATES_PATH="web-ext-artifacts/bookleaf-extension-updates.json"

npx wrangler r2 object put "$R2_BUCKET_NAME/bookleaf-extension.xpi" --file="$XPI_PATH" --content-type=application/x-xpinstall --remote
npx wrangler r2 object put "$R2_BUCKET_NAME/bookleaf-extension-updates.json" --file="$UPDATES_PATH" --remote
