## 1. Dependencies & Lib Updates

- [x] 1.1 Install `@dnd-kit/core` in the frontend package
- [x] 1.2 Add `folder_id?: string | null` to `UpdateImageParams` in `images.ts`
- [x] 1.3 Add `description?: string | null` and `source_url?: string | null` to `InitiateUploadParams` in `images.ts` and pass them through in `initiateUpload`
- [x] 1.4 Add `moveFolder(getToken, id, name, parentId)` function to `folders.ts` calling `PUT /folders/:id` with `{ name, parent_id }`

## 2. Move Image to Folder (D&D)

- [x] 2.1 Wrap `ImageCard` with `useDraggable` carrying `{ type: 'image', imageId, currentFolderId }` with 8px activation constraint; dim card to opacity 0.4 while dragging
- [x] 2.2 Wrap `FolderItem` in `FolderSidebar` with `useDroppable` carrying `{ type: 'folder', folderId }`; apply accent highlight when `isOver` and active drag type is `'image'`
- [x] 2.3 Make the "Unsorted" `SystemEntry` a `useDroppable` carrying `{ type: 'unsorted' }`; apply accent highlight when `isOver` and active drag type is `'image'`
- [x] 2.4 Add `DndContext` and `DragOverlay` to `AppLayout`; render a thumbnail mini-card (80×80px, cover) in the overlay while an image drag is active
- [x] 2.5 Implement `onDragEnd` in `AppLayout` for image drops: skip no-op (same folder), call `PATCH /images/:id { folder_id }` via `updateImage`, invalidate image queries, show error toast on failure
- [x] 2.6 Write unit tests for the image-move handler: success scenario (folder_id updated) and failure scenario (error toast shown)

## 3. Move Folder (D&D)

- [x] 3.1 Add `useDraggable` to `FolderItem` carrying `{ type: 'folder', folderId, name, parentId }` with 8px activation constraint
- [x] 3.2 Add circular-drop guard helper: given a flat folder list and a dragged folder ID, return the full set of descendant IDs (including self); used to block invalid drops
- [x] 3.3 Update `FolderItem`'s `useDroppable` to highlight only when active drag type is `'folder'` and the target is not in the dragged folder's subtree
- [x] 3.4 Add root drop zone below the folder list in `FolderSidebar`: `useDroppable` with `{ type: 'root' }`, visible only while a folder drag is active (show dashed border + "Move to root" label)
- [x] 3.5 Extend `onDragEnd` in `AppLayout` for folder drops: run circular guard, skip no-ops, call `moveFolder` for folder-on-folder and folder-on-root, invalidate folder queries, show error toast on failure
- [x] 3.6 Write unit tests for the circular-drop guard helper: returns correct subtree, blocks self-drop, blocks descendant-drop
- [x] 3.7 Write unit tests for the folder-move handler: success (parent_id updated), no-op (same parent), circular guard blocks call

## 4. OS File Auto-Upload

- [x] 4.1 Add `dragover` and `dragleave` listeners on the `<main>` element in `AppLayout`; set `isFileDragOver` state only when `dataTransfer.types.includes('Files')`
- [x] 4.2 Render a full-page overlay inside `<main>` when `isFileDragOver` is true; overlay shows "Drop to upload" text and dismisses on `dragleave` or after drop
- [x] 4.3 Add `drop` listener on `<main>`; validate MIME type, show error toast on invalid type, otherwise run the 3-step upload sequence using current `folderId`
- [x] 4.4 After successful auto-upload: invalidate image queries, call `getImage(getToken, image_id)` to fetch the full image, call `setSelectedImage` to open the right panel
- [x] 4.5 Auto-focus the title input in `RightPanel` when the panel opens for a newly uploaded image; add a `titleInputRef` and a `useEffect` that focuses it when `image.id` changes
- [x] 4.6 Write unit tests for the auto-upload handler: success scenario (right panel opened, queries invalidated) and failure scenario (error toast shown, overlay dismissed)

## 5. Upload Modal Redesign

- [x] 5.1 Introduce `previewUrl` state (string | null) in `UploadModal`; generate via `URL.createObjectURL` when a file is staged; revoke in `handleClose` and when file is cleared
- [x] 5.2 Replace drop zone with thumbnail preview row when `previewUrl` is set: show `<img src={previewUrl}>` (48px tall, cover), filename, and × remove button
- [x] 5.3 Auto-fill the title input with `fileBaseName(file.name)` when a file is staged (set as value, not placeholder)
- [x] 5.4 Add `detailsOpen` boolean state and a clickable "Add details ▸" / "Add details ▾" toggle below the title input; collapsed by default
- [x] 5.5 Render notes textarea and source URL input inside the collapsible section; both empty by default
- [x] 5.6 Pass `description` (if non-empty) and `source_url` (if non-empty) to `initiateUpload` in the upload mutation
- [x] 5.7 Reset `detailsOpen`, notes, source URL, and `previewUrl` inside `handleClose`
- [x] 5.8 Write unit tests for `UploadModal`: success scenario with "Add details" fields filled (description and source_url sent), failure scenario (modal stays open, error toast shown)
