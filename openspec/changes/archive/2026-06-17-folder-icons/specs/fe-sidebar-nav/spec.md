## MODIFIED Requirements

### Requirement: System folder navigation entries
The sidebar SHALL display three pinned system entries at the top of the navigation: **All**, **Unsorted**, and **Trash**. All and Unsorted SHALL use the default text color. Trash SHALL use a muted/de-emphasized color. Selecting a system entry SHALL navigate to the corresponding route (`/app`, `/app/unsorted`, `/app/trash`). When folder icons are enabled, each system entry SHALL display a fixed icon to the left of its label (see the `folder-icon-customization` capability for the specific icons and the visibility toggle).

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
The sidebar SHALL render the user folder list as a tree, derived from the `parent_id` field returned by `GET /folders`. Root folders (those with `parent_id: null`) are displayed at the top level. Each folder with children SHALL show an expand/collapse toggle. Depth-based indentation SHALL increase per nesting level. When folder icons are enabled, each folder row SHALL display the folder's icon (or the default, if unset) between the expand/collapse toggle and the folder name (see the `folder-icon-customization` capability).

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
