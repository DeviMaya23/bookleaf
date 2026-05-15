## 1. Routing

- [x] 1.1 Add `/folders/:folderId` route to `App.tsx` rendering `<AppLayout />`
- [x] 1.2 Update `AppLayout.tsx` to read `folderId` from `useParams` and pass it to `ImageGrid`

## 2. Folder Sidebar Navigation

- [x] 2.1 Add `useNavigate` to `FolderSidebar` and navigate to `/folders/<id>` on folder item click
- [x] 2.2 Add `useLocation` to `FolderSidebar` and apply active styles to the currently selected folder / "Unsorted" item
- [x] 2.3 Wire "Unsorted" item to navigate to `/` on click

## 3. Image API

- [x] 3.1 Create `src/lib/images.ts` with `getImages(getToken, folderId: string | null, cursor?: string)` that calls `GET /images?folderId=<value>[&cursor=<cursor>]` (send `"null"` string when no folder)
- [x] 3.2 Add `deleteImage(getToken, id: string)` to `src/lib/images.ts` that calls `DELETE /images/:id`

## 4. Image Grid

- [x] 4.1 Use `useInfiniteQuery(['images', folderId ?? null], ...)` in `ImageGrid`, passing `cursor` from `pageParam` on subsequent pages
- [x] 4.2 Render a spinner while `isLoading` is true (initial load)
- [x] 4.3 Render empty state ("🖼 No images here yet") when all pages combined yield zero images
- [x] 4.4 Apply responsive grid classes: `grid-cols-2` (mobile) and `lg:grid-cols-6` (desktop)
- [x] 4.5 Build `ImageCard` component: renders thumbnail `<img>` and title with `truncate` class
- [x] 4.6 Show "Load more" button below grid when the last page's `next_cursor` is non-null; hide it on the last page
- [x] 4.7 Wire "Load more" button to `fetchNextPage`; show a loading indicator on the button while `isFetchingNextPage` is true

## 5. Delete Flow

- [x] 5.1 Wrap each `ImageCard` in a `ContextMenu` with a "Delete" `ContextMenuItem`
- [x] 5.2 On "Delete" select, open a confirmation `Dialog` showing the image title
- [x] 5.3 On confirm, call `deleteImage` mutation and invalidate `['images']` query on success
- [x] 5.4 On cancel, close dialog with no side effects

## 6. Unit Tests

- [x] 6.1 Test `ImageGrid`: success scenario (renders cards from mock data) and failure scenario (shows empty state when all pages are empty)
- [x] 6.2 Test `ImageGrid` pagination: success scenario ("Load more" button appears when `next_cursor` is non-null) and failure scenario (button hidden when `next_cursor` is null)
- [x] 6.3 Test delete flow in `ImageGrid`: success scenario (calls delete and invalidates) and failure scenario (dialog cancel makes no request)
