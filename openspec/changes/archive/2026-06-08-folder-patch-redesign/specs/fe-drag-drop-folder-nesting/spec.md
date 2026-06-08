## MODIFIED Requirements

### Requirement: Dropping folder onto another folder nests it

When a folder drag ends over a valid folder drop target, `onDragEnd` SHALL first verify the target is not in the dragged folder's subtree. If valid, it SHALL call `updateFolder(getToken, folder.id, { parent_id: targetFolderId })`, sending only the `parent_id` field — the folder's `name` is left untouched. On success the folder list SHALL be invalidated. On error a toast SHALL be shown.

#### Scenario: Folder dropped onto another folder becomes its child

- **WHEN** the user drops folder A onto folder B (and B is not a descendant of A)
- **THEN** `updateFolder` is called with `id` = A's id and `{ "parent_id": "<B's id>" }`, with no `name` field in the body
- **AND** the folder list is refreshed
- **AND** a success toast is shown

#### Scenario: Dropping a folder onto its current parent is a no-op

- **WHEN** the user drops a folder onto its existing parent folder
- **THEN** no `updateFolder` request is made

#### Scenario: Circular nesting is prevented

- **WHEN** the user attempts to drop folder A onto one of A's own descendants
- **THEN** no `updateFolder` request is made
- **AND** no error toast is shown

#### Scenario: Move failure shows error toast

- **WHEN** the `updateFolder` request fails
- **THEN** an error toast is shown

---

### Requirement: Root drop zone below the folder list promotes folder to root

A droppable zone with `{ type: 'root' }` SHALL be rendered below the last folder item in the sidebar nav. It SHALL be visible only while a folder drag is active, indicated by a dashed border and a label such as "Move to root". Dropping a folder onto it SHALL call `updateFolder(getToken, folder.id, { parent_id: null })`, sending only the `parent_id` field — the folder's `name` is left untouched.

#### Scenario: Root drop zone appears only during folder drag

- **WHEN** no drag is active
- **THEN** the root drop zone is not visible in the sidebar

- **WHEN** a folder drag is active
- **THEN** the root drop zone appears below the folder list

#### Scenario: Dropping a subfolder onto the root zone promotes it to root

- **WHEN** the user drops a nested folder onto the root drop zone
- **THEN** `updateFolder` is called with `id` = the folder's id and `{ "parent_id": null }`, with no `name` field in the body
- **AND** the folder list is refreshed
- **AND** a success toast is shown

#### Scenario: Dropping a root folder onto the root zone is a no-op

- **WHEN** the user drops a folder that already has no parent onto the root zone
- **THEN** no `updateFolder` request is made
