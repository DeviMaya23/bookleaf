## Why

The tag input currently requires users to remember and type their existing tag names exactly. A filtered suggestion dropdown lets users discover and reuse tags faster and prevents near-duplicate tags (e.g. "nature" vs "natural"). The `allTags` cache already exists in `RightPanel` from the previous cycle, so this is purely a UI enhancement with no new API calls.

## What Changes

- `TagInput` gains an optional `suggestions` prop (`{ id: string; name: string }[]`) — filtered and passed by the parent
- While the user is typing, a dropdown appears below the input showing matching suggestions (prefix or substring match, excluding already-applied tags)
- Clicking a suggestion or pressing arrow keys + Enter commits that tag
- `RightPanel` computes the filtered suggestion list from `allTags` and passes it to `TagInput`

## Capabilities

### New Capabilities

- `fe-tag-suggestions`: Autocomplete suggestion dropdown in TagInput

### Modified Capabilities

- `fe-image-tagging`: TagInput now accepts a `suggestions` prop; RightPanel passes filtered suggestions

## Impact

- `frontend/src/components/TagInput.tsx` — add `suggestions` prop, dropdown UI, keyboard navigation
- `frontend/src/components/RightPanel.tsx` — compute and pass filtered suggestions to `TagInput`
- No backend changes, no new API calls
