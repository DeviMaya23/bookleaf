## Why

Some sites (e.g. Squarespace-hosted pages) suppress right-click context menus and `dragstart` events, making it impossible to save images through the existing right-click or drag-and-drop flows. An image picker overlay that scans the page DOM sidesteps these restrictions entirely, and as a side benefit enables saving multiple images at once — something no current flow supports.

## What Changes

- Add a new keyboard shortcut `Alt+Shift+I` (`browse-images` command) that opens an image picker overlay injected into the current tab via the content script
- The overlay collects all `<img>` elements from the page DOM, resolves the best available URL (via `srcset` parsing and existing high-res rules), sorts images largest-to-smallest by pixel area, and presents a scrollable grid with checkboxes
- User selects one or more images and confirms; each selected image is saved via the existing `handleSave` pipeline; a single aggregated toast reports the outcome ("Saved X images" / "X saved, Y failed")
- The `browse-images` shortcut is surfaced in the popup Settings page alongside the existing `snip-capture` shortcut
- Extract color tokens from hard-coded hex values in the shadow DOM `<style>` block into CSS custom properties; apply dark mode palette when the user has dark mode enabled — this unifies theming across the toast, drop zone, and new picker overlays

## Capabilities

### New Capabilities

- `extension-image-picker`: keyboard-triggered overlay that collects, displays, and batch-saves `<img>` elements from the current page

### Modified Capabilities

- `extension-in-page-toast`: toast now reads dark mode preference and applies the dark color palette
- `extension-drag-drop-save`: drop zone now reads dark mode preference and applies the dark color palette
- `extension-popup-settings`: settings page surfaces the new `browse-images` keyboard shortcut

## Impact

- `extensions/manifest.json` — new `browse-images` command entry
- `extensions/src/content/index.ts` — new picker overlay (DOM, styles, state), color token CSS variables, dark mode application for toast and drop zone
- `extensions/src/background/index.ts` — new `handleBrowseImagesCommand` handler wired to `browser.commands.onCommand`; new message type `picker-save` with an array of `srcUrl` values
- `extensions/src/popup/Settings.tsx` — render the `browse-images` shortcut row
- `extensions/src/lib/` — new `imagePicker.ts` module for pure functions: srcset parsing, image collection, sorting (unit-testable in isolation)
- No backend changes; no new dependencies
