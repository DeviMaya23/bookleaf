#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

GECKO_ID="bookleaf@evimay.me"
UPDATE_LINK="https://bookleaf-files.bookverse.fyi/bookleaf-extension.xpi"
ARTIFACTS_DIR="web-ext-artifacts"

VERSION=$(node -p "require('./manifest.json').version")

XPI_PATH=$(ls -t "$ARTIFACTS_DIR"/*.xpi | head -n 1)

if command -v sha256sum >/dev/null 2>&1; then
  HASH=$(sha256sum "$XPI_PATH" | cut -d' ' -f1)
else
  HASH=$(shasum -a 256 "$XPI_PATH" | cut -d' ' -f1)
fi

cat > "$ARTIFACTS_DIR/bookleaf-extension-updates.json" <<EOF
{
  "addons": {
    "$GECKO_ID": {
      "updates": [
        {
          "version": "$VERSION",
          "update_link": "$UPDATE_LINK",
          "update_hash": "sha256:$HASH"
        }
      ]
    }
  }
}
EOF

echo "Wrote $ARTIFACTS_DIR/bookleaf-extension-updates.json (version $VERSION, hash sha256:$HASH)"
