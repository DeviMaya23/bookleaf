## Why

The browser extension popup is a bare skeleton — plain unstyled buttons with no branding, no user context, and no feedback on what the extension has done. Polishing it makes the extension feel like a real product and surfaces the "recently saved" history so users can see their saves without opening the full app.

## What Changes

- Redesign the popup (320px wide) with a proper header, user row, and dark mode toggle
- Logged-out state: branded layout with tagline and full-width CTA
- Logged-in state: avatar, username, "Open ↗" link, recently saved thumbnail strip (up to 5), "Log out" footer
- Empty state: friendly hint when no images have been saved from the extension yet
- Dark mode preference persisted in `chrome.storage.local`, toggled from the user row
- Store username at login time by decoding the Kinde JWT payload (no extra network call)
- After each successful save, generate a 60×60 JPEG thumbnail via `OffscreenCanvas` + `createImageBitmap` in the background service worker and store it locally (max 5, FIFO)
- Popup reads `recentSaves` from `chrome.storage.local` on mount — no backend calls needed

## Capabilities

### New Capabilities

- `extension-popup-ui`: Visual design and state machine for the extension popup — logged-out, logged-in, empty state, and dark mode
- `extension-recent-saves`: Background thumbnail generation and local storage of the five most recently saved images, read by the popup on open

### Modified Capabilities

_(none — no existing spec requirements are changing)_

## Impact

- `extensions/src/popup/App.tsx` — full redesign
- `extensions/src/popup/index.html` — width 280px → 320px
- `extensions/src/lib/storage.ts` — new helpers: `getRecentSaves`, `addRecentSave`, `getDarkMode`, `setDarkMode`; username stored alongside auth
- `extensions/src/lib/auth.ts` — decode JWT payload to extract and store `given_name` (or `name`) at login
- `extensions/src/background/index.ts` — after successful save, generate thumbnail and call `addRecentSave`
- No backend changes; no new API calls from the popup
