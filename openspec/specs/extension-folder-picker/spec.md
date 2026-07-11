# Spec: Extension Folder Picker

## Purpose

Defines the folder tree utility functions (build and filter) and the FolderPicker panel view used in the extension popup to let users select a default save destination folder.

## Requirements

### Requirement: Folder tree utilities

`extensions/src/lib/folderTree.ts` SHALL export two pure functions ported from the FE, operating on a local `FolderNode` type defined in that file:

```ts
interface FolderNode {
  id: string
  name: string
  parent_id: string | null
  children: FolderNode[]
}
```

- `buildFolderTree(folders: Array<{ id: string; name: string; parent_id: string | null }>): FolderNode[]` — builds a tree from a flat list. Root nodes (those with `parent_id === null` or whose parent is absent from the list) are returned as top-level entries. Children are attached in API response order.
- `filterFolderTree(nodes: FolderNode[], term: string): FolderNode[]` — returns a filtered tree where a node is included if its name contains `term` (case-insensitive) OR if any of its descendants match. When a node is included via descendant match, only the matching descendants are included as children (non-matching siblings are dropped). When a node itself matches, its non-matching children are dropped.

#### Scenario: Root folders with no parent appear at tree top level

- **WHEN** `buildFolderTree` receives folders where some have `parent_id: null`
- **THEN** those folders appear as root nodes with no parent in the returned tree

#### Scenario: Child folders are nested under their parent

- **WHEN** `buildFolderTree` receives a folder with a `parent_id` matching another folder's `id`
- **THEN** that folder appears in the `children` array of its parent node

#### Scenario: Filter includes a parent when its name matches, children dropped

- **WHEN** `filterFolderTree` is called with a term matching "Clothes" and "Clothes" has a child "Shoes" that does not match
- **THEN** the result includes "Clothes" with an empty `children` array; "Shoes" is not present

#### Scenario: Filter includes a parent when a descendant matches

- **WHEN** `filterFolderTree` is called with a term matching "Shoes" and "Shoes" is a child of "Clothes" which is a child of "North"
- **THEN** the result includes "North" → "Clothes" → "Shoes" (all ancestors of the match are present)

#### Scenario: Non-matching node with no matching descendants is excluded

- **WHEN** `filterFolderTree` is called with a term that matches neither a node nor any of its descendants
- **THEN** that node does not appear in the result

#### Scenario: Empty term returns all nodes unchanged

- **WHEN** `filterFolderTree` is called with an empty string
- **THEN** all nodes are returned with their full original children

### Requirement: FolderPicker panel view

The FolderPicker panel SHALL be a full-width popup view (same 320px width as the rest of the popup) accessible from the Settings view. It SHALL:

- Display a back arrow header labelled "Default folder" (same header style as the Settings view).
- Render a text input at the top of the body for filtering, with placeholder `"Search folders…"`.
- Render a "None (Unsorted)" row below the filter input as a fixed first option (always visible, not filtered out).
- Render the folder tree below the fixed row: root-level folders at zero indentation; each level of nesting adds 12px of left padding. Folder names are truncated with ellipsis if they overflow the panel width.
- Call `getFolders()` on mount. While loading, the tree area SHALL display a disabled placeholder and the filter input SHALL be disabled. On fetch error, the tree area SHALL show only the "None (Unsorted)" row with the filter input re-enabled.
- When the filter input is non-empty, apply `filterFolderTree` to the tree and render the filtered result. The "None (Unsorted)" row is always shown regardless of filter value.
- Clicking any folder row (including "None (Unsorted)") SHALL immediately call `setDefaultFolder({ id, name })` for a named folder or `clearDefaultFolder()` for "None", then navigate back to the Settings view.
- The currently stored default folder (if any) SHALL be visually indicated (e.g. a checkmark or distinct text weight) in the tree.

#### Scenario: FolderPicker renders full tree when filter is empty

- **WHEN** FolderPicker mounts and `getFolders()` resolves with a nested set of folders and the filter input is empty
- **THEN** all folders are rendered in tree order with their children indented beneath them

#### Scenario: Typing in the filter narrows the tree

- **WHEN** the user types "shoe" into the filter input and the tree contains North → Clothes → Shoes
- **THEN** the visible tree shows North → Clothes → Shoes (all ancestors of the match retained); other non-matching branches are hidden

#### Scenario: Selecting a folder persists it and navigates back

- **WHEN** the user clicks a folder row in the FolderPicker
- **THEN** `setDefaultFolder({ id, name })` is called with that folder's data and the popup navigates back to the Settings view

#### Scenario: Selecting "None (Unsorted)" clears the stored default and navigates back

- **WHEN** the user clicks the "None (Unsorted)" row
- **THEN** `clearDefaultFolder()` is called and the popup navigates back to the Settings view

#### Scenario: Currently stored folder is visually indicated

- **WHEN** FolderPicker mounts and a default folder is stored
- **THEN** the row corresponding to the stored folder is visually distinguished from unselected rows

#### Scenario: Fetch error shows only "None (Unsorted)"

- **WHEN** `getFolders()` rejects on FolderPicker mount
- **THEN** only the "None (Unsorted)" row is shown; the stored preference is unchanged
