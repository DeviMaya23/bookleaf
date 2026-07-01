## ADDED Requirements

### Requirement: Popup trigger for snip capture

The background service worker SHALL handle a `{ type: "trigger-snip" }` runtime message. Upon receiving it, the background SHALL query the active tab in the current window and execute the same capture logic as `handleSnipCommand`: call `browser.tabs.captureVisibleTab`, then send `{ type: "snip-frame", dataUrl }` to the active tab's content script. If the active tab has no injected content script, the send SHALL fail silently.

#### Scenario: Popup trigger initiates the snip flow

- **WHEN** the background receives `{ type: "trigger-snip" }` from the popup
- **THEN** the active tab's visible viewport is captured via `captureVisibleTab`
- **AND** `{ type: "snip-frame", dataUrl }` is sent to the active tab's content script
- **AND** the snip overlay appears identically to the keyboard-shortcut-triggered flow

#### Scenario: Popup trigger on a restricted page fails silently

- **WHEN** the background receives `{ type: "trigger-snip" }` and the active tab is a restricted page (e.g. `chrome://`)
- **THEN** the `sendMessage` call throws and the error is silently caught with no user-visible effect
