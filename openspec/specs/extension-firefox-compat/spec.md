# Spec: Extension Firefox Compatibility

## Purpose

Defines the requirements for making the Bookleaf browser extension work on both Chrome and Firefox from a single codebase, covering the dual-target build system, cross-browser API usage via polyfill, and graceful handling of browser-specific capability gaps.

## Requirements

### Requirement: Extension builds for both Chrome and Firefox from a single codebase

The extension project SHALL produce a working build for Chrome (`npm run build`) and working builds for Firefox (`npm run build:firefox` for dev, `npm run build:firefox:prod` for production) from the same source files. All source `.ts` and `.tsx` files SHALL use `browser.*` APIs via `webextension-polyfill` exclusively — no direct `chrome.*` calls.

Chrome output SHALL be written to `dist/chrome/`. Firefox output SHALL be written to `dist/firefox/`. A `build:all` script SHALL run both the Chrome build and the dev Firefox build in sequence. `make ext-build` SHALL invoke `build:all`.

The Firefox build SHALL transform the manifest via `vite.config.ts`:
- Inject `browser_specific_settings.gecko.id`, with the value determined by build mode:
  - `firefox` (dev) → `"bookleaf-dev@evimay.me"`
  - `firefox-production` (prod) → `"bookleaf@evimay.me"`
- Convert `background.service_worker` to `background.scripts` (array containing the same path), removing `background.type`

Neither transformation SHALL appear in the Chrome build.

#### Scenario: Dev Firefox build produces a manifest with the dev gecko ID

- **WHEN** `npm run build:firefox` is run
- **THEN** the output at `dist/firefox/manifest.json` contains `browser_specific_settings.gecko.id` set to `"bookleaf-dev@evimay.me"`
- **AND** `dist/firefox/manifest.json` contains `background.scripts` as an array
- **AND** `dist/firefox/manifest.json` does not contain `background.service_worker`

#### Scenario: Production Firefox build produces a manifest with the production gecko ID

- **WHEN** `npm run build:firefox:prod` is run
- **THEN** the output at `dist/firefox/manifest.json` contains `browser_specific_settings.gecko.id` set to `"bookleaf@evimay.me"`
- **AND** `dist/firefox/manifest.json` contains `background.scripts` as an array
- **AND** `dist/firefox/manifest.json` does not contain `background.service_worker`

#### Scenario: Chrome build does not include gecko fields or scripts background

- **WHEN** `npm run build` is run
- **THEN** the output at `dist/chrome/manifest.json` does not contain `browser_specific_settings`
- **AND** `dist/chrome/manifest.json` contains `background.service_worker`

### Requirement: Context menu registers on both browsers via polyfill

The background script SHALL use `browser.contextMenus` and `browser.runtime` (from `webextension-polyfill`) for all context menu registration and runtime event handling. The `onInstalled` listener SHALL be declared `async` and SHALL `await browser.contextMenus.removeAll()` before calling `browser.contextMenus.create`.

#### Scenario: Context menu item exists after install on Firefox

- **WHEN** the Firefox extension is installed or updated
- **THEN** a single context menu item "Save to Bookleaf" is registered for image elements
- **AND** no duplicate menu items exist from a previous install

#### Scenario: Context menu item exists after install on Chrome

- **WHEN** the Chrome extension is installed or updated
- **THEN** a single context menu item "Save to Bookleaf" is registered for image elements

### Requirement: Thumbnail generation uses a capability guard for OffscreenCanvas

`OffscreenCanvas` is supported in Firefox 105+ (released Sept 2022) and is available in background scripts on both Chrome and modern Firefox. Thumbnail generation SHALL be wrapped in a `typeof OffscreenCanvas !== "undefined"` guard as a defensive fallback for environments where it may be absent. When the guard evaluates to false, the extension SHALL still record the save to recent saves storage with an empty `dataUrl`.

#### Scenario: Thumbnail generated on Chrome

- **WHEN** an image is saved successfully on Chrome
- **THEN** a 60×60 JPEG thumbnail is generated and stored as `dataUrl`
- **AND** the thumbnail is displayed in the popup's recent saves strip

#### Scenario: Thumbnail generated on modern Firefox (105+)

- **WHEN** an image is saved successfully on Firefox 105+
- **THEN** `typeof OffscreenCanvas !== "undefined"` evaluates to true
- **AND** a 60×60 JPEG thumbnail is generated and stored as `dataUrl`
- **AND** the thumbnail is displayed in the popup's recent saves strip

#### Scenario: Save recorded without thumbnail when OffscreenCanvas is unavailable

- **WHEN** an image is saved successfully in an environment where `OffscreenCanvas` is unavailable
- **THEN** a success notification is shown
- **AND** `addRecentSave` is called with `dataUrl: ""`
- **AND** the entry appears in the recent saves list in the popup
