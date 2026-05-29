## 1. TagInput — suggestions prop and dropdown UI

- [x] 1.1 Add `suggestions?: { id: string; name: string }[]` prop to `TagInput` interface
- [x] 1.2 Add `selectedIndex` state (number, -1 = none) for keyboard navigation
- [x] 1.3 Render the suggestion dropdown (absolutely positioned below the container) when `val` is non-empty and filtered suggestions exist
- [x] 1.4 Apply substring filter on `suggestions` using internal `val` state before rendering

## 2. TagInput — interaction behaviour

- [x] 2.1 Clicking a suggestion calls `onChange` with that tag appended (using its real ID) and closes the dropdown
- [x] 2.2 ↓/↑ arrow keys increment/decrement `selectedIndex`, wrapping at boundaries
- [x] 2.3 Enter with `selectedIndex >= 0` commits the highlighted suggestion (bypasses raw commit)
- [x] 2.4 Escape key resets `selectedIndex` to -1 and hides the dropdown
- [x] 2.5 On blur, hide dropdown after 150ms delay to allow click events on suggestions to fire first
- [x] 2.6 Reset `selectedIndex` to -1 whenever `val` changes

## 3. RightPanel — pass suggestions prop

- [x] 3.1 Compute `suggestionsForInput` from `allTags` excluding tags already in `tags` state, and pass to `<TagInput suggestions={suggestionsForInput} />`

## 4. Tests

- [x] 4.1 Unit test `TagInput` — success: clicking a suggestion calls `onChange` with the suggestion's ID (not a blank placeholder)
- [x] 4.2 Unit test `TagInput` — success: pressing ↓ then Enter commits the first suggestion
- [x] 4.3 Unit test `TagInput` — failure: already-applied tags do not appear as suggestions (filtered out before render)
