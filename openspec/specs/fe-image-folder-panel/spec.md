## Purpose

Provides a reusable `FolderInput` component for managing image folder assignments via pill chips and a filtered suggestions dropdown.

## Requirements

### Requirement: FolderInput component

The system SHALL provide a `FolderInput` component at `frontend/src/components/FolderInput.tsx`.

Props:
```ts
interface FolderInputProps {
  folders: { id: string; name: string }[]
  onChange: (folders: { id: string; name: string }[]) => void
  disabled?: boolean
  suggestions?: { id: string; name: string }[]
}
```

Behaviour:
- Renders current folder assignments as removable pill chips
- An inline text input filters `suggestions` by name (case-insensitive substring match) when non-empty
- A filtered dropdown is shown below the container when the input is non-empty and there are matching suggestions not already selected
- Selecting a suggestion from the dropdown adds it to the list and clears the input
- Clicking the ✕ on a pill removes that folder from the list
- Pressing **Backspace** when the input is empty removes the last folder
- Blurring the input without selecting from the dropdown clears the input without committing anything (no raw-value creation)
- Pressing **Escape** closes the dropdown and clears the input
- Arrow keys navigate the dropdown; **Enter** selects the highlighted suggestion
- The component calls `onChange` with the updated list; it does not know about the API

#### Scenario: Existing folders render as pills

- **WHEN** the `folders` prop contains items
- **THEN** each folder is displayed as a pill with its `name` and a ✕ remove button

#### Scenario: Typing filters the suggestions dropdown

- **WHEN** the user types text in the folder input
- **THEN** suggestions matching the typed text (case-insensitive) and not already in `folders` are shown in a dropdown
- **AND** suggestions already in `folders` are excluded

#### Scenario: Selecting a suggestion adds the folder

- **WHEN** the user clicks a suggestion in the dropdown
- **THEN** `onChange` is called with that folder appended to the list
- **AND** the input is cleared

#### Scenario: Pressing Enter on a highlighted suggestion adds the folder

- **WHEN** the user navigates to a suggestion with arrow keys and presses Enter
- **THEN** `onChange` is called with that suggestion appended to the list
- **AND** the dropdown closes

#### Scenario: Blurring without selecting does not add anything

- **WHEN** the user types text and clicks away without selecting a suggestion
- **THEN** `onChange` is NOT called
- **AND** the input is cleared

#### Scenario: Backspace on empty input removes last folder

- **WHEN** the input is empty and the user presses Backspace
- **THEN** `onChange` is called with the last folder removed

#### Scenario: Removing a folder via its pill button

- **WHEN** the user clicks ✕ on a folder pill
- **THEN** `onChange` is called with that folder removed from the list

#### Scenario: Disabled state prevents interaction

- **WHEN** the `disabled` prop is true
- **THEN** the input is non-interactive and pills have no remove buttons
