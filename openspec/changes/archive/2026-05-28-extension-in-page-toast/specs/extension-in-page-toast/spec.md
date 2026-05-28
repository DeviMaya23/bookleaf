## ADDED Requirements

### Requirement: Content script message listener
The extension SHALL include a content script (`src/content/index.ts`) declared in `manifest.json` under `content_scripts` matching `<all_urls>`. The content script SHALL listen for messages from the background service worker via `browser.runtime.onMessage` and render an in-page toast on receipt.

#### Scenario: Content script receives toast message and shows toast
- **WHEN** the background sends `{ type: 'toast', variant: 'success' | 'error', title: string, body: string }` to the active tab
- **THEN** the content script appends a toast element to the page and displays it

#### Scenario: Content script ignores unrelated messages
- **WHEN** a message with a `type` other than `'toast'` is received
- **THEN** no toast is shown and no error is thrown

### Requirement: Shadow DOM isolation
The content script SHALL render the toast inside a Shadow Root attached to a dedicated host element appended to `document.body`. All toast styles SHALL be scoped to the Shadow Root so that host-page CSS cannot affect the toast's appearance.

#### Scenario: Toast is isolated from host-page CSS
- **WHEN** a host page applies a CSS reset or conflicting class names
- **THEN** the toast's layout, colors, and typography are unaffected

### Requirement: Toast visual design
The toast SHALL be positioned fixed at the bottom-right of the viewport. It SHALL display two lines: the first line bold (the `title`), the second line in normal weight (the `body`). Success toasts SHALL use a green accent; error toasts SHALL use a red accent. The toast SHALL auto-dismiss after 4 seconds with a fade-out animation.

#### Scenario: Success toast appearance
- **WHEN** a `variant: 'success'` message is received
- **THEN** the toast is shown with a green accent, bold title, and normal-weight body text

#### Scenario: Error toast appearance
- **WHEN** a `variant: 'error'` message is received
- **THEN** the toast is shown with a red accent, bold title, and normal-weight body text

#### Scenario: Toast auto-dismisses
- **WHEN** a toast is shown
- **THEN** it fades out and is removed from the DOM after 4 seconds

### Requirement: Manifest permission update
The `manifest.json` `permissions` array SHALL NOT include `"notifications"`. The `manifest.json` SHALL declare the content script under `content_scripts`.

#### Scenario: notifications permission removed
- **WHEN** the extension manifest is inspected
- **THEN** `"notifications"` is absent from the `permissions` array

#### Scenario: content script declared
- **WHEN** the extension manifest is inspected
- **THEN** a `content_scripts` entry exists with `"matches": ["<all_urls>"]` pointing to the content script
