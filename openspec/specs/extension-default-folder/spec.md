# Spec: Extension Default Folder

## Purpose

Defines how the browser extension lets users configure a default save destination folder: how the chosen folder is persisted in extension storage, how the folder list is fetched from the API, how the Settings UI exposes the preference, and how the default folder is forwarded on both single-image and batch saves.

## Requirements

### Requirement: Default folder storage

The extension SHALL persist the user's chosen default folder as a `{ id: string; name: string }` pair in `browser.storage.local` under the key `bookleaf_default_folder`. When no folder is chosen, the key SHALL be absent (or `null`). `storage.ts` SHALL expose:
- `getDefaultFolder(): Promise<{ id: string; name: string } | null>`
- `setDefaultFolder(folder: { id: string; name: string }): Promise<void>`
- `clearDefaultFolder(): Promise<void>`

The `name` is stored alongside the `id` solely for use in success toast messages. If the folder is later renamed in the app, the stored name may be stale until the user reopens Settings and re-selects it; this is an accepted limitation.

#### Scenario: Default folder is stored and retrieved

- **WHEN** `setDefaultFolder({ id: "abc", name: "Design" })` is called
- **THEN** `getDefaultFolder()` returns `{ id: "abc", name: "Design" }`

#### Scenario: No default folder returns null

- **WHEN** no default folder has been stored
- **THEN** `getDefaultFolder()` returns `null`

#### Scenario: Clearing default folder removes it

- **WHEN** `clearDefaultFolder()` is called after a folder was stored
- **THEN** `getDefaultFolder()` returns `null`

### Requirement: Folder list fetch from API

`api.ts` SHALL expose a `getFolders()` function that calls `GET /folders` via `apiFetch` and returns a `Promise<Array<{ id: string; name: string; parent_id: string | null }>>`. On a non-OK response it SHALL throw an error.

#### Scenario: Successful fetch returns folder list with parent_id

- **WHEN** `GET /folders` responds with a 200 and a JSON array of folder objects
- **THEN** `getFolders()` resolves with an array of `{ id, name, parent_id }` objects

#### Scenario: Non-OK response throws

- **WHEN** `GET /folders` responds with a non-2xx status
- **THEN** `getFolders()` rejects with an error

### Requirement: Save destination section in Settings UI

The Settings panel (`Settings.tsx`) SHALL render a **"Save destination"** section between the drag-to-save row and the Hotkeys section header. The section SHALL contain:

- A section label styled identically to the existing "Hotkeys" label (10px, uppercase, `c.textSec`).
- A row with a text label "Default folder" and a button flush-right that displays the current default folder name, or `"None"` when no default is stored.
- Clicking the button SHALL navigate to the FolderPicker panel view.
- The displayed name SHALL be read from `getDefaultFolder()` on Settings mount and updated when the user returns from the FolderPicker panel.

#### Scenario: Settings shows stored folder name on the button

- **WHEN** Settings mounts and a default folder `{ id: "xyz", name: "Mood" }` is stored
- **THEN** the "Default folder" button reads "Mood"

#### Scenario: Settings shows "None" when no default is stored

- **WHEN** Settings mounts and no default folder is stored
- **THEN** the "Default folder" button reads "None"

#### Scenario: Clicking the button navigates to FolderPicker

- **WHEN** the user clicks the "Default folder" button in Settings
- **THEN** the popup navigates to the FolderPicker panel view

#### Scenario: Returning from FolderPicker reflects the updated selection

- **WHEN** the user selects a folder in FolderPicker and the popup navigates back to Settings
- **THEN** the "Default folder" button displays the name of the newly selected folder

### Requirement: Default folder forwarded on single save

All single-image save paths (`persistImage`) SHALL read `getDefaultFolder()` from storage before constructing the `POST /images` request. If a default folder is configured, its `id` SHALL be sent as `folder_id` in the request body. If no default is configured, `folder_id` SHALL be omitted (behaviour unchanged from before).

The success toast body SHALL reflect the destination:
- No default folder configured: `"Added to Unsorted."`
- Default folder configured: `"Added to [folder name]."`

#### Scenario: Single save with default folder sends folder_id

- **WHEN** a default folder `{ id: "f1", name: "Inspo" }` is stored and the user saves an image
- **THEN** `POST /images` is sent with `folder_id: "f1"` in the request body
- **AND** the success toast body reads `"Added to Inspo."`

#### Scenario: Single save with no default folder omits folder_id

- **WHEN** no default folder is stored and the user saves an image
- **THEN** `POST /images` is sent without a `folder_id` field
- **AND** the success toast body reads `"Added to Unsorted."`

#### Scenario: Stale or deleted folder_id is handled by the backend

- **WHEN** the stored default folder no longer exists in the backend
- **THEN** the extension still sends the stored `folder_id`; the backend silently falls back to Unsorted; no error is shown to the user

### Requirement: Default folder forwarded on batch save

`handlePickerSaveMessage` SHALL read `getDefaultFolder()` once before fan-out and pass the resulting `id` (if any) into each `handleSave` call. The aggregated success toast body SHALL reflect the destination:
- No default folder configured: `"X images added to Unsorted."`
- Default folder configured: `"X images added to [folder name]."`

#### Scenario: Batch save with default folder sends folder_id on each save

- **WHEN** a default folder `{ id: "f1", name: "Inspo" }` is stored and the user confirms a batch of 3 images
- **THEN** each of the 3 `POST /images` calls includes `folder_id: "f1"`
- **AND** the aggregated success toast body reads `"3 images added to Inspo."`

#### Scenario: Batch save with no default folder omits folder_id

- **WHEN** no default folder is stored and the user confirms a batch of 3 images
- **THEN** each `POST /images` call omits `folder_id`
- **AND** the aggregated success toast body reads `"3 images added to Unsorted."`
