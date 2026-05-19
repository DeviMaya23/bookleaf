## 1. Backend — Trash Pagination Fix

- [ ] 1.1 Add `DeletedAt *time.Time` field to `ImageCursor` struct in `image_pagination.go`
- [ ] 1.2 Add `deleted_at` field to `cursorPayload` JSON struct (optional, omitempty) in `image_pagination.go`
- [ ] 1.3 Update `EncodeCursor` to include `DeletedAt` when non-nil
- [ ] 1.4 Update `DecodeCursor` to populate `DeletedAt` from JSON payload
- [ ] 1.5 Update `ListTrashed` in `image_repository.go` to sort by `deleted_at ASC, id ASC`
- [ ] 1.6 Update cursor condition in `ListTrashed` repository to use `(deleted_at, id) > (?, ?)` with `cursor.DeletedAt` and `cursor.ID`
- [ ] 1.7 Update `ListTrashed` in `image_usecase.go` to build `NextCursor` with `DeletedAt: &last.DeletedAt.Time`
- [ ] 1.8 Write unit tests for `ListTrashed` usecase (success: returns oldest-deleted-first; failure: repo error)
- [ ] 1.9 Write integration test for `ListTrashed` repository (assert `deleted_at ASC` ordering)

## 2. Frontend Lib — API Functions

- [ ] 2.1 Add optional `parentId?: string` parameter to `createFolder` in `folders.ts`, include `parent_id` in the POST body when provided
- [ ] 2.2 Add `getAllImages` function to `images.ts` (calls `GET /images` with no filter params)
- [ ] 2.3 Add `getTrashedImages` function to `images.ts` (calls `GET /images/trash`, supports cursor param)
- [ ] 2.4 Add `restoreImage` function to `images.ts` (calls `POST /images/:id/restore`)

## 3. Frontend Routing & Layout

- [ ] 3.1 Add `/unsorted` and `/trash` routes to `App.tsx`
- [ ] 3.2 Define `AppView` discriminated union type: `{ type: 'all' } | { type: 'unsorted' } | { type: 'trash' } | { type: 'folder'; id: string }`
- [ ] 3.3 Update `AppLayout.tsx` to derive `AppView` from the current route and pass it to `ImageGrid`

## 4. Frontend — ImageGrid Trash Mode

- [ ] 4.1 Update `ImageGrid` props to accept `view: AppView` instead of `folderId: string | null`
- [ ] 4.2 Select query key and fetch function based on `view.type` (all/unsorted/folder → `getImages` variants; trash → `getTrashedImages`)
- [ ] 4.3 In trash mode, replace the "Delete" context menu item with a "Restore" item that calls `restoreImage`
- [ ] 4.4 On successful restore, invalidate the trash query and show a success toast
- [ ] 4.5 On restore failure, show an error toast

## 5. Frontend — Sidebar Rewrite

- [ ] 5.1 Write `buildFolderTree(folders: Folder[]): FolderNode[]` utility in `FolderSidebar.tsx` (flat `parent_id` list → nested tree, with cycle guard)
- [ ] 5.2 Implement recursive `FolderItem` component with depth-based indentation (`paddingLeft: 8 + depth * 14`) and expand/collapse toggle
- [ ] 5.3 Add system entries section: All (→ `/`), Unsorted (→ `/unsorted`), Trash (→ `/trash`, muted color)
- [ ] 5.4 Add "FOLDERS" section label and horizontal divider between system entries and user folder tree
- [ ] 5.5 Add "New subfolder" item to the folder context menu, opening `FolderNameDialog` with `parentId` pre-set
- [ ] 5.6 Wire "New subfolder" confirm to `createFolder` with `parentId`, then invalidate folders query
- [ ] 5.7 Update active folder highlight to work with `AppView` (match system routes or `folder.id`)
- [ ] 5.8 Adjust sidebar footer padding to match design; confirm `ProfileMenu` renders correctly in the new layout

## 6. Bruno API File

- [ ] 6.1 Create Bruno request file for `GET /images/trash` (list trashed images)
- [ ] 6.2 Create Bruno request file for `POST /images/:id/restore` (restore image)
