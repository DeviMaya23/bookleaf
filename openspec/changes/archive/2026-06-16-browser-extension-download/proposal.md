## Why

The browser extension is not yet approved by the Chrome Web Store or listed on Firefox AMO, but users should be able to install it now. We need a way to expose self-hosted extension builds for direct download before store approval lands.

## What Changes

- Add a new public `/extensions` page using `SimplePageLayout`, with install instructions for both Firefox (signed `.xpi` via AMO unlisted channel) and Chrome (`.zip`, load-unpacked via developer mode)
- Add an **Extensions** section to the Settings modal alongside Account / App / Advanced, with the same download links and instructions
- Add `src/lib/downloads.ts` to hold the two R2-hosted download URL constants used by both surfaces
- Add an **Extensions** footer link on the landing page next to About / Privacy / AI Notes

## Capabilities

### New Capabilities

- `fe-extension-downloads`: A `src/lib/downloads.ts` module exporting stable R2-hosted download URL constants (`EXTENSION_FIREFOX_URL`, `EXTENSION_CHROME_URL`) shared across the public page and settings section

### Modified Capabilities

- `fe-content-pages`: Add `/extensions` route — a new `SimplePageLayout`-based page with Firefox and Chrome install instructions and download links
- `fe-settings-modal`: Add an **Extensions** section to the modal's left-nav and content area, following the same section component pattern as App / Account / Advanced

## Impact

- `src/lib/downloads.ts` — new file (new pattern: shared URL constants module)
- `src/pages/ExtensionsPage.tsx` — new file
- `src/features/settings/components/ExtensionsSection.tsx` — new file
- `src/features/settings/components/SettingsModal.tsx` — add `extensions` to `SECTIONS`, render `ExtensionsSection`
- `src/pages/LandingPage.tsx` — add Extensions link to footer
- `src/App.tsx` — add `/extensions` route
- No backend changes. No new npm dependencies.
- Requires a new Cloudflare R2 bucket (`bookleaf-public`) with public access and a custom domain (e.g. `downloads.bookleaf.app`) — this is infrastructure setup done outside the codebase.
