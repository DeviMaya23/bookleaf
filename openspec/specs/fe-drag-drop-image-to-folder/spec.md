## ADDED Requirements

### Requirement: Image cards are draggable

Each image card in `ImageGrid` SHALL be wrapped with a dnd-kit `useDraggable` hook. The drag item data SHALL carry `{ type: 'image', imageId: string, currentFolderId: string | null }`. Dragging SHALL be activated only after the pointer moves at least 8px (activation constraint) to avoid interfering with click-to-select and scroll.

#### Scenario: User begins dragging an image card

- **WHEN** the user presses and holds on an image card then moves the pointer at least 8px
- **THEN** dnd-kit activates the drag and the image card becomes the active drag item

#### Scenario: Click on image card still opens right panel

- **WHEN** the user clicks an image card without moving the pointer 8px
- **THEN** the right panel opens for that image and no drag is initiated

---

### Requirement: Floating drag overlay for image drags

A `DragOverlay` portal SHALL render a thumbnail mini-card while an image is being dragged, following the pointer. The overlay SHALL display the image thumbnail at a fixed small size (approx 80×80px, cover crop) with a subtle drop shadow. The original card in the grid SHALL reduce opacity to 0.4 while dragging.

#### Scenario: Overlay appears during image drag

- **WHEN** an image drag is active
- **THEN** a floating thumbnail mini-card follows the pointer
- **AND** the original card in the grid is visually dimmed

#### Scenario: Overlay disappears on drag end

- **WHEN** the drag ends (drop or cancel)
- **THEN** the floating overlay disappears and the original card returns to full opacity

---

### Requirement: Sidebar folder items are drop targets for images

Each folder item in `FolderSidebar` SHALL be wrapped with `useDroppable` carrying `{ type: 'folder', folderId: string }`. While an image drag is active and the pointer is over a folder item, that folder item SHALL highlight (accent background, full opacity text).

#### Scenario: Folder item highlights when image is dragged over it

- **WHEN** the user drags an image card and hovers over a folder item in the sidebar
- **THEN** the folder item is visually highlighted with an accent background

#### Scenario: Highlight clears when pointer leaves

- **WHEN** the user moves the dragged image off a folder item
- **THEN** the folder item returns to its default visual state

---

### Requirement: Dropping image onto a folder moves it

When an image drag ends over a folder drop target, `AppLayout`'s `onDragEnd` handler SHALL call `POST /images/:id/move-folder` with `{ from_folder_id: currentFolderId, to_folder_id: targetFolderId }`. On success the image queries SHALL be invalidated. On error a toast SHALL be shown.

#### Scenario: Image dropped onto a folder is moved into that folder

- **WHEN** the user drops an image card onto a sidebar folder item
- **THEN** `POST /images/:imageId/move-folder` is called with `{ "from_folder_id": "<currentFolderId>", "to_folder_id": "<targetFolderId>" }`
- **AND** the image list is refreshed
- **AND** a success toast is shown

#### Scenario: Dropping image onto its current folder is a no-op

- **WHEN** the user drops an image onto the folder it is already in (currentFolderId == targetFolderId)
- **THEN** no request is made

#### Scenario: Move failure shows error toast

- **WHEN** the POST request fails
- **THEN** an error toast is shown
- **AND** the image list is refreshed to restore correct state

---

### Requirement: "Unsorted" system entry is a drop target that clears folder assignment

The "Unsorted" sidebar entry SHALL be a `useDroppable` target carrying `{ type: 'unsorted' }`. Dropping an image onto it SHALL call `POST /images/:id/move-folder` with `{ from_folder_id: currentFolderId, to_folder_id: null }`.

#### Scenario: Dropping image onto Unsorted removes its folder assignment

- **WHEN** the user drops an image card onto the "Unsorted" sidebar entry
- **THEN** `POST /images/:imageId/move-folder` is called with `{ "from_folder_id": "<currentFolderId>", "to_folder_id": null }`
- **AND** the image list is refreshed
- **AND** a success toast is shown

#### Scenario: Dropping an already-unsorted image onto Unsorted is a no-op

- **WHEN** the user drops an image that has no folder (currentFolderId == null) onto "Unsorted"
- **THEN** no request is made

---

### Requirement: moveImageFolder API function

The system SHALL provide a `moveImageFolder(getToken, imageId, fromFolderId, toFolderId)` function in `frontend/src/lib/images.ts` that calls `POST /images/:imageId/move-folder`.

```ts
export async function moveImageFolder(
  getToken: () => Promise<string | null>,
  imageId: string,
  fromFolderId: string | null,
  toFolderId: string | null,
): Promise<void>
```

The function SHALL throw on any non-2xx response.

#### Scenario: moveImageFolder calls correct endpoint

- **WHEN** `moveImageFolder` is called with a valid imageId, fromFolderId, and toFolderId
- **THEN** a `POST` request is made to `/images/:imageId/move-folder` with the correct JSON body
