## Why

The gallery has no drag-and-drop support: users cannot reorganize images or folders without navigating menus, and uploading an image by dragging a file onto the page is not possible. Adding these interactions removes friction from the core moodboarding workflow and brings the experience closer to tools like Eagle and Raindrop.

## What Changes

- Image cards in the gallery become draggable; users can drop them onto sidebar folders to move them
- Dropping an image onto "Unsorted" clears its folder assignment
- Folder items in the sidebar become draggable and droppable; users can nest folders by dragging or promote them to root by dropping on the empty space below the folder list
- Dropping a file from the OS onto the main content area auto-uploads it to the currently open folder, opens the right panel, and focuses the title field — no modal shown
- The upload modal is redesigned: selecting a file replaces the drop zone with an inline thumbnail preview; the title is pre-filled; a collapsible "Add details" section exposes Notes and Source URL fields

## Capabilities

### New Capabilities

- `fe-drag-drop-image-to-folder`: Drag an image card from the gallery and drop it onto a sidebar folder to move it; drop onto "Unsorted" to remove folder assignment
- `fe-drag-drop-folder-nesting`: Drag a folder item in the sidebar onto another folder to nest it, or onto the root drop zone below the list to promote it to root
- `fe-drag-drop-file-upload`: Drag a file from the OS onto the main content area to auto-upload without the modal; right panel opens on success with title auto-focused
- `fe-upload-modal-redesign`: Redesigned upload modal with inline thumbnail preview, pre-filled title, and collapsible "Add details" section for notes and source URL

### Modified Capabilities

- `fe-image-upload-flow`: `InitiateUploadParams` gains optional `description` and `source_url` fields; these are passed through to `POST /images` from the modal's "Add details" section

## Impact

- **New dependency**: `@dnd-kit/core` added to frontend
- **Frontend components**: `AppLayout`, `FolderSidebar`, `ImageGrid`, `UploadModal`
- **Frontend lib**: `images.ts` (`UpdateImageParams`, `InitiateUploadParams`), `folders.ts` (new `moveFolder` function)
- **Backend APIs used**: `PATCH /images/:id` (already accepts `folder_id`), `PUT /folders/:id` (already accepts `parent_id`), `GET /images/:id` (used to fetch full image after auto-upload)
- No backend changes required
