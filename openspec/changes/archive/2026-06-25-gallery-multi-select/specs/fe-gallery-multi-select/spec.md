## ADDED Requirements

### Requirement: Select mode toggle in the gallery toolbar

The system SHALL render a select-mode toggle control in `GalleryToolbar`, placed beside the existing Filters button. The control SHALL only be available on fine-pointer devices — it SHALL NOT be rendered on coarse-pointer devices (`useIsCoarsePointer()` is true). The control SHALL NOT be rendered while viewing the trash (`view.type === 'trash'`), the same scoping already used for the Filters button — neither bulk action (add-to-folder, move-to-trash) is meaningful for already-trashed images. Bulk operations for the trash view (e.g. mass restore) are out of scope for this change.

#### Scenario: Toggle is visible on fine-pointer devices outside trash

- **WHEN** a fine-pointer user views the gallery toolbar in a folder, unsorted, or all-images view
- **THEN** a select-mode toggle is visible beside the Filters button

#### Scenario: Toggle is absent on coarse-pointer devices

- **WHEN** a coarse-pointer user views the gallery toolbar
- **THEN** no select-mode toggle is rendered

#### Scenario: Toggle is absent in the trash view

- **WHEN** a fine-pointer user views the trash
- **THEN** no select-mode toggle is rendered
- **AND** select mode cannot be entered while viewing the trash

### Requirement: Filter and sort controls are disabled while in select mode

The system SHALL disable (visually greyed, non-interactive) the sort dropdown trigger, the Filters dropdown trigger, and the name search input in `GalleryToolbar` whenever select mode is active — all three narrow or reorder what's shown in the gallery, so all three are frozen together. Their current values SHALL remain visible but unchangeable until select mode is exited.

#### Scenario: Sort, filter, and search controls are disabled in select mode

- **WHEN** select mode is active
- **THEN** the sort dropdown trigger, the Filters dropdown trigger, and the name search input are all disabled

#### Scenario: Sort, filter, and search controls are re-enabled after exiting select mode

- **WHEN** select mode is turned off
- **THEN** the sort dropdown trigger, the Filters dropdown trigger, and the name search input return to their normal enabled state

### Requirement: Per-card context menu is disabled while in select mode

The system SHALL suppress the right-click context menu on image cards while select mode is active. Right-clicking a card SHALL have no effect in this mode.

#### Scenario: Right-click does nothing in select mode

- **WHEN** select mode is active
- **AND** the user right-clicks an image card
- **THEN** no context menu appears

#### Scenario: Right-click context menu returns after exiting select mode

- **WHEN** select mode is turned off
- **AND** the user right-clicks an image card
- **THEN** the normal context menu (Delete, or Restore/Delete permanently in trash) appears as before

### Requirement: Clicking an image card toggles its selection in select mode

The system SHALL, while select mode is active, toggle an image card's membership in `selectedIds` on a plain click — adding it if absent, removing it if present — instead of opening the viewer or right panel. The clicked image SHALL become the new selection anchor (`mainSelectedId`) on every plain click, regardless of whether the click added or removed it from the selection.

#### Scenario: Plain click adds an unselected image to the selection

- **WHEN** select mode is active
- **AND** the user clicks an image card that is not currently selected
- **THEN** that image is added to `selectedIds`
- **AND** that image becomes the new selection anchor

#### Scenario: Plain click removes a selected image from the selection

- **WHEN** select mode is active
- **AND** the user clicks an image card that is currently selected
- **THEN** that image is removed from `selectedIds`
- **AND** that image still becomes the new selection anchor

#### Scenario: Clicking a card in select mode does not open the viewer or image panel

- **WHEN** select mode is active
- **AND** the user clicks any image card
- **THEN** neither the lightbox/viewer nor the `image`-mode right panel opens

### Requirement: Shift-click replaces the selection with a range from the anchor

The system SHALL, while select mode is active, replace the entire contents of `selectedIds` with the contiguous range of images between the current selection anchor (`mainSelectedId`) and the shift-clicked image (inclusive), computed over the gallery's current display order. The anchor SHALL NOT change as a result of a shift-click — it only changes on a plain click. If no anchor exists yet (`mainSelectedId` is null), a shift-click SHALL behave exactly like a plain click on that image.

#### Scenario: Shift-click selects the range between the anchor and the clicked image

- **WHEN** select mode is active
- **AND** the user has previously plain-clicked image A (now the anchor)
- **AND** the user shift-clicks image D, which appears after A with images B and C between them in display order
- **THEN** `selectedIds` becomes exactly `{A, B, C, D}`
- **AND** the anchor remains image A

#### Scenario: A second shift-click recomputes the range from the same anchor

- **WHEN** select mode is active and the anchor is image A
- **AND** the user shift-clicks image D, then shift-clicks image B
- **THEN** after the second shift-click, `selectedIds` becomes exactly the range between A and B
- **AND** the anchor remains image A throughout both shift-clicks

#### Scenario: Shift-click with no prior anchor behaves like a plain click

- **WHEN** select mode is active
- **AND** no image has been clicked yet (`mainSelectedId` is null)
- **AND** the user shift-clicks an image
- **THEN** `selectedIds` becomes a set containing only that image
- **AND** that image becomes the new anchor

### Requirement: No modifier-key (ctrl/cmd) selection in this iteration

The system SHALL NOT support ctrl/cmd-click as a selection modifier. Only plain click (toggle) and shift-click (range) are supported selection gestures.

#### Scenario: Ctrl/cmd-click has no special selection behavior

- **WHEN** select mode is active
- **AND** the user ctrl/cmd-clicks an image card
- **THEN** the click is treated as a plain click (toggle), with no distinct ctrl/cmd behavior

### Requirement: Selected cards show a distinct visual indicator

The system SHALL render a visually distinct indicator (e.g. a border/ring) on any image card whose ID is present in `selectedIds`. This indicator SHALL be visually distinguishable from the existing drag drop-target indicator used outside select mode.

#### Scenario: Selected card shows the selection indicator

- **WHEN** an image is present in `selectedIds`
- **THEN** its card renders the selection indicator

#### Scenario: Unselected card shows no selection indicator

- **WHEN** an image is not present in `selectedIds`
- **THEN** its card does not render the selection indicator

### Requirement: Navigating to a different view exits select mode entirely; toggling off only clears the selection

The system SHALL clear `selectedIds` and the selection anchor (`mainSelectedId`) whenever select mode is turned off. The system SHALL additionally turn `selectMode` itself off whenever the active view/folder changes (the same trigger that already clears `selectedImage`/`viewerImage`/`lightboxImage` on navigation) — navigation does not leave the user sitting in an empty select mode in the new view.

#### Scenario: Turning off select mode clears the selection

- **WHEN** select mode is active with one or more images selected
- **AND** the user turns select mode off
- **THEN** `selectedIds` and the anchor are both cleared
- **AND** select mode remains off

#### Scenario: Navigating to a different folder exits select mode entirely

- **WHEN** select mode is active with one or more images selected
- **AND** the user navigates to a different folder or view
- **THEN** `selectedIds` and the anchor are both cleared
- **AND** `selectMode` itself is turned off
- **AND** the user must re-enable select mode via the toolbar toggle to select images in the new view

### Requirement: Marquee selection is out of scope

The system SHALL NOT implement click-and-drag rectangle (marquee) selection in this change. Only click and shift-click selection gestures are supported.

#### Scenario: Click-and-drag over empty grid space does not select images

- **WHEN** select mode is active
- **AND** the user presses and drags over empty space in the gallery grid
- **THEN** no images become selected as a result of that drag
