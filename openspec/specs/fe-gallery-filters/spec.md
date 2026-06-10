# fe-gallery-filters Specification

## Purpose
Defines the gallery filtering UI/UX: a "Filters" control in the gallery toolbar that lets users narrow the image list by file type, tags, and (where applicable) folder, plus the resulting active-filter chip row, view-specific filter sections, and the client- and server-side mechanics for applying selected filters.
## Requirements
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

### Requirement: Filters button shows a count of active filters

The "Filters" button SHALL display a badge with the total count of selected filter values across all sections (tags + file types + folders), and SHALL render in its active visual state (matching the existing sort button's `default`/`outline` variant switch) whenever that count is greater than zero. The badge SHALL be omitted when the count is zero.

#### Scenario: Badge reflects total selected filter count

- **WHEN** the user has selected 2 tags and 1 file type
- **THEN** the "Filters" button shows a badge with the value `3`
- **AND** the button renders in its active visual state

#### Scenario: No badge when no filters are selected

- **WHEN** no tags, file types, or folders are selected
- **THEN** the "Filters" button shows no badge and renders in its normal (inactive) visual state

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

### Requirement: Tag and folder options are sourced from existing data

The Tags section SHALL list all of the user's tags (via the existing `getTags()`), displayed by name. The Folder section (All view only) SHALL list all of the user's folders (via the existing `getFolders()`), displayed by name.

#### Scenario: Tags section lists the user's tags

- **WHEN** the user opens the filter panel
- **THEN** the Tags section shows one checkbox per tag returned by `getTags()`, labeled with the tag's name

#### Scenario: Folder section lists the user's folders

- **WHEN** the user opens the filter panel while viewing All
- **THEN** the Folder section shows one checkbox per folder returned by `getFolders()`, labeled with the folder's name

### Requirement: Selecting filters is multi-select and does not require a separate apply step

Each filter section SHALL allow selecting any number of values — multi-select via checkboxes for Tags and Folder, and via toggle controls for File type. Toggling a checkbox or toggle control SHALL immediately update the active filter state — there is no separate "Apply"/"Done" action required, and the panel SHALL remain open so the user can toggle multiple values in succession.

#### Scenario: Toggling multiple controls keeps the panel open

- **WHEN** the user checks a tag, then toggles a file type, within the same panel session
- **THEN** both selections take effect immediately
- **AND** the panel remains open after each toggle

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

### Requirement: Active filters surface as removable chips below the toolbar

When one or more filters are selected (across any section), the system SHALL display a row below the toolbar containing one chip per selected value (tag, file type, or folder), each showing its label and a remove (`×`) control, plus a "Clear all" action that resets all selections to empty. The chip row SHALL NOT be rendered — and SHALL reserve no vertical space — when no filters are selected. Chips SHALL use this codebase's existing tag-pill styling (as used in `TagInput.tsx`), not the design handoff's bespoke chip styling.

#### Scenario: Chip row appears when a filter is selected

- **WHEN** the user selects a tag in the filter panel
- **THEN** a row appears below the toolbar showing one chip labeled with the tag's name, with a remove control
- **AND** a "Clear all" action is shown alongside it

#### Scenario: Removing a chip clears that filter

- **WHEN** the user clicks a chip's remove control
- **THEN** that value is deselected in the filter panel
- **AND** the chip is removed from the row
- **AND** the image list updates to reflect the remaining filters

#### Scenario: Clear all resets every filter

- **WHEN** the user clicks "Clear all"
- **THEN** all selected tags, file types, and folders are deselected
- **AND** the chip row disappears
- **AND** the image list updates to show the unfiltered results for the active view

#### Scenario: No chip row when nothing is selected

- **WHEN** no filters are selected
- **THEN** no chip row is rendered below the toolbar

### Requirement: Filter selections reset when switching views

The system SHALL clear all selected tags, file types, and folders whenever the user navigates to a different view (a different folder, or All/Unsorted/Trash) — mirroring the existing search-term and sort reset behavior (`fe-gallery-search`, `fe-gallery-sort`). Filter state is per-view and session-local; it is never carried over from one view to another nor persisted across sessions.

#### Scenario: Switching views clears active filters

- **WHEN** the user has active filters selected in one view and navigates to a different view
- **THEN** all filter selections are cleared
- **AND** the new view displays its unfiltered image list (subject to that view's own default request)

### Requirement: Selected filters are sent as query parameters for All and Unsorted views

`getImages` and `getAllImages` (`src/lib/images.ts`) SHALL accept optional `tagIds`, `mimeTypes`, and `folderIds` parameters (arrays of strings). When non-empty, each SHALL be serialized as a comma-separated value and included as the `tag_ids`, `mime_types`, and `folder_ids` query parameters respectively on `GET /images`. `ImageGrid`'s query key for the `'all'` and `'unsorted'` views SHALL include the active filter selections so that changing filters triggers a re-fetch, with `useInfiniteQuery`'s existing `placeholderData: keepPreviousData` keeping prior results visible during the transition.

#### Scenario: Selected tags and file types are sent as query parameters

- **WHEN** the user selects two tags and one file type while viewing All
- **THEN** `GET /images` is called with `tag_ids=<id1>,<id2>` and `mime_types=image/jpeg`
- **AND** the displayed images update to the filtered, paginated results once the response arrives

#### Scenario: Selected folders are sent as a query parameter in the All view

- **WHEN** the user selects one folder while viewing All
- **THEN** `GET /images` is called with `folder_ids=<folderId>`

#### Scenario: No filter parameters are sent when nothing is selected

- **WHEN** no filters are selected while viewing All or Unsorted
- **THEN** `GET /images` is called without `tag_ids`, `mime_types`, or `folder_ids` parameters

### Requirement: Selected filters are applied client-side for Folder view

In Folder view, selected tags and file types SHALL be applied as a client-side filter over the already-fetched full image list (via `image.tags` and `image.mime_type`), composing with the existing client-side title-search filter. An image SHALL be shown only if it matches the search term (if any), AND has at least one tag in the selected tags (if any are selected), AND has a `mime_type` in the selected file types (if any are selected). No additional network request SHALL be made when filters change in Folder view.

#### Scenario: Filtering by tag in folder view is client-side

- **WHEN** the user selects a tag while viewing a folder
- **THEN** the displayed images are filtered to those whose `tags` include the selected tag
- **AND** no additional network request is made

#### Scenario: Combined search and filter in folder view

- **WHEN** the user has a search term and a selected file type active while viewing a folder
- **THEN** the displayed images are those matching both the search term and the selected file type

