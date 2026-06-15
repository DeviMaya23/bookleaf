## ADDED Requirements

### Requirement: Share folder section in FolderPanelContent

The system SHALL render a "Share folder" section in `FolderPanelContent`, positioned between the Notes section and the Export folder footer, using the same bordered-row layout as the Notes section. The section SHALL contain a switch labeled "Share folder" reflecting whether the folder currently has an active share link. When a share link is active, the section SHALL also show a read-only, truncated field displaying `<origin>/share/<token>` and a copy button.

#### Scenario: Section renders between Notes and Export

- **WHEN** the folder panel is shown for a folder
- **THEN** a "Share folder" section with a switch is rendered after the Notes section and before the Export folder button

### Requirement: Share state is loaded on mount

The system SHALL fetch the folder's share status via `GET /folders/:id/share` when the panel opens for a folder. A 404 response SHALL be treated as "not shared" (switch off), not as an error.

#### Scenario: Folder has no active share

- **WHEN** `GET /folders/:id/share` returns 404 for the active folder
- **THEN** the switch is rendered in the off position
- **AND** no share link field is shown

#### Scenario: Folder has an active share

- **WHEN** `GET /folders/:id/share` returns 200 with a token for the active folder
- **THEN** the switch is rendered in the on position
- **AND** the share link field shows `<origin>/share/<token>` with a copy button

### Requirement: Enabling sharing

The system SHALL call `POST /folders/:id/share` when the user turns the switch on while sharing is off. On success, the switch SHALL move to the on position and the share link field SHALL appear showing `<origin>/share/<token>`.

#### Scenario: User enables sharing

- **WHEN** the user toggles the switch on while sharing is off
- **THEN** `POST /folders/:id/share` is called
- **AND** on success the switch moves to the on position and the share link field shows `<origin>/share/<token>` with a copy button

### Requirement: Disabling sharing requires confirmation

The system SHALL show a confirmation dialog when the user turns the switch off while sharing is on, warning that the existing link will stop working and that re-enabling will generate a new link. `DELETE /folders/:id/share` SHALL only be called, and the switch SHALL only move to the off position, if the user confirms. Cancelling SHALL leave the switch on, the share link field visible, and make no request.

#### Scenario: User confirms disabling sharing

- **WHEN** the user toggles the switch off while sharing is on
- **AND** confirms the dialog
- **THEN** `DELETE /folders/:id/share` is called
- **AND** the switch moves to the off position and the share link field is hidden

#### Scenario: User cancels disabling sharing

- **WHEN** the user toggles the switch off while sharing is on
- **AND** cancels the dialog
- **THEN** no request is made
- **AND** the switch remains on and the share link field remains visible

### Requirement: Copy share link to clipboard

The system SHALL provide a copy button next to the share link field that copies the full link (`<origin>/share/<token>`) to the clipboard and shows a success toast. If the clipboard write fails, an error toast SHALL be shown instead.

#### Scenario: User copies the share link

- **WHEN** sharing is enabled and the user clicks the copy button
- **THEN** the full share link is copied to the clipboard
- **AND** a success toast is shown

#### Scenario: Copy fails

- **WHEN** sharing is enabled and the clipboard write fails
- **THEN** an error toast is shown
