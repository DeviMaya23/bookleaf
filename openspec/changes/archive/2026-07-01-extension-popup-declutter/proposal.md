## Why

The main popup's footer holds only a "Log out" button — a rarely-used action occupying the most visible real estate in the popup. Meanwhile, snip and image picker, two of the extension's most useful active features, are keyboard-shortcut-only with no discoverable entry point from the popup itself.

## What Changes

- The "Log out" button is removed from the main popup footer and moved to the bottom of the Settings view, styled in red and right-aligned to signal a destructive action.
- The main popup footer gains two icon buttons: `Scissors` (snip) and `LayoutGrid` (image picker), each with a `title` tooltip. Clicking either closes the popup and triggers the corresponding action on the active tab.
- The background service worker gains two new inbound message types (`trigger-snip`, `trigger-image-picker`) that run the same tab logic as the existing keyboard command handlers.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `extension-popup-ui`: Footer changes — "Log out" removed; two icon-button triggers added.
- `extension-popup-settings`: "Log out" row added at the bottom of the Settings view (red, right-aligned).
- `extension-snip-capture`: New trigger entry point — popup button sends `trigger-snip` to the background, which runs the same capture logic as the keyboard command.
- `extension-image-picker`: New trigger entry point — popup button sends `trigger-image-picker` to the background, which runs the same `open-image-picker` dispatch logic as the keyboard command.

## Impact

- `extensions/src/popup/App.tsx`: remove logout from footer; add snip and image picker icon buttons to footer; pass `onSnip`/`onImagePicker` handlers that send runtime messages then call `window.close()`.
- `extensions/src/popup/Settings.tsx`: add a Log out row at the bottom (red text, right-aligned); accept and wire `onLogout` prop.
- `extensions/src/background/index.ts`: add `trigger-snip` and `trigger-image-picker` cases to `runtime.onMessage`; each resolves the active tab and reuses existing tab-messaging logic.
