## MODIFIED Requirements

### Requirement: TagInput component

The system SHALL provide a `TagInput` component at `frontend/src/components/TagInput.tsx`.

Props:
```ts
interface TagInputProps {
  tags: { id: string; name: string }[]
  onChange: (tags: { id: string; name: string }[]) => void
  disabled?: boolean
  suggestions?: { id: string; name: string }[]
}
```

Behaviour:
- Renders current tags as removable pill chips
- An inline text input allows typing a new tag name
- Pressing **Enter** or **comma** commits the typed value (trimmed, lowercased), unless a suggestion is highlighted — in that case Enter commits the highlighted suggestion
- Pressing **Backspace** when the input is empty removes the last tag
- Blurring the input with a non-empty value commits it
- The component calls `onChange` with the updated list; it does not know about the API
- When `suggestions` is provided and the input is non-empty, a filtered dropdown is shown below the container (see `fe-tag-suggestions` spec)

#### Scenario: Existing tags render as pills

- **WHEN** `tags` prop contains items
- **THEN** each tag is displayed as a pill with its `name` and a remove button

#### Scenario: Pressing Enter commits a new tag

- **WHEN** the user types a name and presses Enter
- **AND** no suggestion is highlighted
- **THEN** `onChange` is called with the new tag appended

#### Scenario: Pressing comma commits a new tag

- **WHEN** the user types a name and presses comma
- **THEN** `onChange` is called with the new tag appended and the comma is not included in the name

#### Scenario: Pressing Backspace on empty input removes last tag

- **WHEN** the input is empty and the user presses Backspace
- **THEN** `onChange` is called with the last tag removed

#### Scenario: Removing a tag via its remove button

- **WHEN** the user clicks the ✕ on a tag pill
- **THEN** `onChange` is called with that tag removed from the list

#### Scenario: Blurring with non-empty input commits the value

- **WHEN** the user types a name and clicks away without pressing Enter
- **THEN** `onChange` is called with the typed name appended
