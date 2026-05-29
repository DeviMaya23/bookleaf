## Why

The backend tagging system (tags table, image_tags join table, tag CRUD endpoints, and `PATCH /images/:id` tag replacement) shipped in the previous cycle. The right panel has no tag UI yet, leaving that backend work unused. This change wires the frontend to those existing endpoints.

## What Changes

- Add a `tags` field to the FE `Image` type (the backend already returns it)
- Add a `tags.ts` lib module with `getTags` and `createTag` API functions
- Add a `TagInput` component that renders existing tags as removable pills and accepts new tag names via keyboard input
- Update `RightPanel` to: load the image's current tags, show the `TagInput` section, auto-save on every add/remove via `PATCH /images/:id`
- Fetch all user tags on app load via `GET /tags` so the component can resolve names to IDs without a round-trip per keystroke

## Capabilities

### New Capabilities

- `fe-image-tagging`: TagInput component and tag management logic in the right panel, including the tags API lib

### Modified Capabilities

- `fe-right-panel`: Adding a Tags section requirement (TagInput between Source and Details, auto-saved on each change)

## Impact

- `frontend/src/lib/images.ts` — `Image` type gains `tags: { id: string; name: string }[]`
- `frontend/src/lib/tags.ts` — new file: `getTags`, `createTag`
- `frontend/src/components/TagInput.tsx` — new component
- `frontend/src/components/RightPanel.tsx` — adds tags state, `useQuery` for all tags, tag save logic
- `frontend/src/App.tsx` or query provider — `useQuery(['tags'])` hoisted so cache is shared
- No new routes; no changes to the extension or backend
