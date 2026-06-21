# extension-snip-capture

## Purpose

TBD

## Requirements

### Requirement: Hotkey command registration

The extension SHALL register a keyboard command (e.g. "Snip to Bookleaf") via the manifest `commands` key. The command SHALL be remappable through the browser's native shortcut settings (`chrome://extensions/shortcuts` on Chrome, the shortcut editor under `about:addons` on Firefox). The extension SHALL NOT provide its own remapping UI in this change.

#### Scenario: Default hotkey triggers the snip flow

- **WHEN** the user presses the extension's registered default keyboard shortcut while a tab is focused
- **THEN** the snip capture flow begins for the active tab (per the Capture trigger requirement)

#### Scenario: User-remapped hotkey triggers the snip flow

- **WHEN** the user has remapped the command to a different key combination via the browser's native shortcut settings, and presses that combination
- **THEN** the snip capture flow begins identically to the default-hotkey case

### Requirement: Capture trigger and frozen overlay

When the hotkey command fires, the background service worker SHALL capture the active tab's currently visible viewport via `captureVisibleTab()` and send the resulting image to the active tab's content script, which SHALL render a full-viewport overlay showing that captured image as a frozen background, dimmed except where a selection rectangle is being drawn. The overlay SHALL only render in the top frame of the page (the content script is not injected into iframes).

#### Scenario: Hotkey on a normal webpage shows the frozen overlay

- **WHEN** the user presses the hotkey while viewing a regular webpage
- **THEN** the visible viewport is captured
- **AND** a dimmed overlay appears showing that captured frame, with a crosshair cursor available for selection

#### Scenario: Hotkey on a page without an injected content script does nothing visible

- **WHEN** the user presses the hotkey on a page where the content script cannot run (e.g. a browser-internal page)
- **THEN** no overlay appears and no capture is attempted, consistent with other content-script-dependent features on such pages

### Requirement: Drag-to-select capture region

While the overlay is shown, the user SHALL be able to draw exactly one selection rectangle by pressing the mouse button, dragging, and releasing it over the frozen frame. The dimmed area SHALL be cut out to reveal the underlying frozen frame's real pixels within the current selection rectangle as it is drawn. No resize handles SHALL be shown after the rectangle is drawn, and no preview or confirmation step SHALL be presented — releasing the mouse button immediately finalizes the selection.

#### Scenario: Dragging draws a live selection rectangle

- **WHEN** the user presses the mouse button on the frozen overlay and drags
- **THEN** a selection rectangle grows to follow the cursor
- **AND** the area inside the rectangle shows the frozen frame's pixels un-dimmed

#### Scenario: Releasing the mouse finalizes the selection with no further interaction

- **WHEN** the user releases the mouse button after dragging a selection rectangle
- **THEN** the selection is immediately finalized at its current bounds
- **AND** no resize handles, preview, or confirmation dialog is shown

### Requirement: Crop and auto-save on mouseup

On mouseup, the extension SHALL crop the frozen captured image to the finalized selection rectangle's bounds using a canvas, producing an image blob, and SHALL immediately invoke the capture save flow (`handleCapture`) with that blob — no user action beyond the mouseup is required to save.

#### Scenario: Mouseup triggers an immediate save

- **WHEN** the user finishes a drag-and-release selection over the frozen overlay
- **THEN** the selected region is cropped from the captured frame
- **AND** the cropped blob is passed to the capture save flow without further user interaction

### Requirement: Cancel via Escape

While the overlay is shown (whether before a selection is started or while one is being dragged), pressing Escape SHALL dismiss the overlay immediately with no save attempted and no blob produced.

#### Scenario: Escape before starting a selection cancels the overlay

- **WHEN** the overlay is shown and the user presses Escape before pressing the mouse button
- **THEN** the overlay is dismissed
- **AND** no capture or save occurs

#### Scenario: Escape while dragging a selection cancels the overlay

- **WHEN** the user is mid-drag on a selection rectangle and presses Escape
- **THEN** the overlay is dismissed immediately
- **AND** no crop or save occurs, regardless of the in-progress rectangle's bounds

### Requirement: Capture save flow uses tab metadata, no per-site resolution

The capture save flow SHALL set `pageUrl` to the active tab's `url` and `title` to the active tab's `title`, unconditionally. No per-site alt-text, permalink, or card-DOM resolution (as used by the right-click and drag-drop flows) SHALL apply to a snip capture, since a snip has no referenced DOM element to resolve metadata from.

#### Scenario: Snip on any page uses the tab's own URL and title

- **WHEN** a snip is captured and saved on any webpage, regardless of site
- **THEN** the saved image's `source_url` is set to the active tab's `url`
- **AND** the saved image's title is set to the active tab's `title`

### Requirement: Capture save flow shares persistence with existing save flows

The capture save flow (`handleCapture`) SHALL perform the same authenticated persistence steps as the existing right-click and drag-drop flows — auth/token validation, thumbnail generation, image and thumbnail upload, in-page toast notification, and recent-save bookkeeping — via the same shared persistence logic used by those flows, taking an already-available image blob directly rather than fetching one from a `srcUrl`.

#### Scenario: Successful snip save shows the same success toast as other flows

- **WHEN** a snip is captured while the user is authenticated and the upload succeeds
- **THEN** an in-page toast with title "Saved to Bookleaf." and body "Added to Unsorted." is shown
- **AND** the saved image appears in recent saves

#### Scenario: Unauthenticated snip save is rejected

- **WHEN** a snip is captured and no valid token exists in storage
- **THEN** no upload is attempted
- **AND** an in-page toast with title "Bookleaf" and body "Please log in first." is shown

#### Scenario: Snip save failure shows the same error toast as other flows

- **WHEN** a snip is captured and any step of the upload sequence fails
- **THEN** an in-page toast with title "Couldn't save image." and body "Check your connection and try again." is shown
