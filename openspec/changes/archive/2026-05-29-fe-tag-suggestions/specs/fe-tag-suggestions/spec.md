## Purpose

Defines the autocomplete suggestion dropdown capability for `TagInput`: filtered suggestion list, mouse and keyboard interaction, and the filtering logic in `RightPanel`.

## ADDED Requirements

### Requirement: TagInput renders a suggestion dropdown while typing

The system SHALL render a dropdown list of suggestions below the tag input container whenever the current input value is non-empty and at least one suggestion is available.

The `TagInput` component SHALL accept an optional `suggestions` prop:

```ts
suggestions?: { id: string; name: string }[]
```

The dropdown SHALL:
- Appear positioned directly below the input container
- Display each suggestion as a selectable row showing the tag name
- Hide when the input is cleared, when a suggestion is selected, when Escape is pressed, or when focus leaves the input area

#### Scenario: Dropdown appears when typing matches a suggestion

- **WHEN** the user types a value that matches one or more suggestions
- **THEN** a dropdown list appears below the input showing the matching tag names

#### Scenario: Dropdown is hidden when input is empty

- **WHEN** the input field is empty
- **THEN** no dropdown is shown, even if suggestions are available

#### Scenario: Dropdown is hidden when no suggestions match

- **WHEN** the user types a value that matches none of the suggestions
- **THEN** the dropdown is not shown

#### Scenario: Clicking a suggestion commits the tag

- **WHEN** the user clicks a suggestion in the dropdown
- **THEN** `onChange` is called with that tag appended to the current list
- **AND** the dropdown closes
- **AND** the input field is cleared

#### Scenario: Pressing Escape dismisses the dropdown

- **WHEN** the dropdown is visible and the user presses Escape
- **THEN** the dropdown closes without committing any tag
- **AND** the typed input value is preserved

---

### Requirement: TagInput supports keyboard navigation in the dropdown

The system SHALL support arrow-key navigation within the suggestion dropdown.

Behaviour:
- Pressing **↓** moves the highlight to the next suggestion (wraps to first from last)
- Pressing **↑** moves the highlight to the previous suggestion (wraps to last from first)
- Pressing **Enter** when a suggestion is highlighted commits that suggestion (not the raw typed value)
- Pressing **Enter** when no suggestion is highlighted commits the raw typed value (existing behaviour preserved)

#### Scenario: Arrow down highlights the first suggestion

- **WHEN** the dropdown is visible and the user presses ↓
- **THEN** the first suggestion is highlighted

#### Scenario: Enter commits the highlighted suggestion

- **WHEN** a suggestion is highlighted and the user presses Enter
- **THEN** `onChange` is called with that suggestion's tag (using its existing ID, not a blank placeholder)
- **AND** the dropdown closes

#### Scenario: Enter without highlight falls through to raw commit

- **WHEN** the dropdown is visible but no suggestion is highlighted and the user presses Enter
- **THEN** `onChange` is called with the raw typed value (existing commit behaviour)

---

### Requirement: RightPanel passes filtered suggestions to TagInput

`RightPanel` SHALL compute a filtered suggestion list from `allTags` and pass it to `TagInput` via the `suggestions` prop.

Filtering rules:
- Include only tags whose `name` contains the current input value as a substring (case-insensitive)
- Exclude tags already present in the current `tags` state (by `id`)
- The filtered list is recomputed whenever `allTags`, `tags`, or the current input changes

Because `val` (the current input string) lives inside `TagInput`, `RightPanel` cannot filter by it directly. The `TagInput` component itself SHALL apply the substring filter on the `suggestions` prop using its internal `val` state before rendering the dropdown.

#### Scenario: Suggestions exclude already-applied tags

- **GIVEN** the image has tag "nature" already applied
- **AND** "nature" exists in `allTags`
- **WHEN** the user types "nat"
- **THEN** "nature" does not appear in the suggestion dropdown

#### Scenario: RightPanel passes all user tags as suggestions

- **WHEN** `RightPanel` renders with a non-empty `allTags` cache
- **THEN** `TagInput` receives all user tags (minus already-applied ones) via the `suggestions` prop
