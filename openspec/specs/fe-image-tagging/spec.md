## Purpose

Defines the frontend tag management capability: the tags API lib, the `TagInput` component, and the integration of tag state into the right panel.

## Requirements

### Requirement: Tags API lib

The system SHALL provide a `frontend/src/lib/tags.ts` module with two functions:

- `getTags(getToken)` — calls `GET /tags` and returns `Tag[]`
- `createTag(getToken, name)` — calls `POST /tags` with `{ name }` and returns the created `Tag`

```ts
export interface Tag {
  id: string
  name: string
}
```

`createTag` SHALL throw on any non-2xx response except 409. On a 409 response, the caller is responsible for re-fetching tags and resolving the ID (the tag already exists for this user).

#### Scenario: getTags returns user's tags

- **WHEN** `GET /tags` responds with a list of tags
- **THEN** `getTags` returns a `Tag[]` with `id` and `name` for each entry

#### Scenario: createTag returns new tag on success

- **WHEN** `POST /tags` responds with 201
- **THEN** `createTag` returns the created `Tag` object with a valid UUID `id`

#### Scenario: createTag throws on server error

- **WHEN** `POST /tags` responds with 500
- **THEN** `createTag` throws an error

---

### Requirement: Image type includes tags

The system SHALL add a `tags` field to the `Image` interface in `frontend/src/lib/images.ts`.

```ts
tags: { id: string; name: string }[]
```

The field SHALL be typed as a non-optional array (the backend guarantees it is never null).

#### Scenario: Image type compiles with tags field

- **WHEN** the TypeScript project is compiled
- **THEN** `Image.tags` is a required `{ id: string; name: string }[]` field with no type errors

---

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

---

### Requirement: RightPanel fetches all user tags

The system SHALL use `useQuery(['tags'], () => getTags(getToken))` inside `RightPanel` to load all user tags, with `staleTime: 60_000`.

This query populates the name→ID lookup used during tag commit.

#### Scenario: User tags are loaded when the right panel mounts

- **WHEN** the right panel mounts
- **THEN** `GET /tags` is called and the result is cached under key `['tags']`

#### Scenario: Cached tags are reused within stale window

- **WHEN** the right panel is re-opened within 60 seconds of the last fetch
- **THEN** no new `GET /tags` request is made

---

### Requirement: RightPanel initialises tag state from image

The system SHALL maintain local tag state as `{ id: string; name: string }[]` in `RightPanel`, initialised from `image.tags` and reset whenever `image.id` changes.

#### Scenario: Tag state is seeded from image.tags on open

- **WHEN** the right panel opens for an image with existing tags
- **THEN** the `TagInput` renders those tags as pills

#### Scenario: Tag state resets when a different image is selected

- **WHEN** the user selects a different image while the panel is already open
- **THEN** the `TagInput` reflects only the new image's tags

---

### Requirement: Tag add fires PATCH with full tag set

When the user commits a tag name in `TagInput`, `RightPanel` SHALL:

1. Search `allTags` cache for a tag with that name (case-insensitive)
2. If found — reuse its ID
3. If not found — call `createTag(getToken, name)` to get a new ID, then call `queryClient.setQueryData(['tags'], ...)` to add it to the cache
4. Append the resolved `{ id, name }` to local tag state
5. Call `PATCH /images/:id` with `{ tags: <full updated UUID array> }`
6. On success — invalidate `['images']` and show a success toast
7. On error — show an error toast; do not revert local state (the save failed, panel still shows the intent)

#### Scenario: Adding an existing tag reuses its ID

- **GIVEN** the user's tag list contains `{ id: "abc", name: "nature" }`
- **WHEN** the user types "nature" and presses Enter
- **THEN** no `POST /tags` request is made
- **AND** `PATCH /images/:id` is called with `"abc"` in the tags array

#### Scenario: Adding a new tag creates it then patches the image

- **GIVEN** no tag named "concept" exists in the user's tag list
- **WHEN** the user types "concept" and presses Enter
- **THEN** `POST /tags` is called with `{ "name": "concept" }`
- **AND** `PATCH /images/:id` is called with the new tag's UUID in the tags array

#### Scenario: Successful tag add shows a toast

- **WHEN** `PATCH /images/:id` succeeds after a tag add
- **THEN** a success toast is shown

#### Scenario: Failed PATCH shows error toast

- **WHEN** `PATCH /images/:id` fails
- **THEN** an error toast is shown

---

### Requirement: Tag remove fires PATCH with full tag set

When the user removes a tag pill in `TagInput`, `RightPanel` SHALL remove that tag from local state and call `PATCH /images/:id` with `{ tags: <remaining UUID array> }`.

#### Scenario: Removing a tag patches the image

- **WHEN** the user clicks ✕ on a tag pill
- **THEN** `PATCH /images/:id` is called with that tag's UUID excluded from the array

#### Scenario: Removing the last tag sends an empty array

- **WHEN** the user removes the only remaining tag
- **THEN** `PATCH /images/:id` is called with `{ "tags": [] }`
