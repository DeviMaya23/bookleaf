# Spec: Extension Drag Drop Save

## Purpose

Defines the drag-and-drop save flow for the "Save to Bookleaf" browser extension: how the content script resolves image source, title, and permalink context at `dragstart`, how the transient drop zone is shown and removed, how Bookleaf's own app is excluded from this flow, and how a `drop` on the zone triggers the existing save pipeline, as part of the extension save flow defined in `extension-save-image`.

## Requirements

### Requirement: Drag-time image source resolution

On `dragstart`, the content script SHALL attempt to resolve a `srcUrl` for the dragged element using, in order: (1) `resolveCardImageSrc(target)` when `shouldResolveCardDom(pageUrl)` matches the current page, (2) `target.src` when `target` is itself an `<img>` element, otherwise (3) the `src` of the first `<img>` descendant found via `target.querySelector("img")`. If none of these resolve a value, `srcUrl` SHALL be `null` for that drag gesture.

#### Scenario: Dragging a card image on a registered card site

- **WHEN** the user starts dragging an element on a page matching `shouldResolveCardDom` (e.g. a Pinterest or Instagram card), and `event.target` is the card's `<a>` wrapper
- **THEN** the content script SHALL resolve `srcUrl` via `resolveCardImageSrc(event.target)`, identical to the existing `contextmenu` capture path

#### Scenario: Dragging on a card site where the card resolver finds nothing

- **WHEN** the user starts dragging on a page matching `shouldResolveCardDom`, but `resolveCardImageSrc(event.target)` returns `null` (e.g. a full-image/lightbox view not wrapped in the site's card structure) and `event.target` is itself an `<img>`
- **THEN** the content script SHALL fall through to resolve `srcUrl` as `event.target.src`, rather than resolving `srcUrl` to `null`

#### Scenario: Dragging a plain image on an unregistered site

- **WHEN** the user starts dragging an `<img>` element on a page not matching `shouldResolveCardDom`
- **THEN** the content script SHALL resolve `srcUrl` as `event.target.src`

#### Scenario: Dragging a non-`<img>` wrapper that contains an image descendant

- **WHEN** the user starts dragging an element that is not itself an `<img>` and is not resolvable via the card-site path (e.g. a `<div draggable="true">` wrapper used by a site's own drag-and-drop UI library), and that element has an `<img>` descendant
- **THEN** the content script SHALL resolve `srcUrl` as the `src` of the first `<img>` descendant found via `target.querySelector("img")`

#### Scenario: Dragging an element with no resolvable image

- **WHEN** the user starts dragging an element that is neither on a registered card site, nor itself an `<img>`, nor has an `<img>` descendant
- **THEN** `srcUrl` SHALL resolve to `null` for that drag gesture

### Requirement: Drag-time title and permalink resolution

On `dragstart`, when `srcUrl` resolves to a non-null value, the content script SHALL also resolve `title` using the existing per-site resolvers (`resolveAltText`, `resolveTweetText`, `resolveFacebookAltText`, matched by the same site predicates already used by the `contextmenu` capture path) and SHALL resolve a candidate permalink via `event.target.closest("a")?.href`.

#### Scenario: Title resolves via existing per-site resolver

- **WHEN** a drag starts on a site matched by `isTwitterUrl`, `isImgurUrl`, `isInstagramUrl`, or `isFacebookUrl`, and the corresponding resolver returns a non-empty string
- **THEN** the content script SHALL capture that string as `title` for this drag gesture

#### Scenario: No site-specific title resolver matches

- **WHEN** a drag starts on a site not matched by any title resolver
- **THEN** `title` SHALL be left unresolved at capture time, falling back to the same default the right-click flow uses (the tab's title) when the save is performed

### Requirement: Drop zone visibility and positioning

The content script SHALL render a transient drop zone, anchored at the pointer's `dragstart` client coordinates, if and only if the drag-to-save setting is enabled (per `getDragEnabled`, default `true`) AND `dragstart` resolved a non-null `srcUrl`. The drag-enabled check SHALL be read fresh from storage on each `dragstart`, not cached at content-script load time, so a setting change takes effect on the very next drag gesture without requiring a page reload. The drop zone SHALL be removed on `dragend`, regardless of whether a `drop` occurred. As a safety net for sites whose own drag-and-drop UI cancels native dragging (e.g. by calling `preventDefault()` on `dragstart`, which prevents `dragend` from ever firing), the content script SHALL also remove the drop zone on `pointerup` and `mouseup`.

#### Scenario: Drop zone appears for a resolvable image drag when drag-to-save is enabled

- **WHEN** drag-to-save is enabled and `dragstart` resolves a non-null `srcUrl`
- **THEN** the content script SHALL render a drop zone near the `dragstart` pointer position

#### Scenario: Drop zone does not appear for an unresolvable drag

- **WHEN** `dragstart` resolves `srcUrl` as `null`
- **THEN** the content script SHALL NOT render a drop zone for that drag gesture

#### Scenario: Drop zone does not appear when drag-to-save is disabled

- **WHEN** drag-to-save is disabled, regardless of whether `dragstart` would otherwise resolve a non-null `srcUrl`
- **THEN** the content script SHALL NOT render a drop zone for that drag gesture

#### Scenario: Drop zone is removed when the drag ends without a drop

- **WHEN** a drag that produced a drop zone ends via `dragend` without a `drop` on the zone (e.g. the user cancels the drag or drops outside the zone)
- **THEN** the drop zone SHALL be removed from the page

#### Scenario: Drop zone is removed via the pointerup/mouseup safety net when dragend never fires

- **WHEN** a drag that produced a drop zone is taken over by the host page's own drag-and-drop library (e.g. dnd-kit, react-dnd) such that native `dragend` never fires
- **THEN** the drop zone SHALL still be removed from the page once `pointerup` or `mouseup` fires

### Requirement: Drag-and-drop save is disabled on Bookleaf's own app

The content script SHALL NOT capture drag context or render a drop zone when the current page's origin matches the Bookleaf web app's origin (`VITE_APP_URL`).

#### Scenario: Dragging an image within the Bookleaf app

- **WHEN** the user drags an element on a page whose origin matches `VITE_APP_URL`
- **THEN** the content script SHALL NOT resolve drag context or render a drop zone for that drag gesture, regardless of whether the dragged element would otherwise resolve a `srcUrl`

### Requirement: Drop triggers save via the existing save pipeline

On `drop` within the rendered drop zone, the content script SHALL send a single message to the background service worker containing the `srcUrl`, resolved `title` (or `undefined`), and resolved permalink (or `undefined`) captured at `dragstart`. The background SHALL resolve the final `pageUrl` by applying the existing `resolveLinkPermalink` check to the captured permalink (falling back to the page URL when it does not match), then SHALL invoke the existing `handleSave` function with the resulting values, identically to the right-click save path.

#### Scenario: Drop on the zone saves the image

- **WHEN** the user drops a drag (that produced a non-null `srcUrl` at `dragstart`) onto the rendered drop zone
- **THEN** the background SHALL call `handleSave` with that `srcUrl`, the resolved or fallback `title`, and the `pageUrl` resolved via `resolveLinkPermalink`, the same way it would for an equivalent right-click save

#### Scenario: Drop zone prevents default browser navigation on drop

- **WHEN** the user drops an image onto the rendered drop zone
- **THEN** the content script SHALL call `preventDefault` on the `dragover` and `drop` events for that zone, so the browser does not navigate to or open the dragged image

### Requirement: Drop zone color scheme respects dark mode preference

The content script SHALL read `getDarkMode()` from storage each time the drop zone is rendered. If dark mode is enabled, the drop zone SHALL render using the dark color palette. If dark mode is disabled or unset, it SHALL use the light color palette. Theming SHALL be applied via the same CSS custom property and `data-theme` mechanism used by the toast and picker overlay.

#### Scenario: Drop zone displays with light palette when dark mode is off

- **WHEN** `getDarkMode()` resolves to `false` and a drag gesture triggers the drop zone
- **THEN** the drop zone renders with the light color palette

#### Scenario: Drop zone displays with dark palette when dark mode is on

- **WHEN** `getDarkMode()` resolves to `true` and a drag gesture triggers the drop zone
- **THEN** the drop zone renders with the dark color palette
