## Context

`GET /share/:token` and `GET /share/:token/export` already exist and return a folder's name, notes, and images (`title`, `thumbnail_url`, `full_res_url`) with no auth. The frontend has no route for `/share/:token` yet, no lightbox anywhere in the app, and no precedent for a fixed-viewport page outside `AppLayout`/`AuthGuard`.

Two existing pieces this design leans on heavily:
- `MasonryLayout` + `masonry.ts` (`frontend/src/features/gallery/components/`): column-balanced grid driven by `image.width`/`image.height`, typed to `images: Image[]` from `frontend/src/lib/images.ts`.
- `PublicThemeLock` (`frontend/src/components/PublicThemeLock.tsx`): wraps `<Outlet/>` in `data-theme="warm"`, already used for `/`, `/about`, `/privacy`, `/ai-notes`.

## Goals / Non-Goals

**Goals:**
- Render a read-only, unauthenticated gallery view at `/share/:token` matching the warm-theme layout reference (topbar, masonry grid, side panel).
- Reuse `MasonryLayout` for the grid without modifying its existing type signature (it's used by the authenticated gallery's drag-and-drop flow).
- Support double-click → lightbox with prev/next/escape, and a hover download button per card.
- Handle three states: populated, empty folder, invalid/expired token.

**Non-Goals:**
- Any owner-facing controls (rename, notes editing, share toggle/copy) — those stay in `FolderPanelContent`.
- Refreshing presigned URLs mid-session if they expire while the page is open.
- Mobile/responsive layout beyond what falls out naturally from the existing masonry column logic.

## Decisions

### 1. Reuse `MasonryLayout` via a data adapter, not a type change

`MasonryLayout`'s props are `images: Image[]` and `renderCard: (image: Image, imgHeight: number, isDropTarget: boolean) => ReactNode`, and it reads `image.width`, `image.height`, and `image.id` (for the drop-indicator check) directly off the `Image` type from `lib/images.ts`.

**Decision**: in `frontend/src/features/share-viewer/`, map the `GET /share/:token` response's images into the `Image` shape via a small local adapter (`toGalleryImage`-style function), filling fields the share page doesn't have with neutral placeholders:

```ts
{
  id: String(index),        // stable for this static list; no real image id in the response
  title: img.title,
  description: null,
  mime_type: '',
  source_url: null,
  folder_ids: [],
  thumbnail_url: img.thumbnail_url,
  width: img.width,
  height: img.height,
  file_size: null,
  tags: [],
  position: null,
  created_at: '',
  updated_at: '',
}
```

Pass `dropIndicatorId={null}` (always `false` for the drop-indicator check) and `containerWidth` from the same `ResizeObserver` pattern `ImageGrid` uses.

**Alternative considered**: genericize `MasonryLayout`'s props (`images: T[]`, `renderCard: (image: T, ...) => ReactNode` with `T extends { id: string; width: number | null; height: number | null }`). Rejected — `MasonryLayout` is shared with the authenticated gallery's drag-and-drop grid; changing its signature touches code outside this feature's blast radius for a one-page benefit. The adapter is a few lines, local to `share-viewer`, and leaves `MasonryLayout`/`masonry.ts` completely untouched.

### 2. Backend: add `width`/`height` and `download_url` to `SharedImage`/`sharedImageResponse`

`domain.Image.Width`/`Height` (`*int`) already exist and are populated on upload. `usecase.SharedImage` and `handler.sharedImageResponse` gain matching `*int` `width`/`height` fields, set directly from `domain.Image.Width`/`Height` in `GetSharedFolder` — no new queries, no new dependency.

`SharedImage`/`sharedImageResponse` also gain a `DownloadURL`/`download_url` (`string`) field, generated in `GetSharedFolder` via the existing `StorageService.GeneratePresignedDownloadURL(ctx, img.R2Path, filename, presignedGetTTL)` (already used by `image_usecase.go` and `folder_usecase.go` for authenticated downloads), with `filename` derived the same way via the existing `downloadFileExtension(img.MIMEType)` helper. This is one additional presigned-URL call per image on `GetSharedFolder` — see decision 5 for why it's needed.

### 3. New `share-viewer` feature directory, own page component (not nested in `AppLayout`)

```
frontend/src/features/share-viewer/
  components/
    SharePage.tsx           # route element: fetches data, renders one of the 3 states
    SharedFolderPanel.tsx   # read-only side panel
    SharedImageCard.tsx     # renderCard for MasonryLayout: thumbnail + hover download button
    Lightbox.tsx            # fullscreen overlay, prev/next/escape
  lib/
    toGalleryImage.ts       # adapter from decision 1
```

`SharePage` is registered in `App.tsx` under `PublicThemeLock` (it only sets `data-theme="warm"`, imposing no layout, so it's safe to nest a fixed-viewport page under it):

```tsx
<Route element={<PublicThemeLock />}>
  ...
  <Route path="/share/:token" element={<SharePage />} />
</Route>
```

`SharePage` owns a `h-screen flex flex-col` shell (topbar + `flex-1 flex` body with grid/panel), independent of `AppLayout` — it has no sidebar, toolbar, or auth-gated chrome to share with it.

### 4. Data fetching: single `useQuery` driving all three states

```ts
const { data, error, isLoading } = useQuery({
  queryKey: ['share', token],
  queryFn: () => getSharedFolder(token),
  retry: false,
})
```

`getSharedFolder(token)` in `frontend/src/lib/share.ts` calls `apiFetch('/share/...', () => Promise.resolve(undefined))` (no Kinde token — `apiFetch` already omits the `Authorization` header when `getToken()` resolves `undefined`) and throws on non-OK so react-query surfaces it via `error`.

- `isLoading` → simple centered spinner (matches `ImageGrid`'s `Loader2` pattern), no shell yet (avoids a flash of empty topbar/panel before we know if the token is valid).
- `error` (404 or any other failure) → invalid-token state: centered `AlertCircle` (the `text-destructive` pattern from `BatchUploadModal`) + "This link is invalid or has expired" + "Shared via Bookleaf" branding link, no topbar/grid/panel.
- `data` with `images.length === 0` → full shell, image area shows `ImageGrid`'s empty-state pattern ("No images here yet"), `SharedFolderPanel`'s export button disabled.
- `data` with images → full shell, masonry grid + lightbox.

### 5. Hover download button

`SharedImageCard` wraps its content in `className="group relative ..."`. The button is `<Button variant="secondary" size="icon-sm" className="absolute bottom-2 right-2 rounded-full opacity-0 group-hover:opacity-100 transition-opacity">` with a `Download` icon, rendered as `<a href={img.download_url} download>` (via Base UI's render prop, per the existing "no `asChild`" convention).

**Revised from the original plan to use `full_res_url`**: R2 presigned GET URLs don't set `Content-Disposition: attachment`, so an `<a href={full_res_url} download>` doesn't actually trigger a download cross-origin — the browser just navigates to/displays the image. `download_url` (decision 2) is presigned via `GeneratePresignedDownloadURL`, which sets `Content-Disposition: attachment; filename=...`, so the link downloads the file with a sensible filename as intended.

### 6. Lightbox

`Lightbox` takes `images: SharedImage[]` (the raw API shape — `full_res_url` needed, not present on the adapted `Image`) and `initialIndex: number`, manages its own `index` state internally:

- `Escape` closes (window `keydown` listener, mirrors `ImageViewer`'s pattern).
- `ArrowLeft`/`ArrowRight` navigate when `hasPrev`/`hasNext`.
- Click on the backdrop closes; click on the image/controls does not (via `stopPropagation`).
- Renders `full_res_url` directly in an `<img>`, plus close/prev/next `Button`s (`size="icon-sm"`, `rounded-full`) and a counter/caption, styled per the design reference but using theme tokens rather than the mock's hardcoded dark colors — confirm look matches the mock's intent (overlay is dark regardless of the page's "warm" theme, same as the in-app `ImageViewer`).

`SharePage` tracks `lightboxIndex: number | null`; `SharedImageCard`'s `onDoubleClick` sets it to that card's index into the original (un-adapted) `images` array.

## Risks / Trade-offs

- **Presigned URL TTL**: `full_res_url` (and `thumbnail_url`) are presigned with `presignedGetTTL`. A recipient who keeps the page open longer than the TTL gets broken images/downloads on next interaction. → Accepted for this change; no refetch-on-expiry mechanism. Worth a follow-up if it proves to be a real issue.
- **Round-robin column distribution**: `MasonryLayout` assigns `item[i] -> column[i % numCols]`, not height-balanced. With the design mock's varied aspect ratios, columns can end up visibly uneven. → Inherited from the existing gallery; consistent UX, not a regression.
- **Synthetic `id` in the adapter**: using `String(index)` as `Image.id` is only safe because this list is static (no reordering/dnd on the share page) and `MasonryLayout` only uses `id` for the drop-indicator comparison, which is always `false` here (`dropIndicatorId={null}`).

## Migration Plan

Additive only — new route, new feature directory, new backend response fields (nullable, ignored by existing consumers). No data migration, no rollback concerns beyond a normal revert.
