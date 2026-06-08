### Requirement: System folder navigation entries
The sidebar SHALL display three pinned system entries at the top of the navigation: **All**, **Unsorted**, and **Trash**. All and Unsorted SHALL use the default text color. Trash SHALL use a muted/de-emphasized color. Selecting a system entry SHALL navigate to the corresponding route (`/`, `/unsorted`, `/trash`).

#### Scenario: All three system entries are visible
- **WHEN** the sidebar is rendered
- **THEN** "All", "Unsorted", and "Trash" entries are displayed in that order at the top of the nav

#### Scenario: Active system entry is highlighted
- **WHEN** the current route matches a system entry's route
- **THEN** that entry is shown with an active/selected style

#### Scenario: Trash entry is visually de-emphasized
- **WHEN** the sidebar is rendered
- **THEN** the "Trash" entry uses a muted color distinct from "All" and "Unsorted"

---

### Requirement: Nested folder tree rendering
The sidebar SHALL render the user folder list as a tree, derived from the `parent_id` field returned by `GET /folders`. Root folders (those with `parent_id: null`) are displayed at the top level. Each folder with children SHALL show an expand/collapse toggle. Depth-based indentation SHALL increase per nesting level.

#### Scenario: Root folders appear at top level
- **WHEN** `GET /folders` returns folders with `parent_id: null`
- **THEN** those folders are rendered at the top level of the user folder tree

#### Scenario: Child folders appear indented under their parent
- **WHEN** `GET /folders` returns a folder with a `parent_id` matching another folder's id
- **THEN** the child folder is rendered indented below its parent

#### Scenario: Expand toggle shows and hides children
- **WHEN** the user clicks the expand toggle on a folder that has children
- **THEN** the children are shown if previously hidden, or hidden if previously shown

#### Scenario: Folder without children shows no toggle
- **WHEN** a folder has no children
- **THEN** no expand/collapse toggle is rendered for that folder

---

### Requirement: New subfolder via context menu
The sidebar SHALL include a "New subfolder" item in the context menu for each user folder. Selecting it SHALL open the folder name dialog with `parent_id` pre-set to the right-clicked folder's id. On confirm, the system SHALL call `POST /folders` with both `name` and `parent_id`, then refetch the folder list.

#### Scenario: New subfolder creates a child folder
- **WHEN** the user right-clicks a folder, selects "New subfolder", enters a name, and confirms
- **THEN** `POST /folders` is called with the entered name and the parent folder's id as `parent_id`
- **AND** the folder list is refetched and the new child folder appears under the parent

#### Scenario: User cancels subfolder creation
- **WHEN** the user opens the new subfolder dialog and dismisses it without confirming
- **THEN** no API call is made and the folder list is unchanged

---

### Requirement: Section label separating system entries from user folders
The sidebar SHALL render a visual section label (e.g., "FOLDERS") between the system entries and the user folder tree. A horizontal divider SHALL also appear between the Trash entry and the section label. An icon button SHALL be displayed adjacent to the section label that opens the new-folder dialog.

#### Scenario: Section label is present
- **WHEN** the sidebar is rendered
- **THEN** a "FOLDERS" label is visible between the Trash entry and the user folder list

#### Scenario: New-folder icon button is present beside the section label
- **WHEN** the sidebar is rendered
- **THEN** a small icon button is visible adjacent to the "FOLDERS" label

#### Scenario: New-folder icon button opens the new-folder dialog
- **WHEN** the user clicks the icon button beside the "FOLDERS" label
- **THEN** the new-folder dialog opens, identical to the dialog opened by the footer "+ New folder" button

---

### Requirement: Folder list filter input
The sidebar SHALL display a "Filter folders…" input in the bottom section of the sidebar, above the `ProfileMenu` (username/avatar), that filters the displayed folders by name, case-insensitively, as the user types. Filtering SHALL operate on the already-loaded folder list with no additional network request. When a folder matches but its ancestors do not, the matching folder's ancestor chain SHALL also remain visible so the match's place in the hierarchy is preserved.

#### Scenario: Filter input is positioned above the profile menu
- **WHEN** the sidebar is rendered
- **THEN** the "Filter folders…" input appears in the bottom section of the sidebar, above the `ProfileMenu`

#### Scenario: Typing filters the folder tree instantly
- **WHEN** the user types a term into the folder filter input
- **THEN** folders whose name contains the term, case-insensitively, are displayed
- **AND** no additional network request is made

#### Scenario: Ancestors of a matching folder remain visible
- **WHEN** the filter term matches a nested folder but not its parent folders
- **THEN** the parent folders are still displayed, providing hierarchy context for the match

#### Scenario: Non-matching branches are hidden
- **WHEN** the filter term does not match a folder or any of its descendants
- **THEN** that folder and its subtree are not displayed

#### Scenario: Clearing the filter restores the full tree
- **WHEN** the user clears the folder filter input
- **THEN** the full folder tree is displayed again
