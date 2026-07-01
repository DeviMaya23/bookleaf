# Spec: Extension Image Picker

## Purpose

Defines the image picker feature for the "Save to Bookleaf" browser extension: how the keyboard command opens a picker overlay showing all images found on the current page, how images are collected and ranked, how the user selects images and confirms a batch save, and how aggregated feedback is shown.

## Requirements

### Requirement: Browse images keyboard command

The `browse-images` command SHALL be declared in `manifest.json` under `commands` with suggested key `Alt+Shift+I` for all platforms and description `"Browse images on this page"`. The background service worker SHALL handle this command in `browser.commands.onCommand` by querying the active tab and sending `{ type: "open-image-picker" }` to that tab's content script via `browser.tabs.sendMessage`. If the active tab has no injected content script (e.g. a `chrome://` or `about:` page), the send SHALL fail silently.

#### Scenario: Keyboard shortcut triggers picker

- **WHEN** the user presses `Alt+Shift+I` on a normal web page
- **THEN** the background sends `{ type: "open-image-picker" }` to the active tab's content script

#### Scenario: Shortcut on a restricted page is a no-op

- **WHEN** the user presses `Alt+Shift+I` on a page where the content script is not injected
- **THEN** the `sendMessage` call throws and the error is silently caught with no user-visible effect

### Requirement: Image collection from page DOM

The content script SHALL export a pure function `collectPageImages(document: Document): PageImage[]` in `src/lib/imagePicker.ts`. The function SHALL query all `img` elements via `document.querySelectorAll('img')`, exclude elements matching either filter below, resolve the best source URL and srcset width for each remaining element via `resolveBestImageSrc`, and return results sorted descending by sort key: `srcsetWidth` when available, otherwise `√(naturalWidth × naturalHeight)` (so both metrics are in comparable linear units).

Exclusion filters:
- `src` is empty or starts with `blob:`
- `img.complete === false && img.naturalWidth === 0` (lazy-loaded, not yet decoded)

`PageImage` type: `{ src: string; naturalWidth: number; naturalHeight: number; srcsetWidth: number | null }`

#### Scenario: Images sorted by srcsetWidth when available

- **WHEN** an image has a `srcset` with a `1024w` entry but the browser loaded it at `550 × 550` (due to `sizes`), alongside another image with `naturalWidth` of `600`
- **THEN** the srcset image sorts above the 600px image because `srcsetWidth` (1024) beats `√(600 × h)`

#### Scenario: Images without srcset sorted by naturalWidth

- **WHEN** the page contains multiple loaded images with no `srcset`
- **THEN** `collectPageImages` returns them ordered from largest `naturalWidth × naturalHeight` to smallest

#### Scenario: Lazy-loaded image not yet decoded is excluded

- **WHEN** an `img` element has `complete === false` and `naturalWidth === 0`
- **THEN** `collectPageImages` does not include it in the result

#### Scenario: Image with empty src is excluded

- **WHEN** an `img` element has an empty `src` attribute
- **THEN** `collectPageImages` does not include it in the result

#### Scenario: Image with blob src is excluded

- **WHEN** an `img` element has a `src` starting with `blob:`
- **THEN** `collectPageImages` does not include it in the result

#### Scenario: Page with no qualifying images returns empty array

- **WHEN** all `img` elements on the page are excluded by the filters
- **THEN** `collectPageImages` returns an empty array

### Requirement: srcset parsing for highest-resolution URL

The content script SHALL export a pure function `parseSrcset(srcset: string): { url: string; width: number } | null` in `src/lib/imagePicker.ts`. The function SHALL parse the `srcset` attribute string, identify all entries with a `w`-descriptor (e.g. `image.jpg 1600w`), and return the URL and intrinsic pixel width of the entry with the highest `w` value. If the string is empty or contains no valid `w`-descriptor entries, the function SHALL return `null`.

#### Scenario: Multiple w-descriptor entries returns widest entry

- **WHEN** `parseSrcset` receives `"small.jpg 400w, medium.jpg 800w, large.jpg 2400w"`
- **THEN** it returns `{ url: "large.jpg", width: 2400 }`

#### Scenario: Single w-descriptor entry returns its entry

- **WHEN** `parseSrcset` receives `"image.jpg 1200w"`
- **THEN** it returns `{ url: "image.jpg", width: 1200 }`

#### Scenario: No w-descriptors returns null

- **WHEN** `parseSrcset` receives a srcset string with only x-descriptors or no descriptors
- **THEN** it returns `null`

#### Scenario: Empty string returns null

- **WHEN** `parseSrcset` receives an empty string
- **THEN** it returns `null`

### Requirement: Best source URL resolution

The content script SHALL export a pure function `resolveBestImageSrc(img: HTMLImageElement): { src: string; srcsetWidth: number | null }` in `src/lib/imagePicker.ts`. Resolution SHALL follow this priority order:

1. Resolve a srcset string from `img.srcset` if non-empty, otherwise fall back to the `data-srcset` attribute (used by lazy-loading libraries such as the WordPress Lazy Load plugin that store the real srcset in `data-srcset` before the image is scrolled into view). If either yields a non-null result from `parseSrcset`, return `{ src: entry.url, srcsetWidth: entry.width }`.
2. If `resolveHighResUrl(img.src)` returns a non-null URL (platform-specific upgrade, e.g. Twitter, Pinterest), return `{ src: upgradedUrl, srcsetWidth: null }`.
3. Return `{ src: img.src, srcsetWidth: null }`.

#### Scenario: srcset widest URL takes priority over src

- **WHEN** an img has both a `srcset` with w-descriptors and a `src`
- **THEN** `resolveBestImageSrc` returns `{ src: <widest url>, srcsetWidth: <width value> }`

#### Scenario: data-srcset used when srcset is empty

- **WHEN** an img has an empty `srcset` attribute but a non-empty `data-srcset` attribute with w-descriptors (lazy-loaded, not yet scrolled into view)
- **THEN** `resolveBestImageSrc` returns `{ src: <widest url>, srcsetWidth: <width value> }` from `data-srcset`

#### Scenario: High-res platform rule applied when no srcset

- **WHEN** an img has no srcset but its `src` matches a known high-res rule (e.g. a Twitter thumbnail URL)
- **THEN** `resolveBestImageSrc` returns `{ src: <upgraded url>, srcsetWidth: null }`

#### Scenario: Falls back to src when no srcset and no platform rule matches

- **WHEN** an img has no srcset and its `src` does not match any high-res rule
- **THEN** `resolveBestImageSrc` returns `{ src: img.src, srcsetWidth: null }`

### Requirement: Picker overlay UI

When the content script receives `{ type: "open-image-picker" }`, it SHALL call `collectPageImages` and:

- If the result is empty, send a `{ type: "toast", variant: "error", title: "Bookleaf", body: "No images found on this page." }` message to the runtime and abort — no overlay is rendered
- Otherwise, render the picker overlay inside the existing shadow DOM

The overlay SHALL consist of:
- A fixed-position, full-viewport backdrop with a semi-transparent dark scrim
- A centered panel with a header, a scrollable image grid, and a footer
- **Header**: label showing total image count (e.g. `"12 images found"`), and a close (`✕`) button
- **Grid**: one thumbnail per collected image; each thumbnail displays the image, its dimensions, and a checkbox; clicking anywhere on a thumbnail toggles its selected state; selected thumbnails are visually distinguished (e.g. checkbox checked, border highlight). Dimensions are derived as: if `srcsetWidth` is known and the image has loaded (`naturalWidth > 0`), display `srcsetWidth × round(srcsetWidth × naturalHeight / naturalWidth)` (the true resolution estimated from the srcset width and the loaded aspect ratio); otherwise display `naturalWidth × naturalHeight`. Dimensions are shown even when `0 × 0` to aid debugging.
- **Footer**: a confirm button labeled `"Save X image"` / `"Save X images"` (count updates as selection changes); the button SHALL be disabled when no thumbnails are selected

Pressing `Escape` or clicking the close button SHALL dismiss the overlay without saving. If another picker overlay is already open when `open-image-picker` is received, the existing one SHALL be replaced.

#### Scenario: No images on page shows error toast instead of overlay

- **WHEN** the content script receives `open-image-picker` and `collectPageImages` returns an empty array
- **THEN** an error toast is shown with body `"No images found on this page."` and no overlay is rendered

#### Scenario: Overlay renders with one thumbnail per collected image

- **WHEN** `collectPageImages` returns N images
- **THEN** the picker overlay contains exactly N thumbnail elements

#### Scenario: Each thumbnail shows intrinsic dimensions

- **WHEN** the picker overlay renders a collected image
- **THEN** each thumbnail displays the image's `naturalWidth × naturalHeight` as a text label (e.g. `"1600 × 900"`); images with unresolved dimensions display `"0 × 0"`

#### Scenario: Clicking a thumbnail toggles its selection

- **WHEN** the user clicks an unselected thumbnail
- **THEN** its checkbox becomes checked and the confirm button count increments

#### Scenario: Clicking a selected thumbnail deselects it

- **WHEN** the user clicks a selected thumbnail
- **THEN** its checkbox becomes unchecked and the confirm button count decrements

#### Scenario: Confirm button is disabled with no selection

- **WHEN** no thumbnails are selected
- **THEN** the confirm button is disabled and cannot be clicked

#### Scenario: Confirm button label reflects selection count

- **WHEN** the user selects 3 thumbnails
- **THEN** the confirm button reads `"Save 3 images"`

#### Scenario: ESC closes overlay without saving

- **WHEN** the picker overlay is open and the user presses `Escape`
- **THEN** the overlay is removed and no save is initiated

#### Scenario: Close button closes overlay without saving

- **WHEN** the user clicks the close button in the overlay header
- **THEN** the overlay is removed and no save is initiated

#### Scenario: Clicking outside the panel closes overlay without saving

- **WHEN** the user clicks on the backdrop area outside the picker panel
- **THEN** the overlay is removed and no save is initiated

#### Scenario: Picker does not open on the Bookleaf app

- **WHEN** the keyboard shortcut is pressed on a page whose origin matches `VITE_APP_URL`
- **THEN** the content script does nothing — no overlay is rendered and no toast is shown

#### Scenario: Opening picker while one is already open replaces it

- **WHEN** a picker overlay is already open and `open-image-picker` is received again
- **THEN** the existing overlay is removed and a fresh one is rendered

### Requirement: Batch save and aggregated feedback

When the user confirms the selection, the content script SHALL close the overlay and send `{ type: "picker-save", images: Array<{ srcUrl: string }> }` to the background. The background SHALL handle this message by invoking `handleSave` for each image in silent mode (suppressing per-image toasts) using `Promise.allSettled`, then send a single aggregated toast to the originating tab:

- All fulfilled: `variant: "success"`, title `"Saved to Bookleaf."`, body `"X images added to Unsorted."`
- Mixed results: `variant: "error"`, title `"Partially saved."`, body `"X saved, Y failed. Check your connection."`
- All rejected: `variant: "error"`, title `"Couldn't save images."`, body `"Check your connection and try again."`

Each `handleSave` call SHALL use the tab's URL as `pageUrl` and the tab's title as `title`, the same defaults used by the drag-save flow.

#### Scenario: All images save successfully

- **WHEN** the user confirms a selection of 4 images and all save without error
- **THEN** a success toast is shown: `"4 images added to Unsorted."`

#### Scenario: Some images fail to save

- **WHEN** the user confirms a selection of 4 images and 1 fails
- **THEN** an error toast is shown: `"3 saved, 1 failed. Check your connection."`

#### Scenario: All images fail to save

- **WHEN** the user confirms a selection and every save call rejects
- **THEN** an error toast is shown with title `"Couldn't save images."` and body `"Check your connection and try again."`

### Requirement: Popup trigger for image picker

The background service worker SHALL handle a `{ type: "trigger-image-picker" }` runtime message. Upon receiving it, the background SHALL query the active tab in the current window and execute the same dispatch logic as `handleBrowseImagesCommand`: send `{ type: "open-image-picker" }` to the active tab's content script via `browser.tabs.sendMessage`. If the active tab has no injected content script, the send SHALL fail silently.

#### Scenario: Popup trigger opens the image picker

- **WHEN** the background receives `{ type: "trigger-image-picker" }` from the popup
- **THEN** `{ type: "open-image-picker" }` is sent to the active tab's content script
- **AND** the picker overlay appears identically to the keyboard-shortcut-triggered flow

#### Scenario: Popup trigger on a restricted page fails silently

- **WHEN** the background receives `{ type: "trigger-image-picker" }` and the active tab is a restricted page
- **THEN** the `sendMessage` call throws and the error is silently caught with no user-visible effect
