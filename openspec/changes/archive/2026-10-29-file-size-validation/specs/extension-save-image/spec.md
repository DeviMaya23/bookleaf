# Spec Delta

## ADDED Requirements

### Requirement: Extension rejects oversized blobs before upload
Before initiating the upload sequence, the extension SHALL check the resolved image blob's byte length. If it exceeds 52,428,800 bytes (50 MiB), the save SHALL be aborted and an error toast SHALL be sent to the active tab. No API call SHALL be made.

#### Scenario: Oversized blob is rejected with a toast
- **WHEN** the resolved image blob is larger than 52,428,800 bytes
- **THEN** no `POST /images` request is made
- **AND** an error toast is shown in the active tab

#### Scenario: Blob within the limit proceeds normally
- **WHEN** the resolved image blob is at or below 52,428,800 bytes
- **THEN** the upload sequence continues as normal

---

### Requirement: Extension sends file_size in the initiate request
The extension's `POST /images` initiate call SHALL include a `file_size` field in the request body equal to the byte length of the image blob being uploaded.

#### Scenario: file_size is included in POST /images body
- **WHEN** the extension initiates an upload
- **THEN** the `POST /images` request body contains `file_size` equal to the blob's byte length
