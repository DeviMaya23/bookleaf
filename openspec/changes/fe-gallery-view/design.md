## Context

The frontend already has `AppLayout` (sidebar + main area), `FolderSidebar` (folder list with react-query), and a stub `ImageGrid` component. Routing exists for `/`, `/login`, and `/callback` via react-router-dom. The app uses `@tanstack/react-query` for data fetching and `useKindeAuth` for auth tokens. There is no active route for `/folders/:folderId` and no API call wired to the image grid yet.

## Goals / Non-Goals

**Goals:**
- Wire `ImageGrid` to `GET /images` with a `folderId` query param derived from the current route
- Add `/folders/:folderId` route that renders the same `AppLayout` with the correct folder context
- Make `FolderSidebar` folder items navigate to the correct route on click
- Show spinner while fetching, empty state when no results, and image cards otherwise
- Image cards show thumbnail and truncated title; right-click exposes delete with a confirmation dialog
- Responsive grid: 6 columns desktop, 2 columns mobile

**Non-Goals:**
- Pagination (handled by a separate spec `image-list-pagination`)
- Upload flow
- Drag-and-drop between folders
- Image detail/preview modal

## Decisions

### 1. Route-driven folder context (URL as source of truth)

Folder selection will be encoded in the URL (`/` for unfoldered, `/folders/:folderId` for a specific folder). `AppLayout` reads `useParams` to get `folderId` and passes it to `ImageGrid`.

**Alternative considered**: Store selected folder in React state or Zustand. Rejected — URL state is shareable, survives refresh, and avoids a new global state dependency.

### 2. Single `AppLayout` rendered on both routes

`App.tsx` will render `<AppLayout />` on both `/` and `/folders/:folderId`. The layout reads the param internally, keeping route configuration minimal.

**Alternative considered**: Separate page components per route. Rejected — they would be identical except for the `folderId` source, creating unnecessary duplication.

### 3. `FolderSidebar` uses `useNavigate` for folder clicks

Clicking a folder calls `navigate('/folders/' + folder.id)`. The "Unsorted" item navigates to `/`. Active state is derived from `useLocation`.

### 4. TanStack Query key includes `folderId`

Query key: `['images', folderId ?? null]`. This gives automatic cache separation per folder, and invalidating `['images']` clears all image caches (useful for post-delete).

### 5. Image API module at `src/lib/images.ts`

Mirrors the existing `src/lib/folders.ts` pattern: plain async functions that accept `getToken`. Keeps components thin and functions unit-testable.

### 6. Delete confirmation uses existing Radix `Dialog` component

Matches the pattern already used in `FolderSidebar` for folder deletion. No new UI dependencies.

## Risks / Trade-offs

- **Thumbnail URL availability** → The image list response must include a `thumbnailUrl` (or equivalent) field. If the backend returns a storage path rather than a public URL, a mapping layer will be needed. Mitigation: confirm with backend spec before implementing the card component.
- **`folderId=null` vs omitted** → Backend `GET /images` expects `folderId=null` as an explicit query param for unfoldered images. The API function must send `?folderId=null` (string) rather than omitting the param. Mitigation: encode explicitly in the `getImages` function.
- **Stale image list after delete** → After a successful delete, `queryClient.invalidateQueries({ queryKey: ['images'] })` refreshes the current view. This is intentionally broad to handle edge cases where folder context changes between the delete trigger and success callback.
