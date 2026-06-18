## Purpose

Provide test infrastructure and conventions for unit testing the browser extension (`extensions/`), enabling fast, isolated tests for pure logic and orchestrator functions without depending on the extension build pipeline or real browser APIs.

## Requirements

### Requirement: Standalone vitest configuration
`extensions/` SHALL have a `vitest.config.ts` separate from `vite.config.ts`, using the `node` test environment, and `package.json` SHALL expose `test` (`vitest run`) and `test:watch` (`vitest`) scripts.

#### Scenario: Test run does not invoke the build's webExtension plugin
- **WHEN** `npm test` is run in `extensions/`
- **THEN** the `vite-plugin-web-extension` build plugin from `vite.config.ts` is not invoked

#### Scenario: Build config is unaffected by test config
- **WHEN** `npm run build` is run in `extensions/`
- **THEN** it uses `vite.config.ts` unchanged, with no dependency on `vitest.config.ts`

### Requirement: I/O-shape test classification
Functions in `extensions/src/lib/` and `extensions/src/background/index.ts` SHALL be classified as one of: pure logic (no I/O), thin browser-adapter (pure passthrough to a browser/fetch API, e.g. `storage.ts` getters/setters), or orchestrator (branches, assembles, or catches errors around calls into adapters or pure logic). Pure logic and orchestrator functions SHALL have unit test coverage. Thin browser-adapters and functions whose only logic wraps a native browser API with no test-environment equivalent (`createImageBitmap`, `OffscreenCanvas`) SHALL be exempt from unit testing.

#### Scenario: Pure logic function is unit tested
- **WHEN** a function in `lib/` contains branching or transformation logic with no I/O (e.g. `resolveHighResUrl`, `decodeJwtPayload`)
- **THEN** it has a corresponding unit test

#### Scenario: Thin adapter function is not unit tested
- **WHEN** a function only forwards to a browser storage/fetch API with no branching logic (e.g. `getAuth`, `setDarkMode`)
- **THEN** it has no unit test

#### Scenario: Orchestrator function is tested with mocked I/O boundary
- **WHEN** a function branches or assembles a result from calls to adapters or pure logic (e.g. `resolveImageBlob`, `saveImage`, `handleSave`, `login`)
- **THEN** it has a unit test that mocks the adapter/I/O boundary and asserts the orchestrator's outcome

#### Scenario: Native-browser-API-bound function is exempt
- **WHEN** a function's only logic wraps `createImageBitmap` or `OffscreenCanvas` with no test-environment equivalent (e.g. `generateThumbnail`)
- **THEN** it has no unit test

### Requirement: Shared webextension-polyfill mock
The test suite SHALL provide a shared mock of the `browser` global (covering `storage.local.{get,set,remove}`, `tabs.sendMessage`, `contextMenus.{removeAll,create}`, `contextMenus.onClicked.addListener`, `runtime.onInstalled.addListener`) usable across `lib/` and `background/` test files, so that importing `background/index.ts` — which registers listeners at module load time — does not throw in a test environment.

#### Scenario: Importing background/index.ts in a test does not throw
- **WHEN** a test file imports `extensions/src/background/index.ts` with the shared `browser` mock in place
- **THEN** the import succeeds without throwing due to a missing `browser` API

### Requirement: validateCandidate split into testable pure functions
`extensions/src/lib/highResFetch.ts` SHALL expose `validateResponseShape(response: Response): boolean` and `validateDimension(dims: { width: number; height: number }): boolean` as independently callable pure functions, each unit tested. `validateCandidate` SHALL retain orchestration of the `createImageBitmap` call and `bitmap.close()`, calling both extracted functions, and SHALL NOT have direct unit test coverage (per the native-browser-API exemption).

#### Scenario: validateResponseShape rejects a non-OK response
- **WHEN** `validateResponseShape` is called with a `Response` whose `ok` is `false`
- **THEN** it returns `false`

#### Scenario: validateResponseShape rejects a disallowed content type
- **WHEN** `validateResponseShape` is called with an `ok` `Response` whose `Content-Type` is not `image/jpeg`, `image/png`, or `image/webp`
- **THEN** it returns `false`

#### Scenario: validateResponseShape accepts an allowed content type
- **WHEN** `validateResponseShape` is called with an `ok` `Response` whose `Content-Type` is one of `image/jpeg`, `image/png`, `image/webp`
- **THEN** it returns `true`

#### Scenario: validateDimension rejects undersized dimensions
- **WHEN** `validateDimension` is called with `{ width, height }` where either value is below 100
- **THEN** it returns `false`

#### Scenario: validateDimension accepts dimensions at or above the threshold
- **WHEN** `validateDimension` is called with `{ width, height }` both at least 100
- **THEN** it returns `true`

### Requirement: Value-return spies only, no fakes
Extension unit tests SHALL use value-return spies for all mocked dependencies. Fakes (test doubles with seeded pre-state and asserted post-state) SHALL NOT be used, since no function in `extensions/src/lib/` or `extensions/src/background/index.ts` reads existing state from a dependency to decide what to write.

#### Scenario: Orchestrator test uses a value-return spy
- **WHEN** a test for an orchestrator function (e.g. `saveImage`) mocks a dependency (e.g. `apiFetch`)
- **THEN** the mock is a spy configured to return a fixed value, not a fake with seeded state
