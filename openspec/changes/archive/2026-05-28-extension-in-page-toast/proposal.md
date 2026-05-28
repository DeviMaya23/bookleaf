## Why

The extension currently uses the OS notification system to surface save results, which feels disconnected from the Bookleaf experience and varies in appearance across operating systems. Replacing it with an in-page toast gives users consistent, on-brand feedback directly where they're working.

## What Changes

- Remove OS notification calls from the extension background script
- Add a content script that receives messages from the background and renders an in-page toast
- Update `manifest.json`: drop the `notifications` permission, declare the new content script
- Toast copy:
  - **Success** — "Saved to Bookleaf." / "Added to Unsorted."
  - **Failure** — "Couldn't save image." / "Check your connection and try again."

## Capabilities

### New Capabilities

- `extension-in-page-toast`: In-page toast notification for the browser extension — a content script that listens for messages from the background service worker and renders a styled, isolated toast (via Shadow DOM) into the active tab's page.

### Modified Capabilities

- `extension-save-image`: The notification mechanism changes from `browser.notifications` to a tab message + content script. No new save behavior, but the feedback channel is different.

## Impact

- **`extensions/manifest.json`**: remove `"notifications"` from permissions; add `content_scripts` entry
- **`extensions/src/background/index.ts`**: remove `notify()`, replace with `browser.tabs.sendMessage`; thread `tabId` through `handleSave`
- **`extensions/src/content/index.ts`**: new file
- No backend, no frontend app changes
- Both Chrome and Firefox builds affected (Firefox manifest transform in `vite.config.ts` does not touch `content_scripts`, so no additional transform needed)
