## 1. Link permalink rule table (Twitter, Facebook)

- [x] 1.1 Create `extensions/src/lib/linkPermalinkRules.ts` exporting a rule table (`{ id, matches(url: URL): boolean }`) and a `resolveLinkPermalink(linkUrl: string): boolean` (or equivalent) helper, with a Twitter status-permalink rule and a Facebook post-permalink rule
- [x] 1.2 Unit test `linkPermalinkRules.ts` per `extension-test-infra` (pure logic): matching cases for Twitter and Facebook, non-matching case for an unregistered site

## 2. Content script DOM resolution (Pinterest)

- [x] 2.1 Add a Pinterest DOM-resolution site rule (matcher for when the content script should attempt resolution) and a pure `resolveCardImageSrc(target: Element): string | null` function performing the `closest('a')` → `querySelector('img')` walk
- [x] 2.2 Unit test `resolveCardImageSrc` per `extension-test-infra` (pure logic) using minimal mock DOM-like objects: finds an image when present, returns `null` when no enclosing link or no descendant `<img>` exists
- [x] 2.3 Wire a capture-phase `contextmenu` listener into `extensions/src/content/index.ts` that calls the Pinterest matcher + `resolveCardImageSrc`, and on a hit sends `{ resolved: { srcUrl } }` via `runtime.sendMessage` (payload typed as `Partial<{ srcUrl: string; title: string }>`)

## 3. Background: per-tab resolved-context store

- [x] 3.1 Add a per-tab in-memory store (e.g. `Map<number, Partial<{ srcUrl: string; title: string }>>`) in `extensions/src/background/index.ts`, with a `runtime.onMessage` listener that overwrites the entry for the sender's `tabId` on each resolved-context message
- [x] 3.2 Add the shared `browser` mock support needed for `runtime.onMessage.addListener` per `extension-test-infra`'s shared polyfill mock requirement, if not already covered

## 4. Background: second context menu item

- [x] 4.1 Register a second `contextMenus.create` item (e.g. `save-to-bookleaf-link`) with `contexts: ["link"]` and `targetUrlPatterns` covering registered link-only card site patterns (starting with Pinterest pin URLs), alongside the existing image-context item, inside the existing `onInstalled` listener
- [x] 4.2 Update the shared `onClicked` listener to branch on `info.menuItemId` between the image-context and link-context items

## 5. Background: click handler resolution logic

- [x] 5.1 For the image-context item: after reading `info.srcUrl`, check `info.linkUrl` against the Section 1 rule table; if matched, use `info.linkUrl` as `source_url` instead of `info.pageUrl`
- [x] 5.2 For the link-context item: read the Section 3 per-tab store for the resolved `srcUrl`; if present, use it as the effective `srcUrl` and set `source_url` to `info.linkUrl`; if absent, send the existing "Couldn't save image." error toast and abort without attempting an upload
- [x] 5.3 Update `handleSave`'s call site(s) to pass through the effective `srcUrl`/`source_url` computed above (no change to `handleSave`'s own signature/logic beyond what it already accepts)
- [x] 5.4 Unit test the updated click-handler logic per `extension-test-infra` (orchestrator): linkUrl override applied/not applied for image-context; resolved-srcUrl used/absent for link-context, including the no-upload-on-missing-resolution case

## 6. Verification

- [x] 6.1 Grep `extensions/src` for all call sites of `handleSave`, `contextMenus.onClicked`, and `info.pageUrl`/`info.linkUrl` usage to confirm no other entry point bypasses the new resolution logic
- [x] 6.2 Manually verify in a loaded dev build: Twitter media-tab save records the tweet permalink as `source_url`; Facebook post save records the post permalink; Pinterest grid card save succeeds with a high-res image and the pin permalink as `source_url`
- [x] 6.3 Run `npm run build` and `npm run type-check` in `extensions/` and fix any issues (`extensions/` has no `lint` script, unlike `frontend/`; `type-check` is its closest equivalent)
- [x] 6.4 Run `npm test` in `extensions/` and confirm all new and existing unit tests pass
