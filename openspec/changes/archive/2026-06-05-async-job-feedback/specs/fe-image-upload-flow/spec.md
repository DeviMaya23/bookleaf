## MODIFIED Requirements

### Requirement: Upload success flow (modal)

On successful upload via the upload modal, the system SHALL close the modal, show a success toast, and refresh the image list. The modal SHALL close immediately on success regardless of any async job state.

#### Scenario: Success closes modal and shows toast

- **WHEN** all 3 upload steps succeed via the upload modal
- **THEN** the modal closes
- **AND** a success toast is shown
- **AND** the image list is refreshed

---

### Requirement: Upload success flow (drag-and-drop)

The drag-and-drop file upload path (dropping a single file onto the main content area) uses the same 3-step upload sequence (`POST /images`, `PUT` to R2, `POST /images/:id/complete`). On success it SHALL refresh the image list and trigger the same post-upload feedback hook as the modal path. There is no modal to close.

#### Scenario: Single-file drop succeeds

- **WHEN** a single supported image file is dropped onto the main content area
- **AND** all 3 upload steps succeed
- **THEN** the image list is refreshed
- **AND** the post-upload feedback hook is triggered with the new image ID

---

## REMOVED Requirements

### Requirement: Folder suggestion view

**Reason**: Folder suggestions are now delivered asynchronously via a post-upload toast after the vision labelling job completes. The inline modal suggestion view required the vision API call to be synchronous with `CompleteUpload`, which is no longer the case.

**Migration**: Folder suggestion UX is handled by the `fe-async-job-feedback` capability — a single-check toast with Accept / Ignore actions replaces the inline modal view. The `POST /images/:id/accept-suggestion` endpoint is unchanged.
