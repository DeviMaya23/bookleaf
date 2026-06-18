## Why

`extensions/` has no test runner and no unit tests at all, unlike the backend (which has an established usecase/handler/repository testing philosophy in CONVENTIONS.md) and the frontend (which has a working vitest suite, not yet audited). As the extension's `lib/` and `background/` logic grows (high-res image resolution, auth token exchange, save pipeline), regressions in this logic currently have no automated guard. We need a testing convention suited to the extension's actual shape — a flat `lib/` plus three entry points, not a layered backend-style architecture — and an initial test runner + test suite to make that convention real rather than aspirational.

## What Changes

- Add a test runner (vitest) to `extensions/`, configured separately from the existing build `vite.config.ts` (which is wired for the multi-target chrome/firefox `vite-plugin-web-extension` build, not test execution).
- Add a new **Browser Extension Testing** section to `CONVENTIONS.md` defining an I/O-shape classification (pure logic / thin browser-adapter / orchestrator) in place of the backend's folder-layer classification, since `extensions/` has no usecase/repository/handler split.
- Refactor `validateCandidate` in `extensions/src/lib/highResFetch.ts` into three functions to make its branching logic independently testable:
  - `validateResponseShape(response)` — pure, checks `response.ok` + content-type allowlist
  - `validateDimension(dims: { width: number; height: number })` — pure, narrow-typed (not `ImageBitmap`) so tests can pass plain literals
  - `validateCandidate` — thin orchestrator retained for the `createImageBitmap` call and `bitmap.close()`, left untested
  - No observable behavior change — `openspec/specs/extension-highres-image-resolve/spec.md`'s requirements are unaffected; this is an internal testability refactor.
- Add unit test coverage for the pure-logic and orchestrator functions in `extensions/src/lib/` (`highResRules.ts`, `highResFetch.ts`, `api.ts`, `auth.ts`, the testable parts of `storage.ts`) and `extensions/src/background/index.ts`.
- Explicitly out of scope (documented as deferred, not silently skipped):
  - `extensions/src/popup/App.tsx` — has orchestration logic (`handleLogin`/`handleLogout`/`handleToggleDark`) but no hook extraction; testing it is deferred pending a decision on whether to apply the frontend's hook-extraction-for-testability pattern here.
  - `extensions/src/content/index.ts` — DOM-only toast rendering, no branching logic worth asserting.
  - Auditing the frontend's existing test suite — separate future effort.
  - CI integration (no new GitHub Actions job) — this proposal only establishes local test infra and conventions.

## Capabilities

### New Capabilities
- `extension-test-infra`: Test runner setup (vitest config, scripts, test environment) for `extensions/`, and the I/O-shape testing classification that governs what gets a unit test there.

### Modified Capabilities
(none — the `validateCandidate` split is an internal refactor with no change to `extension-highres-image-resolve`'s documented requirements)

## Impact

- **Code**: `extensions/package.json` (new devDependencies + test script), new `extensions/vitest.config.ts`, `extensions/src/lib/highResFetch.ts` (function split), new `*.test.ts` files alongside `extensions/src/lib/*.ts` and `extensions/src/background/index.ts`.
- **Docs**: `CONVENTIONS.md` gains a Browser Extension Testing section.
- **Dependencies**: adds `vitest` (and likely `jsdom` if any test needs a DOM-like environment) to `extensions/` — flagging per CLAUDE.md's Decision Boundaries since this is a new dependency in a package that currently has none of its own test tooling.
- **No runtime behavior change** — this is test infrastructure and an internal refactor only.
