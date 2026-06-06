## 1. API Layer

- [x] 1.1 Add `hardDeleteImage(getToken, id)` to `lib/images.ts` calling `DELETE /images/trash/:id`
- [x] 1.2 Add `emptyTrash(getToken)` to `lib/images.ts` calling `DELETE /images/trash`

## 2. Single Permanent Delete (ImageGrid)

- [x] 2.1 Add `onDeletePermanent?: (image: Image) => void` prop to `ImageCard`
- [x] 2.2 Import `ContextMenuSeparator` in `ImageGrid.tsx`
- [x] 2.3 Update the trash context menu in `ImageCard` to render "Restore", a separator, and "Delete permanently" (destructive colour) — "Delete permanently" calls `onDeletePermanent`
- [x] 2.4 Add `confirmDeleteImage` state (`Image | null`) and confirmation `Dialog` in `ImageGrid`
- [x] 2.5 Add `hardDeleteMutation` in `ImageGrid` — on success, remove image from `orderedImages`, invalidate `['images', 'trash']`, show success toast; on error show error toast
- [x] 2.6 Wire `onDeletePermanent` on `ImageCard` to set `confirmDeleteImage`; confirming the dialog fires `hardDeleteMutation`

## 3. Empty Trash (FolderSidebar)

- [x] 3.1 Wrap the Trash `div` in `FolderSidebar` with `ContextMenu`, `ContextMenuTrigger`, and `ContextMenuContent` containing an "Empty trash" `ContextMenuItem` (destructive colour)
- [x] 3.2 Add `confirmEmptyTrash` boolean state and confirmation `Dialog` in `FolderSidebar` — "Empty trash" menu item sets it to `true`
- [x] 3.3 Import `Dialog`, `DialogContent`, `DialogFooter`, `DialogHeader`, `DialogTitle` in `FolderSidebar` (already imported — verify)
- [x] 3.4 Add `emptyTrashMutation` in `FolderSidebar` — on success, invalidate `['images', 'trash']`, show success toast; on error show error toast

## 4. Bruno

- [x] 4.1 Create `DELETE Permanently Delete Image.bru` in the trash requests folder for `DELETE /images/trash/:id`
- [x] 4.2 Create `DELETE Empty Trash.bru` in the trash requests folder for `DELETE /images/trash`

## 5. Unit Tests

- [x] 5.1 Test `ImageGrid` in trash view: right-clicking an image shows "Restore" and "Delete permanently" in the context menu
- [x] 5.2 Test `ImageGrid`: confirming permanent delete calls `hardDeleteMutation` and removes image from view
- [x] 5.3 Test `ImageGrid`: dismissing the confirmation dialog makes no API call
- [x] 5.4 Test `FolderSidebar`: right-clicking the Trash entry shows "Empty trash" context menu item
- [x] 5.5 Test `FolderSidebar`: confirming "Empty trash" calls `emptyTrashMutation`
