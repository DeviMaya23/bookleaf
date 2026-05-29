## 1. Tags API Lib

- [x] 1.1 Create `frontend/src/lib/tags.ts` with `Tag` interface, `getTags`, and `createTag` functions
- [x] 1.2 Add `tags` field (`{ id: string; name: string }[]`) to the `Image` interface in `frontend/src/lib/images.ts`
- [x] 1.3 Update `UpdateImageParams` in `frontend/src/lib/images.ts` to include `tags?: string[]` (UUID array)

## 2. TagInput Component

- [x] 2.1 Create `frontend/src/components/TagInput.tsx` with pill rendering, Enter/comma/Backspace/blur commit behaviour
- [x] 2.2 Write unit tests for `TagInput` — success scenario (adds a tag on Enter) and failure scenario (does not add duplicate name)

## 3. RightPanel Integration

- [x] 3.1 Add `useQuery(['tags'])` in `RightPanel` using `getTags`, `staleTime: 60_000`
- [x] 3.2 Add local tag state (`{ id: string; name: string }[]`) initialised from `image.tags`, reset on `image.id` change
- [x] 3.3 Implement `handleTagsChange` callback: resolve name→ID from cache (or call `createTag` on miss), then call `PATCH /images/:id` with full UUID array
- [x] 3.4 Render the Tags section between Source and Details using `<TagInput>`
- [x] 3.5 Write unit tests for `RightPanel` tag behaviour — success scenario (add existing tag fires PATCH with correct IDs) and failure scenario (PATCH failure shows error toast)
