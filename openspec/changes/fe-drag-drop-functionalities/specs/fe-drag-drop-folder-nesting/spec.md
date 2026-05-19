## ADDED Requirements

### Requirement: Folder items are draggable

Each folder item in `FolderSidebar` SHALL be wrapped with `useDraggable` carrying `{ type: 'folder', folderId: string, name: string, parentId: string | null }`. The same 8px activation constraint SHALL apply.

#### Scenario: User begins dragging a folder item

- **WHEN** the user presses and holds a folder item and moves the pointer at least 8px
- **THEN** dnd-kit activates the drag with the folder item as the active drag item

---

### Requirement: Folder items are also drop targets for folder drags

Each folder item SHALL carry a `useDroppable` with `{ type: 'folder', folderId: string }`. While a folder drag is active and the pointer is over a folder item, that item SHALL highlight — unless the target is in the dragged folder's own subtree (including itself), in which case no highlight is shown.

#### Scenario: Valid folder drop target highlights on hover

- **WHEN** the user drags folder A and hovers over folder B (which is not a descendant of A)
- **THEN** folder B is visually highlighted

#### Scenario: Dragged folder's own subtree does not highlight

- **WHEN** the user drags folder A and hovers over folder A itself or one of its descendants
- **THEN** no highlight appears on that target

---

### Requirement: Dropping folder onto another folder nests it

When a folder drag ends over a valid folder drop target, `onDragEnd` SHALL first verify the target is not in the dragged folder's subtree. If valid, it SHALL call `PUT /folders/:id` with `{ name: folder.name, parent_id: targetFolderId }`. On success the folder list SHALL be invalidated. On error a toast SHALL be shown.

#### Scenario: Folder dropped onto another folder becomes its child

- **WHEN** the user drops folder A onto folder B (and B is not a descendant of A)
- **THEN** `PUT /folders/:folderId` is called with `{ "name": "<A's name>", "parent_id": "<B's id>" }`
- **AND** the folder list is refreshed

#### Scenario: Dropping a folder onto its current parent is a no-op

- **WHEN** the user drops a folder onto its existing parent folder
- **THEN** no PUT request is made

#### Scenario: Circular nesting is prevented

- **WHEN** the user attempts to drop folder A onto one of A's own descendants
- **THEN** no PUT request is made
- **AND** no error toast is shown

#### Scenario: Move failure shows error toast

- **WHEN** the PUT request fails
- **THEN** an error toast is shown

---

### Requirement: Root drop zone below the folder list promotes folder to root

A droppable zone with `{ type: 'root' }` SHALL be rendered below the last folder item in the sidebar nav. It SHALL be visible only while a folder drag is active, indicated by a dashed border and a label such as "Move to root". Dropping a folder onto it SHALL call `PUT /folders/:id` with `{ name: folder.name, parent_id: null }`.

#### Scenario: Root drop zone appears only during folder drag

- **WHEN** no drag is active
- **THEN** the root drop zone is not visible in the sidebar

- **WHEN** a folder drag is active
- **THEN** the root drop zone appears below the folder list

#### Scenario: Dropping a subfolder onto the root zone promotes it to root

- **WHEN** the user drops a nested folder onto the root drop zone
- **THEN** `PUT /folders/:folderId` is called with `{ "name": "<folder's name>", "parent_id": null }`
- **AND** the folder list is refreshed

#### Scenario: Dropping a root folder onto the root zone is a no-op

- **WHEN** the user drops a folder that already has no parent onto the root zone
- **THEN** no PUT request is made
