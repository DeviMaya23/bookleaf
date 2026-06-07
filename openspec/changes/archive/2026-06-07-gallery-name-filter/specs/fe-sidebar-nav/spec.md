## ADDED Requirements

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
