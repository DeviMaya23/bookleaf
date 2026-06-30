## 1. Manifest: Add browse-images command

- [x] 1.1 Add `browse-images` command to `manifest.json` with suggested key `Alt+Shift+I` and description `"Browse images on this page"`

## 2. Image picker lib module

- [x] 2.1 Create `extensions/src/lib/imagePicker.ts` and export `parseSrcset(srcset: string): string | null` — parses `w`-descriptor entries, returns URL of widest; returns `null` if none found
- [x] 2.2 Export `resolveBestImageSrc(img: HTMLImageElement): string | null` — priority: `parseSrcset(img.srcset)` → `resolveHighResUrl(img.src)` → `img.src`; returns `null` if result is empty or `blob:` URI
- [x] 2.3 Export `collectPageImages(document: Document): PageImage[]` — queries all `img`, applies exclusion filters, resolves best src, sorts descending by `naturalWidth * naturalHeight`; export the `PageImage` type
- [x] 2.4 Write unit tests for `parseSrcset`: multiple w-descriptors returns widest, single entry, no w-descriptors returns null, empty string returns null
- [x] 2.5 Write unit tests for `resolveBestImageSrc`: srcset widest takes priority, falls through to high-res rule, falls through to src, returns null for empty src
- [x] 2.6 Write unit tests for `collectPageImages`: sorted largest first, lazy-loaded excluded, empty-src excluded, blob-src excluded, empty page returns empty array

## 3. Overlay color theming

- [x] 3.1 Replace all hardcoded hex values in the shadow DOM `<style>` block in `content/index.ts` with CSS custom properties (`--bl-bg`, `--bl-text`, `--bl-text-sec`, `--bl-border`, `--bl-accent`, `--bl-surface`, `--bl-success`, `--bl-error`) — define both light and dark values via `[data-theme="light"]` and `[data-theme="dark"]` selectors on the shadow host
- [x] 3.2 Add a helper `applyTheme(host: HTMLElement): Promise<void>` in `content/index.ts` that reads `getDarkMode()` and sets `data-theme` on the shadow host accordingly
- [x] 3.3 Call `applyTheme` before showing the toast (wrap `showToast` call sites to apply theme first)
- [x] 3.4 Call `applyTheme` before rendering the drop zone in `renderDropZone`

## 4. Background: browse-images command handler and picker-save message

- [x] 4.1 Add `handleBrowseImagesCommand` in `background/index.ts`: query active tab, send `{ type: "open-image-picker" }` via `browser.tabs.sendMessage`, catch and silently swallow errors (restricted pages)
- [x] 4.2 Wire `handleBrowseImagesCommand` into the existing `browser.commands.onCommand` listener alongside `snip-capture`
- [x] 4.3 Add `handlePickerSaveMessage` in `background/index.ts`: accepts `{ type: "picker-save", images: Array<{ srcUrl: string }> }` and sender tab; calls `handleSave` for each image with `Promise.allSettled`; sends a single aggregated toast (all success / mixed / all failed) to the tab
- [x] 4.4 Wire `handlePickerSaveMessage` into `browser.runtime.onMessage`

## 5. Content script: picker overlay

- [x] 5.1 Add picker overlay CSS to the shadow DOM `<style>` block: backdrop, centered panel, header, scrollable grid, thumbnail card (with checkbox state styles), footer, confirm button, disabled state
- [x] 5.2 Add the `open-image-picker` message type to the `browser.runtime.onMessage` listener in `content/index.ts`; on receipt, call `collectPageImages(document)` — if empty, show error toast `"No images found on this page."` and return
- [x] 5.3 Implement `renderPickerOverlay(images: PageImage[]): void`: build and append the panel to the shadow DOM; track selected state as a `Set<string>` of srcUrls; wire thumbnail click to toggle selection and update the confirm button label and disabled state; wire close button and ESC key to remove the overlay
- [x] 5.4 Implement `removePickerOverlay()`: remove the overlay element and the ESC keydown listener; guard against re-entry
- [x] 5.5 On confirm button click: close overlay, send `{ type: "picker-save", images: [...selected].map(srcUrl => ({ srcUrl })) }` to the background via `browser.runtime.sendMessage`
- [x] 5.6 Call `applyTheme` before rendering the picker overlay

## 6. Popup Settings: browse-images shortcut row

- [x] 6.1 In `Settings.tsx`, add a second `browseShortcut` state variable; fetch the `browse-images` command in the existing `browser.commands.getAll()` call alongside `snip-capture`
- [x] 6.2 Render a `"Browse images hotkey"` row below the snip hotkey row, same visual pattern; clicking the button calls `handleChangeClick` (same handler — opens browser shortcut settings)

## 7. Dimension display in picker thumbnails

- [x] 7.1 Update `collectPageImages` in `imagePicker.ts` to include `naturalWidth` and `naturalHeight` on the returned `PageImage` objects (they are already present in the type — confirm they are populated from the `img` element at collection time)
- [x] 7.2 Update `renderPickerOverlay` in `content/index.ts` to render a dimension label (`"W × H"`) below each thumbnail image; show `"0 × 0"` when either dimension is zero

## 8. Build and lint

- [x] 8.1 Run `npm run build` in `extensions/` and fix any type errors or build failures
- [x] 8.2 Run `npm run lint` in `extensions/` and fix any lint warnings or errors
