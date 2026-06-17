## ADDED Requirements

### Requirement: Folder icon allowlist

The system SHALL define a fixed allowlist of 55 valid icon keys, used both to validate folder icon updates server-side and to populate the icon picker client-side. The allowlist (kebab-case keys, corresponding to lucide-react icon components): `bookmark`, `paperclip`, `folder`, `folder-bookmark`, `folder-closed`, `folder-open`, `folders`, `file-stack`, `file-question-mark`, `file-image`, `book-image`, `image`, `images`, `gpu`, `mirror-rectangular`, `sun`, `moon`, `cloud`, `cloud-fog`, `cloud-drizzle`, `cloud-sun`, `cloudy`, `apple`, `coffee`, `cookie`, `chef-hat`, `sandwich`, `bottle-wine`, `clover`, `club`, `crown`, `gem`, `gift`, `headphones`, `rocket`, `star`, `ghost`, `house`, `heart`, `flower`, `leaf`, `sprout`, `trees`, `map-pin`, `utensils`, `ship-wheel`, `bell`, `alarm-clock`, `album`, `flask-conical`, `snowflake`, `cylinder`, `mail`, `palette`, `trash-2`. The backend list is authoritative; the frontend maintains its own icon-key-to-component mapping covering the same keys.

#### Scenario: Backend validates icon key against allowlist

- **WHEN** a folder update request includes an `icon` value not present in the allowlist
- **THEN** the request is rejected with `400 Bad Request`
- **AND** the folder's `icon` column is not modified

#### Scenario: Backend accepts an allowlisted icon key

- **WHEN** a folder update request includes an `icon` value present in the allowlist
- **THEN** the folder's `icon` column is updated to that value

#### Scenario: Frontend renders an unrecognized stored icon key as the default

- **WHEN** the frontend encounters a folder whose `icon` value has no entry in its icon-key-to-component map
- **THEN** it renders the default `folder` icon instead of erroring

---

### Requirement: Change icon via folder context menu

The right-click context menu for a user-owned folder SHALL include a "Change icon" item that opens a submenu listing all allowlisted icons. Selecting an icon SHALL update the folder's `icon` via `PATCH /folders/:id` and immediately reflect the new icon in the sidebar.

#### Scenario: User changes a folder's icon

- **WHEN** the user right-clicks a folder, selects "Change icon", and picks an icon from the submenu
- **THEN** `PATCH /folders/:id` is called with the selected icon key
- **AND** the folder's icon in the sidebar updates to the selected icon on success

#### Scenario: Submenu lists all allowlisted icons

- **WHEN** the user opens the "Change icon" submenu
- **THEN** all 55 allowlisted icons are displayed as selectable options

---

### Requirement: Default folder icon

A user folder with no `icon` set (`null`) SHALL display the default `folder` icon.

#### Scenario: New folder shows the default icon

- **WHEN** a folder is created without specifying an icon
- **THEN** the sidebar renders the `folder` icon next to its name

---

### Requirement: Fixed icons for system entries

The "All", "Unsorted", and "Trash" sidebar entries SHALL each display a fixed icon that is not user-editable: "All" uses `file-stack`, "Unsorted" uses `file-question-mark`, and "Trash" uses `trash-2`. These entries SHALL NOT offer a "Change icon" option.

#### Scenario: System entries show their fixed icons

- **WHEN** the sidebar is rendered
- **THEN** "All" displays the `file-stack` icon, "Unsorted" displays the `file-question-mark` icon, and "Trash" displays the `trash-2` icon

#### Scenario: System entries have no icon-change option

- **WHEN** the user right-clicks the "All", "Unsorted", or "Trash" entry (where a context menu exists)
- **THEN** no "Change icon" option is present

---

### Requirement: Folder icons visibility toggle

The system SHALL provide a per-user `folder_icons_enabled` preference, defaulting to `true`, that controls whether any folder or system-entry icon is rendered in the sidebar. The toggle SHALL be exposed as a `Switch` in the SettingsModal's App section, persisted via `PATCH /me`.

#### Scenario: Icons render when the preference is enabled

- **WHEN** `folder_icons_enabled` is `true`
- **THEN** the sidebar renders an icon next to every folder and system entry

#### Scenario: Icons are omitted when the preference is disabled

- **WHEN** `folder_icons_enabled` is `false`
- **THEN** no icon `<span>` is rendered next to any folder or system entry
- **AND** the entry's name shifts left to fill the space, with no layout gap left behind

#### Scenario: Toggling the preference persists immediately

- **WHEN** the user toggles the "Folder icons" switch in the App settings section
- **THEN** `PATCH /me` is called with the new `folder_icons_enabled` value
- **AND** the switch is disabled while the request is in flight
