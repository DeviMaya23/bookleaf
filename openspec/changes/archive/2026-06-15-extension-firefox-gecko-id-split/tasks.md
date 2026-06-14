## 1. Implementation

- [x] 1.1 In `extensions/vite.config.ts`, update the Firefox `transformManifest` to set `browser_specific_settings.gecko.id` based on `mode`: `"bookleaf-dev@evimay.me"` for `firefox`, `"bookleaf@evimay.me"` for `firefox-production`.

## 2. Verification

- [x] 2.1 Run `npm run build:firefox` and confirm `dist/firefox/manifest.json` has `browser_specific_settings.gecko.id` set to `"bookleaf-dev@evimay.me"`.
- [x] 2.2 Run `npm run build:firefox:prod` and confirm `dist/firefox/manifest.json` has `browser_specific_settings.gecko.id` set to `"bookleaf@evimay.me"`.
- [x] 2.3 Run `npm run build` (Chrome) and confirm `dist/chrome/manifest.json` still has no `browser_specific_settings`.
- [x] 2.4 Run `npm run type-check` and fix any issues.

## 3. Dev Profile Migration (manual)

- [x] 3.1 Reload the rebuilt dev Firefox build (`bookleaf-dev@evimay.me`) via `about:debugging` and remove the old temporarily-loaded dev add-on if it lingers under the previous shared ID.
- [x] 3.2 Determine the new dev extension's redirect URI (`browser.identity.getRedirectURL()` for `bookleaf-dev@evimay.me`) and register it as an allowed callback in the dev Kinde app.
- [x] 3.3 Confirm login works in the dev build with the new redirect URI, and that the prod build's login is unaffected.
