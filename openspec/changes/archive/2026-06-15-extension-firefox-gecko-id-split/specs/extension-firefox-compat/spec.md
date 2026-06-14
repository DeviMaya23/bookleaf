## MODIFIED Requirements

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
