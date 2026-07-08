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

`api.ts` SHALL expose a `getFolders()` function that calls `GET /folders` via `apiFetch` and returns a `Promise<Array<{ id: string; name: string }>>`. On a non-OK response it SHALL throw an error.

#### Scenario: Successful fetch returns folder list

- **WHEN** `GET /folders` responds with a 200 and a JSON array of folder objects
- **THEN** `getFolders()` resolves with an array of `{ id, name }` objects

#### Scenario: Non-OK response throws

- **WHEN** `GET /folders` responds with a non-2xx status
- **THEN** `getFolders()` rejects with an error

### Requirement: Save destination section in Settings UI

The Settings panel (`Settings.tsx`) SHALL render a **"Save destination"** section between the drag-to-save row and the Hotkeys section header. The section SHALL contain:

- A section label styled identically to the existing "Hotkeys" label (10px, uppercase, `c.textSec`).
- A row with a text label "Default folder" and a native `<select>` element flush-right.
- The `<select>` SHALL be populated with:
  - A first option with value `""` and label `"None (Unsorted)"`.
  - One `<option>` per folder returned by `GET /folders`, using `folder.id` as value and `folder.name` as label, sorted by the API response order.
- The `<select>` SHALL be initialized to the stored default folder ID on mount, or `""` if none is set.
- Changing the `<select>` SHALL immediately call `setDefaultFolder({ id, name })` for a non-empty selection, or `clearDefaultFolder()` for `""`.
- `GET /folders` SHALL be called inside a `useEffect` on Settings mount. While loading, the `<select>` SHALL render with only the `"None (Unsorted)"` option and be disabled. On fetch error, the `<select>` SHALL remain with only `"None (Unsorted)"` and be re-enabled (the stored preference is preserved).

#### Scenario: Settings loads folder list and pre-selects stored default

- **WHEN** Settings mounts and a default folder is stored with id `"xyz"` and the API returns a list containing `{ id: "xyz", name: "Mood" }`
- **THEN** the `<select>` shows "Mood" as the selected option

#### Scenario: Settings shows "None (Unsorted)" when no default is stored

- **WHEN** Settings mounts and no default folder is stored
- **THEN** the `<select>` shows "None (Unsorted)" as the selected option

#### Scenario: Selecting a folder persists it to storage

- **WHEN** the user changes the `<select>` to a folder option
- **THEN** `setDefaultFolder({ id, name })` is called with that folder's id and name

#### Scenario: Selecting "None (Unsorted)" clears the stored default

- **WHEN** the user changes the `<select>` back to "None (Unsorted)"
- **THEN** `clearDefaultFolder()` is called

#### Scenario: Dropdown is disabled while folder list is loading

- **WHEN** Settings has mounted but `GET /folders` has not yet resolved
- **THEN** the `<select>` is disabled and shows only "None (Unsorted)"

#### Scenario: Fetch error leaves dropdown functional with no folder options

- **WHEN** `GET /folders` rejects
- **THEN** the `<select>` is re-enabled with only "None (Unsorted)" and the previously stored folder preference is unchanged

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
