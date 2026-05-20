## Why

Users currently must open the Bookleaf web app to save images, which breaks their browsing workflow. A browser extension lets users save images directly from any webpage without context-switching, and this proposal establishes the foundational scaffold and authentication needed to make that possible.

## What Changes

- New `/extensions` directory containing a Vite + TypeScript project targeting Manifest V3 (MV3)
- Extension supports both Chrome and Firefox via `webextension-polyfill`
- Login UI (popup) that authenticates with the existing Bookleaf backend
- Auth token persisted in `chrome.storage.local` so users stay logged in across sessions
- Background service worker wired up for future image-saving functionality

## Capabilities

### New Capabilities

- `extension-scaffold`: Project structure, Vite build config, MV3 manifest, webextension-polyfill setup, and TypeScript configuration for the browser extension under `/extensions`
- `extension-auth`: Login flow in the extension popup — Kinde OAuth via `chrome.identity.launchWebAuthFlow`, exchanged token stored in `chrome.storage.local`, and session state reflected in the popup UI

### Modified Capabilities

## Impact

- New top-level `/extensions` directory (no changes to existing backend or frontend)
- Depends on existing Kinde-based auth endpoint (`/api/v1/auth/login` or equivalent) for token issuance
- No database or API schema changes required for this proposal
