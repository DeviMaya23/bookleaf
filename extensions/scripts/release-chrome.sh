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

npm run build:chrome:prod

rm -f dist/bookleaf-extension.zip
(cd dist && zip -r bookleaf-extension.zip chrome/)

npx wrangler r2 object put "$R2_BUCKET_NAME/bookleaf-extension.zip" --file="dist/bookleaf-extension.zip" --remote
