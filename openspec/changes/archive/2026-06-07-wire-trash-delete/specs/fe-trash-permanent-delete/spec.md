## ADDED Requirements

### Requirement: Permanently delete a single trashed image via context menu

The image card context menu in the trash view SHALL include a "Delete permanently" item below a separator after "Restore", styled in the destructive colour. Selecting it SHALL open a confirmation dialog warning that the action cannot be undone. Confirming SHALL call `DELETE /images/trash/:id`. On success, the image SHALL be removed from the trash view and a success toast SHALL be shown. On failure, the image SHALL remain and an error toast SHALL be shown.

#### Scenario: User permanently deletes a trashed image

- **WHEN** the user right-clicks a trashed image, selects "Delete permanently", and confirms the dialog
- **THEN** `DELETE /images/trash/:id` is called
- **AND** the image is removed from the trash view
- **AND** a success toast is shown

#### Scenario: Deletion fails

- **WHEN** the user confirms permanent deletion and `DELETE /images/trash/:id` returns an error
- **THEN** the image remains in the trash view
- **AND** an error toast is shown

#### Scenario: User dismisses the confirmation dialog

- **WHEN** the user right-clicks a trashed image, selects "Delete permanently", and cancels the dialog
- **THEN** no API call is made and the image remains in the trash view

---

### Requirement: Empty trash via Trash sidebar context menu

The Trash entry in the sidebar SHALL support a right-click context menu with a single "Empty trash" item styled in the destructive colour. Selecting it SHALL open a confirmation dialog warning that all images will be permanently deleted and the action cannot be undone. Confirming SHALL call `DELETE /images/trash`. On success, the trash view SHALL show the empty state and a success toast SHALL be shown. On failure, an error toast SHALL be shown.

#### Scenario: User empties trash successfully

- **WHEN** the user right-clicks the Trash sidebar entry, selects "Empty trash", and confirms the dialog
- **THEN** `DELETE /images/trash` is called
- **AND** the trash view shows the empty state
- **AND** a success toast is shown

#### Scenario: Empty trash fails

- **WHEN** the user confirms "Empty trash" and `DELETE /images/trash` returns an error
- **THEN** an error toast is shown and the trash view is unchanged

#### Scenario: Empty trash when trash is already empty

- **WHEN** the user confirms "Empty trash" and trash has no images
- **THEN** `DELETE /images/trash` is called, returns 204, and the trash view remains in the empty state
