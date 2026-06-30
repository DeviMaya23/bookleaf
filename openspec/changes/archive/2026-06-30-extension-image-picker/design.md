## Context

The content script already has a shadow DOM host (`#bookleaf-toast-host`) used to inject isolated UI: a toast, a drag-and-drop drop zone, and a snip selection overlay. All are hand-rolled DOM with a single shared `<style>` block. Colors are currently hardcoded as hex literals, and overlays are light-mode only.

The background script handles all save operations (`handleSave`, `handleCapture`) and keyboard commands (`browser.commands.onCommand`). The `snip-capture` command pattern — manifest declares it, background listens and sends a message to the active tab's content script, content script renders the UI — is the established model this change follows.

## Goals / Non-Goals

**Goals:**
- Allow users to browse and batch-save `<img>` elements on the current page via a keyboard-triggered overlay
- Resolve the best available URL for each image (srcset → src → high-res rules)
- Sort images largest-to-smallest to surface the most useful images first
- Unify overlay color tokens as CSS custom properties; apply dark mode palette to toast, drop zone, and picker based on stored preference

**Non-Goals:**
- Images inside shadow DOM roots (not reachable via `querySelectorAll`)
- CSS background images (`background-image: url(...)`)
- `srcset` with pixel-density descriptors (`x`-descriptors) — fall through to `src`
- Filtering images by minimum size (deferred to a future change)
- Persisting picker selection state across page navigations

## Decisions

### 1. Pure functions in `imagePicker.ts`, not inline in `content/index.ts`

`content/index.ts` is already large and mixes DOM logic with event listeners. The srcset parsing and image collection logic are pure, stateless functions with clear inputs and outputs — ideal candidates for a dedicated lib file and unit tests. This follows the existing pattern (`highResRules.ts`, `dragSaveContext.ts`, etc.).

**Alternative considered:** inline everything in `content/index.ts`. Rejected because it makes the file harder to navigate and the logic untestable in isolation.

**`imagePicker.ts` exports:**
```
parseSrcset(srcset: string): string | null
  — parses `w`-descriptor srcset, returns URL of widest descriptor; null if no valid w-descriptors

resolveBestImageSrc(img: HTMLImageElement): string | null  
  — resolves best URL: parseSrcset(img.srcset) → resolveHighResUrl(img.src) → img.src
  — returns null if src is empty or a blob: URI

collectPageImages(): PageImage[]
  — querySelectorAll('img'), apply filters, resolveBestImageSrc, sort by pixel area desc

type PageImage = { src: string; naturalWidth: number; naturalHeight: number }
```

### 2. srcset parsing: `w`-descriptors only, pick widest

Width descriptors (`400w`, `1600w`) are the standard for resolution switching and directly encode pixel width. Parsing them to find the widest is deterministic and requires no knowledge of the current viewport.

Pixel-density descriptors (`1x`, `2x`) are skipped — falling through to `src` is acceptable since the src is typically the 1x version and the high-res rules pipeline may still upgrade it for known platforms.

**Alternative considered:** using `img.currentSrc`. Rejected because it reflects what the browser chose for the current viewport/DPR, not the highest-resolution option available.

### 3. Image collection filters

Two filters only (per design decision):
- Skip if `src` is empty or a `blob:` URI
- Skip if `img.complete === false && img.naturalWidth === 0` (lazy-loaded, not yet decoded)

No minimum size filter. Shadow DOM images are unreachable and silently omitted.

### 4. Overlay trigger: keyboard command → background → content message

Same pattern as `snip-capture`:
1. Manifest declares `browse-images` command (`Alt+Shift+I`)
2. Background `handleBrowseImagesCommand` sends `{ type: "open-image-picker" }` to the active tab
3. Content script receives the message, collects images, renders the picker overlay

**Alternative considered:** triggering directly from popup button. Rejected for v1 — the popup closes itself when a tab loses focus, which would dismiss the overlay before the user can interact with it.

### 5. Picker UI: hand-rolled DOM in shadow DOM, no React

The content script has no React. Bringing React in would meaningfully increase the injected bundle size (popup and content script are separate Vite entry points). The picker's state is simple enough to track with plain JS: a `Set<string>` of selected `src` values, a counter display, and a confirm button.

### 6. Multi-save message and aggregated toast

Content script sends `{ type: "picker-save", images: Array<{ srcUrl: string }> }` to the background. Background fans out `handleSave()` calls with `Promise.allSettled`, then sends a single aggregated toast:

- All succeeded: `"Saved X images to Bookleaf."`
- Partial failure: `"X saved, Y failed. Check your connection."`
- All failed: `"Couldn't save images. Check your connection."`

**Alternative considered:** having the content script send one `drag-save`-style message per image. Rejected because it would produce one toast per image, which is noisy for multi-select saves, and the background would have no way to aggregate the outcome.

### 7. Dark mode theming via CSS custom properties

The shadow DOM `<style>` block is expanded to define two sets of CSS custom properties (`--bl-bg`, `--bl-text`, `--bl-border`, etc.) applied via `[data-theme="dark"]` on the shadow host. When any overlay opens, `getDarkMode()` is read from storage and the `data-theme` attribute is set accordingly on the shadow host element.

All existing `.toast` and `.drop-zone` rules are updated to use these variables instead of hardcoded hex values. The picker overlay uses the same variables.

**Alternative considered:** reading dark mode preference once at content script init time. Rejected because the user may toggle dark mode in the popup while the tab is open; reading at overlay-open time ensures the overlay reflects the current preference without requiring a page reload.

### 8. Settings page: second shortcut row for `browse-images`

The existing Settings component fetches the `snip-capture` command with a single `useState`. We add a second state variable for the `browse-images` shortcut and fetch both in the same `browser.commands.getAll()` call. A new row is rendered below the existing snip hotkey row, same visual style, same "Change" button behavior.

## Risks / Trade-offs

- **Large page image counts** → Picker shows all images with no cap; on image-heavy pages (e-commerce galleries) this could render hundreds of thumbnails. The shadow DOM grid is scrollable, so it remains functional, but performance is untested at scale. A future filter by minimum size would mitigate this.
- **Cross-origin image display in picker** → Thumbnail images in the picker use the page's own `<img>` src directly (no fetch at pick time), so they display without CORS issues. The actual save fetch happens in the background script, which has `host_permissions: <all_urls>` — same as today.
- **srcset parsing edge cases** → Malformed srcset values are ignored per-descriptor; the parser falls through to `src` if no valid `w`-descriptor is found. No error is surfaced to the user.
- **getDarkMode() is async, called at overlay open** → There will be a negligible delay between the keyboard shortcut firing and the overlay appearing (one storage read). Acceptable given the alternative is stale theme data.

## Open Questions

None — all decisions resolved during exploration.
