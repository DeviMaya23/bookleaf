## MODIFIED Requirements

### Requirement: Logged-in state

When a valid auth token exists, the popup SHALL render:
- A header with the extension icon, the "Bookleaf" wordmark, and an "Open ↗" button that opens the Bookleaf web app URL (`VITE_APP_URL`) in a new tab
- A user row with an avatar (`<img>` using stored `bookleaf_avatar` URL if available, otherwise a gradient placeholder div), the stored username, and a dark/light mode toggle icon button
- A "Recently Saved" section (see `extension-recent-saves` spec for data source)
- A footer with two icon buttons: a `Scissors` icon (title `"Snip"`) and a `LayoutGrid` icon (title `"Pick images"`). The footer SHALL NOT contain a "Log out" button.

#### Scenario: Logged-in state renders when valid auth token is stored

- **WHEN** `chrome.storage.local` has a valid non-expired `bookleaf_auth` entry
- **THEN** the popup shows the logged-in layout with header, user row, recently saved section, and footer with two icon buttons

#### Scenario: Open button opens the web app

- **WHEN** the user clicks "Open ↗"
- **THEN** `chrome.tabs.create({ url: VITE_APP_URL })` is called and the popup closes

#### Scenario: Snip button sends trigger message and closes popup

- **WHEN** the user clicks the Scissors icon button in the footer
- **THEN** `browser.runtime.sendMessage({ type: "trigger-snip" })` is sent to the background
- **AND** `window.close()` is called to dismiss the popup

#### Scenario: Image picker button sends trigger message and closes popup

- **WHEN** the user clicks the LayoutGrid icon button in the footer
- **THEN** `browser.runtime.sendMessage({ type: "trigger-image-picker" })` is sent to the background
- **AND** `window.close()` is called to dismiss the popup
