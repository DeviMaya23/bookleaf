# fe-tags-view Specification

## Purpose

Provides a dedicated tags management view where users can see all their tags, rename them, delete them, and navigate to the gallery pre-filtered by a chosen tag.

## Requirements

### Requirement: Tags entry in sidebar navigation

The system SHALL display a **Tags** navigation entry in the folder sidebar, positioned below the Unsorted entry and above the Trash entry. Clicking the entry SHALL navigate to `/app/tags`.

#### Scenario: Tags entry is visible in sidebar

- **WHEN** the user opens the app
- **THEN** a Tags entry is visible in the sidebar below Unsorted and above Trash

#### Scenario: Tags entry navigates to tags view

- **WHEN** the user clicks the Tags entry
- **THEN** the route changes to `/app/tags` and the tags view is rendered in the main content area

#### Scenario: Tags entry shows active state

- **WHEN** the current route is `/app/tags`
- **THEN** the Tags sidebar entry renders in its active/selected visual state

### Requirement: Tags route and view type

The system SHALL register `/app/tags` as a route, rendering `AppLayout`. The `AppView` type SHALL include a `{ type: 'tags' }` variant. `useAppView` SHALL return `{ type: 'tags' }` when the pathname is `/app/tags`.

#### Scenario: useAppView returns tags view at /app/tags

- **WHEN** the pathname is `/app/tags`
- **THEN** `useAppView` returns `{ type: 'tags' }`

### Requirement: Gallery toolbar is hidden in tags view

The system SHALL NOT render the `GalleryToolbar` when the active view is `{ type: 'tags' }`. The image gallery and its controls do not apply to the tags view.

#### Scenario: Toolbar absent in tags view

- **WHEN** the user navigates to `/app/tags`
- **THEN** the gallery toolbar (search, sort, filter, upload) is not visible

### Requirement: Tags view lists all user tags alphabetically grouped

The system SHALL display all of the user's tags in the main content area when the view is `{ type: 'tags' }`. Tags SHALL be sorted alphabetically and grouped under single-letter section headers (e.g. "B", "D"). Each group header SHALL show the letter and the count of tags in that group. Each tag row SHALL display the tag name and its `image_count` (the number of images the tag is applied to).

#### Scenario: Tags are displayed in alphabetical groups

- **WHEN** the user has tags "baking", "dessert", "berries"
- **THEN** the tags view shows a "B" group containing "baking" and "berries", and a "D" group containing "dessert"

#### Scenario: Image count is shown per tag

- **WHEN** a tag has 12 images
- **THEN** the tag row displays 12 as the image count

#### Scenario: Tags view is empty when user has no tags

- **WHEN** the user has no tags
- **THEN** the tags view renders an empty state

### Requirement: Tags view has local search

The system SHALL display a search input at the top of the tags view that filters the displayed tag list by name (case-insensitive substring match). The search applies client-side over the already-fetched tag list; no network request is made when the search term changes.

#### Scenario: Typing in tag search filters the list

- **WHEN** the user types "bak" in the tags view search input
- **THEN** only tags whose name contains "bak" (case-insensitive) are shown

#### Scenario: Clearing search restores full list

- **WHEN** the user clears the search input
- **THEN** all tags are shown again

### Requirement: Tags view fetches tag list with image counts

The system SHALL request `GET /tags?include_count=true` when rendering the tags view, so that each tag row can display how many images it is applied to.

#### Scenario: Tags view calls GET /tags with include_count

- **WHEN** the user navigates to `/app/tags`
- **THEN** the frontend fetches `GET /tags?include_count=true`
- **AND** each tag row shows the returned `image_count`

### Requirement: Clicking a tag navigates to All view with tag pre-filtered

The system SHALL navigate to `/app` when the user clicks a tag name in the tags view, passing the tag's ID as pre-fill data via router state (`{ filterTagIds: [id] }`). The All view SHALL seed its initial tag filter from this router state on mount.

#### Scenario: Clicking a tag navigates to All with filter pre-filled

- **WHEN** the user clicks the tag "baking" in the tags view
- **THEN** the route changes to `/app`
- **AND** the gallery filter is pre-seeded with the "baking" tag selected
- **AND** the image grid shows only images tagged with "baking"

### Requirement: Per-tag action menu (rename and delete)

Each tag row SHALL display a `...` action button (visible on hover for pointer devices). Clicking it SHALL open a menu with two items: **Rename** and **Delete tag**.

#### Scenario: Action menu opens on ... click

- **WHEN** the user clicks the `...` button on a tag row
- **THEN** a menu appears with Rename and Delete tag options

### Requirement: Inline rename with duplicate-name validation

Selecting **Rename** SHALL transform the tag row into an inline text input pre-filled with the current tag name, alongside **Save** and **Cancel** buttons. Submitting a name that already exists for this user SHALL display an inline error "That tag already exists. Try another name." without closing the edit row. Submitting a valid unique name SHALL call `PUT /tags/:id`, update the tag in the list, and return the row to its normal display state.

#### Scenario: Rename row replaces normal row

- **WHEN** the user selects Rename from the tag action menu
- **THEN** the tag row becomes an input field pre-filled with the current name and Save/Cancel buttons are shown

#### Scenario: Saving a unique name updates the tag

- **WHEN** the user changes the name to one that does not already exist and clicks Save
- **THEN** `PUT /tags/:id` is called with the new name
- **AND** the row returns to normal display showing the updated name

#### Scenario: Saving a duplicate name shows inline error

- **WHEN** the user enters a name that matches an existing tag and clicks Save
- **THEN** an inline error "That tag already exists. Try another name." is shown below the input
- **AND** the input remains open

#### Scenario: Cancel discards the rename

- **WHEN** the user clicks Cancel during a rename
- **THEN** the row returns to its normal display state with the original name unchanged

### Requirement: Inline delete confirmation

Selecting **Delete tag** SHALL transform the tag row into an inline confirmation state showing "Remove tag?" with **Cancel** and **Delete** buttons, without using a modal dialog. Clicking **Delete** SHALL call `DELETE /tags/:id`, remove the tag from the list, and invalidate the tags cache. Clicking **Cancel** SHALL return the row to its normal state.

#### Scenario: Delete confirmation row replaces normal row

- **WHEN** the user selects Delete tag from the action menu
- **THEN** the row transforms to show "Remove tag?" with Cancel and Delete buttons

#### Scenario: Confirming delete removes the tag

- **WHEN** the user clicks Delete in the confirmation row
- **THEN** `DELETE /tags/:id` is called
- **AND** the tag is removed from the list

#### Scenario: Cancelling delete restores normal row

- **WHEN** the user clicks Cancel in the delete confirmation row
- **THEN** the row returns to its normal display state
