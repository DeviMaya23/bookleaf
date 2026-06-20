## MODIFIED Requirements

### Requirement: Card-level DOM resolution via content script

For sites where the right-clicked element does not directly carry image data (the right-click target is not an `<img>`), the extension SHALL resolve card-level context via a content script. The content script SHALL listen for the native `contextmenu` event in capture phase. When the event target's enclosing card matches a registered DOM-resolution site rule, the content script SHALL walk the DOM from the event target to locate the card's image (e.g. `closest('a')` then `querySelector('img')` within it) and, if found, send the resolved context to the background service worker via `runtime.sendMessage` before any subsequent context-menu click is processed.

The resolved context payload SHALL be typed as a partial bag of card-level fields (`Partial<{ srcUrl: string; title: string }>`), not a single-purpose value, so that future resolvers can populate additional fields without changing the message contract. This change SHALL only populate `srcUrl`.

The background service worker SHALL store the most recently resolved context per `tabId`, to be read by the context-menu click handler per the Authenticated save flow requirement in `extension-save-image`.

#### Scenario: Pinterest card image is resolved on right-click

- **WHEN** the user right-clicks a Pinterest grid card whose click target is not an `<img>`, and the card's link wrapper contains a descendant `<img>`
- **THEN** the content script sends a resolved context containing that `<img>`'s `src` as `srcUrl` to the background, before the user selects a context menu item

#### Scenario: Instagram grid card image is resolved on right-click

- **WHEN** the user right-clicks an Instagram profile/grid post thumbnail whose click target is the enclosing `<a>` post link rather than the `<img>`, and that link wraps a descendant `<img>`
- **THEN** the content script sends a resolved context containing that `<img>`'s `src` as `srcUrl` to the background, before the user selects a context menu item

#### Scenario: No image found in the card is not resolved

- **WHEN** the user right-clicks an element matching a registered DOM-resolution site rule, but no `<img>` descendant is found via the DOM walk
- **THEN** no resolved context is sent for that right-click
- **AND** any subsequent "Save to Bookleaf" click for that tab SHALL be treated as having no resolved `srcUrl`
