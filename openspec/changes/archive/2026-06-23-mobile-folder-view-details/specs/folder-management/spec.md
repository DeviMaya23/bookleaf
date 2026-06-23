## ADDED Requirements

### Requirement: View folder details via ContextMenu on coarse-pointer devices

The system SHALL show a "View details" item at the top of the folder `ContextMenu` (the same context menu used for "New subfolder", "Rename", "Change icon", and "Delete"), above "New subfolder", when the device is a coarse-pointer device (`useIsCoarsePointer()` is true). Selecting "View details" SHALL navigate to that folder (if it is not already the active folder, the same navigation a tap/click on the folder performs) and open the right panel for that folder as a bottom drawer, per `fe-right-panel`. This mirrors the fine-pointer behavior, where a single click both navigates to the folder and opens its panel — the right panel SHALL NEVER display metadata for a folder other than the currently active one, on either pointer type. The "View details" item SHALL NOT be rendered on fine-pointer devices, since selecting the folder already opens the right panel there.

#### Scenario: View details item shown on a coarse-pointer device

- **WHEN** a user on a coarse-pointer device long-presses a folder in the sidebar to open its context menu
- **THEN** the context menu shows "View details" above "New subfolder"

#### Scenario: View details item not shown on a fine-pointer device

- **WHEN** a user on a fine-pointer device right-clicks a folder in the sidebar to open its context menu
- **THEN** the context menu does not show a "View details" item

#### Scenario: Selecting View details on a non-active folder navigates to it and opens the right panel

- **WHEN** a user on a coarse-pointer device selects "View details" from the context menu of a folder that is not the currently active folder
- **THEN** the system navigates to that folder
- **AND** the right panel opens as a bottom drawer showing that folder's metadata, per `fe-right-panel`

#### Scenario: Selecting View details on the active folder opens the right panel without navigating

- **WHEN** a user on a coarse-pointer device selects "View details" from the context menu of the currently active folder
- **THEN** no navigation occurs
- **AND** the right panel opens as a bottom drawer showing that folder's metadata, per `fe-right-panel`
