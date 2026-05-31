## MODIFIED Requirements

### Requirement: Dropping image onto a folder moves it

When an image drag ends over a folder drop target, `AppLayout`'s `onDragEnd` handler SHALL call `POST /images/:id/move-folder` with `{ from_folder_id: currentFolderId, to_folder_id: targetFolderId }`. On success the image queries SHALL be invalidated. On error a toast SHALL be shown.

#### Scenario: Image dropped onto a folder is moved into that folder

- **WHEN** the user drops an image card onto a sidebar folder item
- **THEN** `POST /images/:imageId/move-folder` is called with `{ "from_folder_id": "<currentFolderId>", "to_folder_id": "<targetFolderId>" }`
- **AND** the image list is refreshed
- **AND** a success toast is shown

#### Scenario: Dropping image onto its current folder is a no-op

- **WHEN** the user drops an image onto the folder it is already in (currentFolderId == targetFolderId)
- **THEN** no request is made

#### Scenario: Move failure shows error toast

- **WHEN** the POST request fails
- **THEN** an error toast is shown
- **AND** the image list is refreshed to restore correct state

---

### Requirement: "Unsorted" system entry is a drop target that clears folder assignment

The "Unsorted" sidebar entry SHALL be a `useDroppable` target carrying `{ type: 'unsorted' }`. Dropping an image onto it SHALL call `POST /images/:id/move-folder` with `{ from_folder_id: currentFolderId, to_folder_id: null }`.

#### Scenario: Dropping image onto Unsorted removes its folder assignment

- **WHEN** the user drops an image card onto the "Unsorted" sidebar entry
- **THEN** `POST /images/:imageId/move-folder` is called with `{ "from_folder_id": "<currentFolderId>", "to_folder_id": null }`
- **AND** the image list is refreshed
- **AND** a success toast is shown

#### Scenario: Dropping an already-unsorted image onto Unsorted is a no-op

- **WHEN** the user drops an image that has no folder (currentFolderId == null) onto "Unsorted"
- **THEN** no request is made

---

### Requirement: moveImageFolder API function

The system SHALL provide a `moveImageFolder(getToken, imageId, fromFolderId, toFolderId)` function in `frontend/src/lib/images.ts` that calls `POST /images/:imageId/move-folder`.

```ts
export async function moveImageFolder(
  getToken: () => Promise<string | null>,
  imageId: string,
  fromFolderId: string | null,
  toFolderId: string | null,
): Promise<void>
```

The function SHALL throw on any non-2xx response.

#### Scenario: moveImageFolder calls correct endpoint

- **WHEN** `moveImageFolder` is called with a valid imageId, fromFolderId, and toFolderId
- **THEN** a `POST` request is made to `/images/:imageId/move-folder` with the correct JSON body
