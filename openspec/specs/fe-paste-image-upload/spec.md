## Purpose

TBD

## Requirements

### Requirement: Global paste listener detects clipboard images
The app SHALL listen for `paste` events on the `document`. When a paste event fires and the clipboard contains an image file, the system SHALL open the single-file upload modal with the image pre-staged. The listener SHALL be active whenever the authenticated app shell is mounted.

#### Scenario: Pasting an image opens the upload modal
- **WHEN** the user presses CTRL+V (or CMD+V on macOS) while an image is in the clipboard
- **AND** no text input or textarea is focused
- **AND** the upload modal is not already open
- **THEN** the single-file upload modal opens
- **AND** the pasted image is pre-staged in the modal's drop zone

#### Scenario: Paste is ignored when a text input is focused
- **WHEN** the user presses CTRL+V while a text `<input>` or `<textarea>` is focused
- **THEN** normal browser text paste behaviour occurs
- **AND** the upload modal does not open

#### Scenario: Paste is ignored when clipboard contains no image
- **WHEN** the user presses CTRL+V while the clipboard contains only text or non-image data
- **THEN** no upload modal opens
- **AND** no error or toast is shown

#### Scenario: Paste is ignored when the upload modal is already open
- **WHEN** the user presses CTRL+V while the upload modal is already open
- **THEN** the modal state is unchanged
- **AND** no second modal or action is triggered

---

### Requirement: Clipboard image is extracted from paste event items
The system SHALL extract the image by iterating `event.clipboardData.items` and selecting the first item where `kind === 'file'` and `type` starts with `'image/'`. The synchronous `item.getAsFile()` method SHALL be used. The async Clipboard API (`navigator.clipboard.read`) SHALL NOT be used.

#### Scenario: First image item in clipboard data is used
- **WHEN** the clipboard contains an image file item
- **THEN** `item.getAsFile()` is called on the first matching item
- **AND** the resulting `File` object is passed to the upload modal
