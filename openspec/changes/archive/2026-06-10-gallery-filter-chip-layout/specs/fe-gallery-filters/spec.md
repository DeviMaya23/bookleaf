## MODIFIED Requirements

### Requirement: Filters control in gallery toolbar

The system SHALL display a "Filters" button in the gallery toolbar (`AppLayout.tsx`'s toolbar row, outside the search/sort `max-w-xs` group — per Option C of the "Filter & Sort" design handoff) for the All, Unsorted, and Folder views. The button SHALL be hidden entirely when viewing Trash. Clicking the button SHALL open a panel built from this codebase's existing `DropdownMenu`/`DropdownMenuContent` primitives, containing the available filter sections — File type, Tags, and (where applicable) Folder — each separated by an inline divider header. The File type section SHALL render as a multi-select toggle/chip row (`ui/toggle-group.tsx` + `ui/toggle.tsx`); the Tags section and the Folder section (where shown) SHALL each render a search input above a `DropdownMenuCheckboxItem` list.

#### Scenario: Filters button opens the filter panel

- **WHEN** the user clicks the "Filters" button in the toolbar while viewing All, Unsorted, or a folder
- **THEN** a panel opens showing the available filter sections, each under its own divider header
- **AND** the File type section shows multi-select toggle/chip controls
- **AND** the Tags section (and Folder section, where shown) each show a search input above a checkbox list

#### Scenario: Filters button is hidden in Trash

- **WHEN** the user is viewing Trash
- **THEN** no "Filters" button is rendered in the toolbar

### Requirement: Filter sections differ by view type

The filter panel SHALL show different sections depending on the active view, in the order File type, Tags, Folder:

- **All**: File type, Tags, Folder
- **Unsorted**: File type, Tags — no Folder section, since the Unsorted view always sends `unfiled=true`, and combining that with any `folder_ids` would be a contradiction (an unfiled image cannot belong to a folder)
- **Folder view**: File type, Tags — no Folder section, since the user is already scoped to one folder

#### Scenario: All view shows all three sections

- **WHEN** the user opens the filter panel while viewing All
- **THEN** the panel shows File type, Tags, and Folder sections, in that order

#### Scenario: Unsorted view omits the Folder section

- **WHEN** the user opens the filter panel while viewing Unsorted
- **THEN** the panel shows only File type and Tags sections

#### Scenario: Folder view omits the Folder section

- **WHEN** the user opens the filter panel while viewing a folder
- **THEN** the panel shows only File type and Tags sections

### Requirement: File type options are a fixed list of stored MIME types

The File type section SHALL list the MIME types that can actually appear on a stored image — `image/jpeg`, `image/png`, `image/gif`, `image/webp`, `image/avif` — displayed as multi-select toggle/chip controls with friendly labels (`JPEG`, `PNG`, `GIF`, `WEBP`, `AVIF`). `image/heic` and `image/svg+xml` SHALL NOT be offered, since `image/heic` is converted to `image/jpeg` before upload and is never a stored value, and `image/svg+xml` is not an accepted upload type.

#### Scenario: File type section shows the fixed label set

- **WHEN** the user opens the filter panel
- **THEN** the File type section shows toggle controls labeled `JPEG`, `PNG`, `GIF`, `WEBP`, `AVIF`, in that order
- **AND** no `SVG` or `HEIC` option is shown

### Requirement: Selecting filters is multi-select and does not require a separate apply step

Each filter section SHALL allow selecting any number of values — multi-select via checkboxes for Tags and Folder, and via toggle controls for File type. Toggling a checkbox or toggle control SHALL immediately update the active filter state — there is no separate "Apply"/"Done" action required, and the panel SHALL remain open so the user can toggle multiple values in succession.

#### Scenario: Toggling multiple controls keeps the panel open

- **WHEN** the user checks a tag, then toggles a file type, within the same panel session
- **THEN** both selections take effect immediately
- **AND** the panel remains open after each toggle

## ADDED Requirements

### Requirement: Tags and Folder sections are searchable and independently scrollable

The Tags section, and the Folder section where shown, SHALL each display a text input above their checkbox list. Typing in this input SHALL filter that section's list to items whose name contains the typed text (case-insensitive substring match); the filter SHALL NOT affect any other section. When the typed text matches no items, the section SHALL render no checkbox items at all — including items that are currently checked. Each section's checkbox list SHALL have its own fixed maximum height with independent vertical scrolling, separate from the filter panel's own scroll region. Typing in either search input SHALL NOT trigger the filter panel's keyboard navigation or close the panel.

#### Scenario: Searching tags filters the tag list

- **WHEN** the user types text into the Tags section's search input
- **THEN** the Tags checkbox list shows only tags whose name contains that text (case-insensitive)
- **AND** the Folder section's list (if shown) is unaffected

#### Scenario: Searching folders filters the folder list

- **WHEN** the user types text into the Folder section's search input
- **THEN** the Folder checkbox list shows only folders whose name contains that text (case-insensitive)
- **AND** the Tags section's list is unaffected

#### Scenario: No matches hides the list entirely, even for checked items

- **WHEN** the user types text into the Tags section's search input that matches no tag, including a tag that is currently checked
- **THEN** the Tags section shows no checkbox items
- **AND** the checked tag remains selected (still reflected in the active filter chip row) even though it is not visible in the list

#### Scenario: Long lists scroll independently within their section

- **WHEN** the Tags section's checkbox list exceeds its fixed maximum height
- **THEN** that section scrolls independently of the Folder section and of the overall filter panel

#### Scenario: Typing in a search input does not trigger menu typeahead or close the panel

- **WHEN** the user types a letter into the Tags or Folder search input while the filter panel is open
- **THEN** the typed character appears in the input
- **AND** the filter panel does not close or shift focus to a matching menu item
