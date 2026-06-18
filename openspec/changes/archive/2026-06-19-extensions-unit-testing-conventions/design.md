## Context

`extensions/` currently has zero test infrastructure: no test runner in `package.json`, no `*.test.ts` files, no test-related devDependencies. Its `vite.config.ts` is a single mode-driven build config wired to `vite-plugin-web-extension`, which transforms `manifest.json` and expects the extension's actual entry points (`background`, `content`, `popup`) — it is not a vitest-friendly config to extend in place.

The backend's testing philosophy (CONVENTIONS.md "Backend Testing") is built on a folder-layer split (`usecase/` / `repository/` / `handler/`) that doesn't exist in `extensions/` — there's a flat `lib/` and three entry points. This design adapts the backend's underlying criterion (test logic that lives exclusively in your code; don't test pass-through; don't unit-test thin external-system adapters) to an I/O-shape classification instead of a folder classification.

## Goals / Non-Goals

**Goals:**
- Stand up a working vitest setup in `extensions/` isolated from the build config.
- Define a reusable classification (pure logic / thin browser-adapter / orchestrator) in CONVENTIONS.md that future extension code can be sorted against without re-litigating it per PR.
- Land initial test coverage for `lib/` and `background/index.ts` using that classification, including the `validateCandidate` split.

**Non-Goals:**
- Testing `popup/App.tsx` or `content/index.ts` (deferred, documented in proposal.md).
- Auditing or changing the frontend's existing vitest suite.
- Wiring a CI job for extension tests (no `ci-pr-extension-checks` workflow in this change).
- Introducing ESLint to `extensions/` — it has none today, and adding one is an unrelated decision.

## Decisions

### Separate `vitest.config.ts`, not a `test` block in `vite.config.ts`
`extensions/vite.config.ts` is a function of `mode` that conditionally wires `webExtension({ browser, transformManifest })`, which expects the real manifest/entry-point structure to transform. Running vitest against that config would invoke that plugin for no reason and couple test execution to build-mode branching it doesn't need. A dedicated `vitest.config.ts` (same pattern as keeping concerns separate, though frontend itself merges them since its `vite.config.ts` has no comparable build-time plugin complexity) keeps test config minimal and independent of build-target changes.

**Alternative considered**: add a `test` block to the existing `vite.config.ts`. Rejected — would run the webExtension plugin during `vitest run` for no benefit and create coupling between build-target changes and test config.

### Test environment: `node`, not `jsdom`
Every function in scope (`lib/`, `background/index.ts`) is logic over URLs, JSON, fetch responses, and browser extension APIs — none touch the DOM. `jsdom` is an extra dependency this change doesn't need. If `popup/App.tsx` testing happens in a later change, `jsdom` can be added then, scoped to that work.

**Alternative considered**: match frontend's `jsdom` environment for consistency. Rejected for now — adding a dependency this change has no use for isn't justified by future, not-yet-decided work.

### Mocking `webextension-polyfill` via `vi.mock`
`background/index.ts` registers listeners (`browser.runtime.onInstalled.addListener`, `browser.contextMenus.onClicked.addListener`) at module load time, so importing it in a test executes that registration immediately. Tests need a `browser` mock with `storage.local.{get,set,remove}`, `tabs.sendMessage`, `contextMenus.{removeAll,create}`, `runtime.onInstalled.addListener`, and `contextMenus.onClicked.addListener` all present as `vi.fn()` (the listener-registration ones as no-ops that don't throw). This mock is shared across `lib/` and `background/` tests, so it lives in one colocated helper rather than being redefined per test file.

### No fakes — only value-return spies
The backend convention distinguishes fakes (pre-state + post-state, for things like `usecase` tests against repository interfaces) from value-return spies (single-step operations, external services). Nothing in `extensions/` reads existing state to decide what to write the way a backend usecase does against a repository — `storage.ts` getters/setters are independent pass-throughs, not stateful collaborators a test needs to seed. So `extensions/` testing only ever needs value-return spies (`vi.fn().mockResolvedValue(...)`), never fakes. This is stated explicitly in the CONVENTIONS.md addition so it isn't rediscovered ad hoc later.

### Test colocation, matching the existing FE convention
CONVENTIONS.md already establishes colocated `*.test.ts` next to the unit it covers ("Test Colocation" under Frontend Development Practices). This isn't a new pattern being introduced for extensions — it's reusing one already documented and is independent of the (deferred) frontend test-suite audit, which is about whether existing FE *tests* follow good practice, not about whether the colocation file-placement rule itself is sound.

### `validateCandidate` split
Per proposal.md: `validateResponseShape(response)` and `validateDimension(dims: { width: number; height: number })` are extracted as pure, independently testable functions. `validateDimension` takes the narrow `{ width, height }` shape rather than `ImageBitmap` specifically so tests pass plain object literals instead of needing a real or mocked `ImageBitmap`. `validateCandidate` keeps the `createImageBitmap` call and `bitmap.close()` and is not unit tested, consistent with `generateThumbnail` staying untested — native canvas/bitmap APis are not unit tested anywhere in this change.

## Risks / Trade-offs

- **Two vite-family configs in one package** (`vite.config.ts` for build, `vitest.config.ts` for tests) could drift or confuse contributors about which one governs what → Mitigated by each having a single, narrow responsibility and the design doc stating the split explicitly.
- **`validateCandidate`'s split changes its internal call graph** but not its behavior — a reviewer skimming the diff without `extension-highres-image-resolve/spec.md` open might mistake it for a behavior change → Mitigated by proposal.md explicitly calling out "no observable behavior change."
- **Mock-heavy orchestrator tests** (`saveImage`, `handleSave`, `login`) risk testing call-sequence rather than outcome if not careful → Mitigated by reusing the backend convention's assertion-quality rule: assert the result/outcome (e.g., toast variant sent, recent-save recorded, thrown error mapped), not that a particular mock was called.

## Migration Plan

Not applicable — this is additive test infrastructure with no runtime behavior change and no existing tests to migrate.
