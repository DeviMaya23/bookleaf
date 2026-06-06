## MODIFIED Requirements

### Requirement: Restore image from trash via context menu

Each image in the trash view SHALL have a context menu with a **Restore** item and, below a separator, a **Delete permanently** item styled in the destructive colour. Selecting Restore SHALL call `POST /images/:id/restore`. On success, the image SHALL be removed from the trash view and a success toast SHALL be shown. Selecting Delete permanently SHALL open a confirmation dialog — see the `fe-trash-permanent-delete` spec for that behaviour.

#### Scenario: User restores an image successfully

- **WHEN** the user right-clicks an image in the trash view and selects "Restore"
- **THEN** `POST /images/:id/restore` is called
- **AND** the image disappears from the trash view
- **AND** a success toast is shown

#### Scenario: Restore fails with an error toast

- **WHEN** `POST /images/:id/restore` returns an error
- **THEN** the image remains in the trash view
- **AND** an error toast is shown
