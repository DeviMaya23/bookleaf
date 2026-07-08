## MODIFIED Requirements

### Requirement: Batch save and aggregated feedback

When the user confirms the selection, the content script SHALL close the overlay and send `{ type: "picker-save", images: Array<{ srcUrl: string }> }` to the background. The background SHALL handle this message by:

1. Reading `getDefaultFolder()` from `browser.storage.local` once before fan-out.
2. Invoking `handleSave` for each image in silent mode (suppressing per-image toasts) using `Promise.allSettled`, passing the default folder `id` (if any) into each call.
3. Sending a single aggregated toast to the originating tab:
   - All fulfilled: `variant: "success"`, title `"Saved to Bookleaf."`, body `"X images added to [folder name]."` when a default folder is configured, or `"X images added to Unsorted."` when no default folder is configured.
   - Mixed results: `variant: "error"`, title `"Partially saved."`, body `"X saved, Y failed. Check your connection."`
   - All rejected: `variant: "error"`, title `"Couldn't save images."`, body `"Check your connection and try again."`

Each `handleSave` call SHALL use the tab's URL as `pageUrl` and the tab's title as `title`, the same defaults used by the drag-save flow.

#### Scenario: All images save successfully with no default folder

- **WHEN** the user confirms a selection of 4 images, all save without error, and no default folder is configured
- **THEN** a success toast is shown: `"4 images added to Unsorted."`

#### Scenario: All images save successfully with a default folder

- **WHEN** the user confirms a selection of 4 images, all save without error, and a default folder `{ id: "f1", name: "Inspo" }` is configured
- **THEN** each of the 4 `POST /images` calls includes `folder_id: "f1"`
- **AND** a success toast is shown: `"4 images added to Inspo."`

#### Scenario: Some images fail to save

- **WHEN** the user confirms a selection of 4 images and 1 fails
- **THEN** an error toast is shown: `"3 saved, 1 failed. Check your connection."`

#### Scenario: All images fail to save

- **WHEN** the user confirms a selection and every save call rejects
- **THEN** an error toast is shown with title `"Couldn't save images."` and body `"Check your connection and try again."`
