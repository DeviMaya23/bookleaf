## Purpose

Defines the visual layout and interaction behaviour of the Bookleaf browser extension popup, covering logged-out and logged-in states, dark mode, user profile storage, and loading state handling.

---

## Requirements

### Requirement: Popup width

The popup SHALL be 320px wide. The `index.html` body width SHALL be updated from 280px to 320px.

#### Scenario: Popup renders at correct width

- **WHEN** the extension popup is opened
- **THEN** the popup container is 320px wide

---

### Requirement: Logged-out state

When no valid auth token exists, the popup SHALL render a branded logged-out layout containing:
- A header with the extension icon (`icons/icon48.png`, 16px) and the "Bookleaf" wordmark
- A centered body with a large dimmed icon, a tagline ("Save images to your Bookleaf collection as you browse the web."), and a full-width dark CTA button labelled "Log in to Bookleaf"
- A footer line: "New here? Sign up free" (the "Sign up free" text is a non-functional placeholder link in this iteration)

#### Scenario: Logged-out state renders when no auth token is stored

- **WHEN** `chrome.storage.local` has no valid `bookleaf_auth` entry
- **THEN** the popup shows the branded logged-out layout with the "Log in to Bookleaf" button

#### Scenario: Login button triggers auth flow

- **WHEN** the user clicks "Log in to Bookleaf"
- **THEN** the existing `login()` flow is invoked and the popup transitions to the logged-in state on success

#### Scenario: Login failure shows error

- **WHEN** the `login()` call throws
- **THEN** an error message "Login failed. Please try again." is displayed in the popup

---

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

---

### Requirement: Empty state

When the user is logged in but `recentSaves` is empty, the popup SHALL render the logged-in layout with the "Recently Saved" section replaced by an empty state: a dimmed icon and the text "Nothing saved yet." with a sub-line "Right-click any image to save it."

#### Scenario: Empty state renders when recentSaves is empty

- **WHEN** the user is authenticated and `recentSaves` in `chrome.storage.local` is an empty array or absent
- **THEN** the popup shows the empty state message instead of the thumbnail strip

---

### Requirement: Dark mode

The popup SHALL support a dark colour palette toggled by the icon button in the user row. The preference SHALL be persisted as `bookleaf_dark_mode` (boolean) in `chrome.storage.local` and read on popup open.

Dark palette:
- Background: `#1c1c1c`; border: `#303030`; divider: `#272727`
- Primary text: `#efefef`; secondary text: `#6e6e6e`; tertiary: `#3a3a3a`
- Hover: `#282828`; accent: `#efefef`; thumbnail border: `rgba(255,255,255,0.07)`

Light palette:
- Background: `#fff`; border: `#e8e8e8`; divider: `#f0f0f0`
- Primary text: `#1c1c1c`; secondary text: `#aaa`; tertiary: `#ddd`
- Hover: `#f5f5f5`; accent: `#1c1c1c`; thumbnail border: `rgba(0,0,0,0.06)`

#### Scenario: Dark mode preference is persisted

- **WHEN** the user toggles dark mode in the popup
- **THEN** `bookleaf_dark_mode` is updated in `chrome.storage.local`
- **AND** the popup immediately re-renders in the new palette

#### Scenario: Dark mode preference is restored on open

- **WHEN** the popup is opened and `bookleaf_dark_mode` is `true` in storage
- **THEN** the popup renders in the dark palette without a flash

#### Scenario: Light mode is the default

- **WHEN** the popup is opened and `bookleaf_dark_mode` is absent from storage
- **THEN** the popup renders in the light palette

---

### Requirement: Username and avatar stored at login

The `login()` function SHALL, after a successful token exchange, decode the `id_token` JWT payload (base64url, middle segment) to extract the user's display name and avatar. It SHALL try `given_name`, then `name`, then `email` for the display name. It SHALL read `picture` for the avatar URL. The display name SHALL be stored under `bookleaf_username` and the avatar URL (if present) under `bookleaf_avatar` in `chrome.storage.local`.

#### Scenario: Username and avatar extracted and stored on successful login

- **WHEN** `login()` completes successfully and the `id_token` payload contains `given_name` and `picture`
- **THEN** `bookleaf_username` is stored with the value of `given_name`
- **AND** `bookleaf_avatar` is stored with the value of `picture`

#### Scenario: Fallback to name when given_name is absent

- **WHEN** the `id_token` payload has no `given_name` but has `name`
- **THEN** `bookleaf_username` is stored with the value of `name`

#### Scenario: Username falls back to email

- **WHEN** the `id_token` payload has neither `given_name` nor `name` but has `email`
- **THEN** `bookleaf_username` is stored with the value of `email`

#### Scenario: Avatar is not stored when picture is absent

- **WHEN** the `id_token` payload has no `picture` claim
- **THEN** `bookleaf_avatar` is not written to storage
- **AND** the popup renders the gradient avatar placeholder instead

#### Scenario: Profile decoded without extra network call

- **WHEN** `login()` exchanges the auth code for tokens
- **THEN** no additional HTTP request to `/userinfo` is made to retrieve the username or avatar

---

### Requirement: Loading state

While the popup is reading initial state from `chrome.storage.local`, it SHALL render a blank container with no flicker of incorrect content.

#### Scenario: Loading state shown while storage is being read

- **WHEN** the popup mounts and the `Promise.all` reading auth, username, recentSaves, and darkMode has not resolved
- **THEN** no content (other than the empty container) is rendered
