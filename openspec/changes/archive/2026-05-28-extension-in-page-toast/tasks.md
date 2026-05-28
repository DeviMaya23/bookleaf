## 1. Manifest

- [x] 1.1 Remove `"notifications"` from `permissions` in `extensions/manifest.json`
- [x] 1.2 Add `content_scripts` entry to `extensions/manifest.json` with `"matches": ["<all_urls>"]` pointing to `src/content/index.ts`

## 2. Content Script

- [x] 2.1 Create `extensions/src/content/index.ts` — attach a Shadow Root to a host `<div>` appended to `document.body`
- [x] 2.2 Implement `showToast(variant: 'success' | 'error', title: string, body: string)` inside the content script — builds the toast DOM inside the Shadow Root
- [x] 2.3 Write scoped CSS inside the Shadow Root: fixed bottom-right positioning, white card with `border-radius: 8px`, `box-shadow`, green/red left-border accent, bold first line, normal-weight second line, fade-out animation
- [x] 2.4 Implement auto-dismiss: remove the toast element after 4 seconds
- [x] 2.5 Register `browser.runtime.onMessage` listener — call `showToast` when `message.type === 'toast'`, ignore all other messages

## 3. Background Script

- [x] 3.1 Add `tabId: number | undefined` to the `handleSave` parameter object
- [x] 3.2 Replace the `notify()` function with a `sendToast(tabId, variant, title, body)` helper that calls `browser.tabs.sendMessage` wrapped in a try/catch (swallow rejections silently)
- [x] 3.3 Update the context menu `onClicked` listener to pass `tab?.id` into `handleSave`
- [x] 3.4 Replace all three `notify(...)` call sites in `handleSave` with `sendToast(tabId, ...)`  using the correct copy per scenario (success, failure, not-logged-in)
- [x] 3.5 Delete the old `notify()` function
