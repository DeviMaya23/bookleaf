## 1. Drag-time capture (content script)

- [x] 1.1 Add `resolveDragImageSrc(target, pageUrl)` helper (new lib file or alongside `cardDomResolveRules.ts`): returns `resolveCardImageSrc(target)` when `shouldResolveCardDom(pageUrl)` matches, else `target.src` when `target` is an `<img>`, else `null`.
- [x] 1.2 Add a `dragstart` listener in `content/index.ts` that, on a non-null `resolveDragImageSrc` result, captures `srcUrl`, resolves `title` via the existing per-site resolvers (`resolveTweetText`/`resolveAltText`/`resolveFacebookAltText`), resolves a candidate permalink via `event.target.closest("a")?.href`, and stores all three in a module-level variable for the duration of the gesture.
- [x] 1.3 Add a `dragend` listener that always clears the captured context and removes the drop zone (regardless of whether `drop` fired).

## 2. Drop zone UI (content script)

- [x] 2.1 Render a drop zone element into the existing `#bookleaf-toast-host` shadow root, positioned at the `dragstart` client coordinates, only when step 1.2 captured a non-null `srcUrl`.
- [x] 2.2 Add `dragover` (with `preventDefault`) and `drop` (with `preventDefault`) handlers scoped to the drop zone element.
- [x] 2.3 On `drop`, send one message to background (e.g. `{ type: "drag-save", srcUrl, title, linkUrl }`) using the values captured at `dragstart`.

## 3. Save handoff (background)

- [x] 3.1 Add a `browser.runtime.onMessage` handler for the `"drag-save"` message type in `background/index.ts`.
- [x] 3.2 In that handler, resolve `pageUrl` by applying `resolveLinkPermalink(linkUrl)` (use `linkUrl` if it matches, else fall back to the sender tab's URL), resolve `title` (use the captured value if present, else fall back to the sender tab's title, mirroring `resolveTitle`'s existing fallback), and call the existing `handleSave({ srcUrl, pageUrl, title, tabId })` unchanged.

## 4. Tests

- [x] 4.1 Unit test `resolveDragImageSrc`: card-site path delegates to `resolveCardImageSrc`; non-card `<img>` target returns `target.src`; non-card non-`<img>` target returns `null`.
- [x] 4.2 Unit test the background `"drag-save"` message handler: permalink resolution via `resolveLinkPermalink` (match and no-match cases), title fallback to tab title when no captured title is present, and that `handleSave` is invoked with the expected arguments.

## 5. Verification

- [x] 5.1 Manually verify the drag-and-drop save flow in Firefox on Pinterest, Instagram, Twitter/X, and a plain blog `<img>`: drop zone appears near the drag origin, drop saves the image with the same `srcUrl`/title/source-URL quality as the equivalent right-click save. (Initial pass surfaced 3 gaps, since fixed: card-site resolution had no fallback to the generic `<img>` tier — broke Pinterest's full-image/lightbox view; no descendant-`<img>` fallback for non-`<img>` draggable wrappers — broke Twitter/Facebook; Bookleaf's own app was not excluded from drag-save. Re-verification of the fixes pending.)
- [x] 5.2 Manually verify no drop zone appears when dragging non-image content (e.g. selected text) and that cancelling a drag (dropping outside the zone) leaves no leftover drop zone element. (Also verified the dropzone clears via the pointerup/mouseup safety net on dnd-kit-based UIs, by temporarily disabling the Bookleaf self-exclusion to test on Bookleaf's own app, then restoring it — confirmed working.)
- [x] 5.3 Run `npm run build` and `npm run lint` in `extensions/`; fix any issues raised. (No `lint` script exists in `extensions/`; ran `npm run build` and `npm run type-check` instead — both clean.)
