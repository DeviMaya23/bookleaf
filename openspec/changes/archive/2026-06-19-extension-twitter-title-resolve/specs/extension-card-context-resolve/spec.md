## MODIFIED Requirements

### Requirement: Card-level DOM resolution via content script

For sites where the right-clicked element does not directly carry image data (the right-click target is not an `<img>`), the extension SHALL resolve card-level context via a content script. The content script SHALL listen for the native `contextmenu` event in capture phase. When the event target's enclosing card matches a registered DOM-resolution site rule, the content script SHALL walk the DOM from the event target to locate the card's image (e.g. `closest('a')` then `querySelector('img')` within it) and, if found, send the resolved context to the background service worker via `runtime.sendMessage` before any subsequent context-menu click is processed.

Independently of `srcUrl` resolution, the content script SHALL also attempt to resolve `title` for sites with a registered title-resolution rule, regardless of whether `srcUrl` was found for that right-click. The content script SHALL send a `resolved` message if at least one field (`srcUrl` or `title`) was resolved; it SHALL NOT require both fields to be present.

The resolved context payload SHALL be typed as a partial bag of card-level fields (`Partial<{ srcUrl: string; title: string }>`), not a single-purpose value, so that future resolvers can populate additional fields without changing the message contract.

The background service worker SHALL store the most recently resolved context per `tabId`, to be read by the context-menu click handler per the Authenticated save flow requirement in `extension-save-image`.

#### Scenario: Pinterest card image is resolved on right-click

- **WHEN** the user right-clicks a Pinterest grid card whose click target is not an `<img>`, and the card's link wrapper contains a descendant `<img>`
- **THEN** the content script sends a resolved context containing that `<img>`'s `src` as `srcUrl` to the background, before the user selects a context menu item

#### Scenario: No image found in the card is not resolved

- **WHEN** the user right-clicks an element matching a registered DOM-resolution site rule, but no `<img>` descendant is found via the DOM walk, and no title-resolution rule also matches
- **THEN** no resolved context is sent for that right-click
- **AND** any subsequent "Save to Bookleaf" click for that tab SHALL be treated as having no resolved `srcUrl`

#### Scenario: Twitter tweet text is resolved on right-click independently of srcUrl

- **WHEN** the user right-clicks an `<img>` inside a tweet on a Twitter/X page (the click target is the `<img>` itself, native `srcUrl` is already available via the browser's context menu)
- **THEN** the content script climbs ancestors from the click target and, on finding the nearest ancestor containing a `[data-testid="tweetText"]` descendant, sends a resolved context containing that element's text content as `title`
- **AND** this happens regardless of whether a `srcUrl` was also resolved by the DOM walk

#### Scenario: No tweetText found sends no title

- **WHEN** the user right-clicks an `<img>` on a Twitter/X page and no ancestor within the climb-depth limit contains a `[data-testid="tweetText"]` descendant (e.g. an image-only tweet)
- **THEN** no `title` field is sent in the resolved context for that right-click

### Requirement: Resolved context store is per-tab and most-recent-wins

The background service worker SHALL key stored resolved contexts by `tabId` and SHALL overwrite any previously stored context for that `tabId` when a new one is received. The stored context for a `tabId` SHALL be treated as belonging to the most recent right-click in that tab.

#### Scenario: A later right-click's resolution replaces an earlier one

- **WHEN** the content script resolves context for one right-click in a tab, and then a second right-click in the same tab resolves a different context before "Save to Bookleaf" is clicked
- **THEN** the click handler uses the second (most recent) resolved context, not the first

## ADDED Requirements

### Requirement: Tweet text DOM-resolution rule

The extension SHALL register a Twitter/X tweet-text resolution rule that, given a right-click target element, climbs ancestors (bounded to a maximum depth) and at each ancestor checks for a descendant matching `[data-testid="tweetText"]`, returning the text content of the first match found (i.e. the match nearest to the click target). The rule SHALL NOT distinguish between a tweet's own text, a quote-tweet's outer text, or an embedded quoted tweet's text — proximity to the click target alone determines the result.

#### Scenario: Image inside a quoted tweet resolves to the quoted tweet's own text

- **WHEN** the user right-clicks an image inside an embedded quoted-tweet card, where the quoted tweet's own `tweetText` is a closer ancestor-descendant than the outer quoting tweet's `tweetText`
- **THEN** the rule resolves to the quoted tweet's text, not the outer quoting tweet's text

#### Scenario: Climb depth limit is respected

- **WHEN** no ancestor within the bounded climb depth contains a `[data-testid="tweetText"]` descendant
- **THEN** the rule resolves no text, even if a matching element exists further up the DOM tree
