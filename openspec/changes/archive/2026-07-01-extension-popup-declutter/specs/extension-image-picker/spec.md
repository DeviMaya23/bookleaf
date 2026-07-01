## ADDED Requirements

### Requirement: Popup trigger for image picker

The background service worker SHALL handle a `{ type: "trigger-image-picker" }` runtime message. Upon receiving it, the background SHALL query the active tab in the current window and execute the same dispatch logic as `handleBrowseImagesCommand`: send `{ type: "open-image-picker" }` to the active tab's content script via `browser.tabs.sendMessage`. If the active tab has no injected content script, the send SHALL fail silently.

#### Scenario: Popup trigger opens the image picker

- **WHEN** the background receives `{ type: "trigger-image-picker" }` from the popup
- **THEN** `{ type: "open-image-picker" }` is sent to the active tab's content script
- **AND** the picker overlay appears identically to the keyboard-shortcut-triggered flow

#### Scenario: Popup trigger on a restricted page fails silently

- **WHEN** the background receives `{ type: "trigger-image-picker" }` and the active tab is a restricted page
- **THEN** the `sendMessage` call throws and the error is silently caught with no user-visible effect
