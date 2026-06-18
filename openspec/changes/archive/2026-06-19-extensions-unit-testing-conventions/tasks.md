## 1. Test Runner Setup

- [x] 1.1 Add `vitest` as a devDependency in `extensions/package.json`
- [x] 1.2 Create `extensions/vitest.config.ts` with `environment: 'node'`, isolated from `vite.config.ts`
- [x] 1.3 Add `test` (`vitest run`) and `test:watch` (`vitest`) scripts to `extensions/package.json`
- [x] 1.4 Create a shared `webextension-polyfill` mock helper (covering `storage.local.{get,set,remove}`, `tabs.sendMessage`, `contextMenus.{removeAll,create}`, `contextMenus.onClicked.addListener`, `runtime.onInstalled.addListener`) usable via `vi.mock` across `lib/` and `background/` test files
- [x] 1.5 Verify `npm test` runs (even with zero test files) and `npm run build` still succeeds, confirming the two configs don't interfere

## 2. CONVENTIONS.md Update

- [x] 2.1 Add a "Browser Extension Testing" section to `CONVENTIONS.md` documenting the I/O-shape classification (pure logic / thin browser-adapter / orchestrator), the native-browser-API exemption (`createImageBitmap`, `OffscreenCanvas`), the value-return-spies-only rule (no fakes), and that `popup/` and `content/index.ts` are deferred out of scope for now

## 3. validateCandidate Split

- [x] 3.1 Extract `validateResponseShape(response: Response): boolean` from `validateCandidate` in `extensions/src/lib/highResFetch.ts`
- [x] 3.2 Extract `validateDimension(dims: { width: number; height: number }): boolean` from `validateCandidate`, narrow-typed (not `ImageBitmap`)
- [x] 3.3 Update `validateCandidate` to call both extracted functions plus the retained `createImageBitmap`/`bitmap.close()` orchestration, preserving current behavior
- [x] 3.4 Confirm `extension-highres-image-resolve` spec scenarios still hold (manual check against `openspec/specs/extension-highres-image-resolve/spec.md` — no spec delta needed since behavior is unchanged)

## 4. Unit Tests — `extensions/src/lib/`

- [x] 4.1 `highResRules.test.ts`: Twitter media match/no-match/transform, Pinterest match/transform
- [x] 4.2 `highResFetch.test.ts`: `resolveHighResUrl` (rule match, no match, invalid URL), `validateResponseShape` (ok/not-ok, allowed/disallowed content-type), `validateDimension` (below/at/above threshold)
- [x] 4.3 `api.test.ts`: `apiFetch` sets `Authorization` header when auth present, omits it when absent
- [x] 4.4 `auth.test.ts`: `decodeJwtPayload` (valid token, malformed token, non-JSON payload), `buildAuthUrl` (with/without audience param)
- [x] 4.5 `auth.test.ts`: `exchangeCodeForTokens` and `login` orchestration — username fallback chain (`given_name` → `name` → `email`), conditional `setAvatar` call, throw on non-OK token response (mock `fetch`, `browser.identity`, and `storage.ts` setters as value-return spies)
- [x] 4.6 `storage.test.ts`: `addRecentSave` caps at 5 entries and prepends newest first

## 5. Unit Tests — `extensions/src/background/index.ts`

- [x] 5.1 `isTokenValid`: valid (not expired), invalid (expired), invalid (null auth)
- [x] 5.2 `blobToDataUrl`: produces expected `data:` URL for a known blob
- [x] 5.3 `resolveImageBlob`: returns high-res candidate when valid, falls back to original `srcUrl` when candidate fetch/validation fails (mock `fetch`, `resolveHighResUrl`, `validateCandidate` as spies)
- [x] 5.4 `saveImage`: happy path assembles `image_id` from init → parallel PUTs → complete; throws on non-OK at each step (init, PUT, complete) — mock `apiFetch`/`fetch` as spies
- [x] 5.5 `handleSave`: sends error toast and returns early when auth invalid; sends success toast and records recent save on success; sends error toast and skips recent save on thrown error (mock `getAuth`, `resolveImageBlob`/`saveImage` path, `addRecentSave`, `sendToast`'s underlying `browser.tabs.sendMessage` as spies)

## 6. Verification

- [x] 6.1 Run `npm test` in `extensions/` and confirm all new tests pass
- [x] 6.2 Run `npm run type-check` in `extensions/` and fix any type errors
- [x] 6.3 Run `npm run build` in `extensions/` and confirm the build is unaffected
