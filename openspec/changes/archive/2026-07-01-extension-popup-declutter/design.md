## Context

The extension popup currently has two entry points for triggering active features (snip, image picker) — both keyboard-only. The popup footer contains a single "Log out" button that is rarely needed. The Settings view already exists as a separate view within the popup and is the appropriate home for destructive account actions.

The background service worker already handles snip and image picker via `browser.commands.onCommand`. The message listener in `background/index.ts` handles `drag-save`, `snip-captured`, `video-frame-captured`, and `picker-save` — but has no inbound message types for initiating a feature trigger from the popup.

## Goals / Non-Goals

**Goals:**
- Replace the footer's logout button with two icon-button triggers (snip, image picker)
- Move logout into the Settings view as a clearly destructive, right-aligned red action
- Enable both triggers to be fired from the popup without duplicating the tab-capture logic

**Non-Goals:**
- Changing the snip or image picker flow itself — only the trigger entry point changes
- Adding any new visual feedback in the popup about trigger success/failure (toasts already appear in the page via the existing in-page toast system)
- Keyboard shortcut changes

## Decisions

### 1. New message types in the background rather than calling tab APIs from the popup

The popup cannot call `browser.tabs.captureVisibleTab` or `browser.tabs.sendMessage` directly — these require permissions that are only available to the background service worker. The popup sends a lightweight trigger message (`trigger-snip` / `trigger-image-picker`) to the background, which resolves the active tab and runs the existing logic.

This keeps all tab-interaction code in the background, consistent with the existing architecture. The alternative — granting the popup the `tabs` permission — would expand the extension's permission surface unnecessarily.

### 2. Popup closes itself with `window.close()` before the action takes effect

For snip, `captureVisibleTab` captures the visible viewport. If the popup is still open when the capture runs, the screenshot includes the popup overlay on top of the page. Calling `window.close()` in the popup before sending the message avoids this.

For image picker the popup closing is less critical (the picker overlay appears inside the page), but consistent UX is better — closing the popup always feels right after triggering an action.

`browser.runtime.sendMessage` is awaited first, then `window.close()` is called. This guarantees the message is delivered before the popup tears down. The alternative (close first, then send) is also safe in practice since `window.close()` schedules the close asynchronously, but awaiting the send first is strictly more reliable.

### 3. Logout moves to Settings as a bottom row, not a separate "Account" section

The Settings view is short (dark mode, drag-to-save, two hotkey rows). Logout fits naturally as a terminal row without needing a new section header. It is styled red and right-aligned to visually distinguish it from the toggle rows above it.

The alternative — a dedicated "Account" view — would add unnecessary navigation depth for a single action.

### 4. Footer becomes two equal-width icon buttons (no labels)

The footer is `padding: 8px 14px 12px` in a 320px-wide popup. Both buttons use `flex: 1` so they split the available width equally, with a fixed `height: 30px`. Labels are omitted to keep the footer compact; `title` attributes provide tooltip text for discoverability. This is consistent with the existing Settings gear icon in the header, which is also label-free with a `title`.

Icons: `Scissors` (lucide-react) for snip, `LayoutGrid` (lucide-react) for image picker. Both are already available in the `lucide-react` dependency used by `App.tsx`.

## Risks / Trade-offs

- **`window.close()` timing vs. message delivery**: In practice this is safe across Chrome and Firefox — the message channel remains open until the window finishes unloading. If this ever regresses in a future browser version, the fallback is to send the message first and close on the message's callback. Worth noting but not worth defensive code now.
- **No success/failure feedback in popup**: The popup is closed when the action runs, so there is no channel to surface errors back to it. This is acceptable because the existing in-page toast system already handles all success/failure feedback for both features.
- **Snip unavailable on restricted pages**: If the user triggers snip from the popup while on a `chrome://` or `about:` page, the background will fail silently (same as the keyboard shortcut). No regression — just the same existing limitation.
