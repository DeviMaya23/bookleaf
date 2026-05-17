## Context

`ImageGrid.tsx` already imports and uses shadcn `Dialog` for the delete confirmation flow. It manages `deleteTarget` state in `ImageGrid` and passes an `onDelete` callback down to `ImageCard`. The pattern is: state lives in `ImageGrid`, cards are stateless. `images.ts` has `getImages` and `deleteImage` but no `getImage` function. The backend `GET /images/:id` handler already generates a presigned GET URL and returns it as `image_url` in `imageDetailResponse`.

## Goals / Non-Goals

**Goals:**
- Left-click on `ImageCard` opens a lightbox Dialog
- `ImageGrid` fetches `GET /images/:id` on card click to get the presigned URL
- Spinner shown while fetching; image shown once URL resolves
- Lightbox dismissable via X button, backdrop click, or ESC
- Image constrained to viewport (`max-h-[90vh] max-w-[90vw] object-contain`)

**Non-Goals:**
- Showing metadata (title, description, dimensions) in the lightbox
- Keyboard navigation between images
- Download button
- Zoom / pan interaction

## Decisions

### 1. Lightbox state lives in `ImageGrid`, not `ImageCard`

`lightboxTarget` state (`Image | null`) follows the same pattern as `deleteTarget`. Only one `Dialog` is mounted in `ImageGrid`; `ImageCard` fires an `onOpen` callback. This avoids mounting one Dialog per card and keeps `ImageCard` stateless.

**Alternative considered**: State local to each `ImageCard`. Rejected — causes N Dialog instances mounted simultaneously and breaks the established pattern.

### 2. `useQuery` with `enabled: !!lightboxTarget`

`getImage(id)` is fetched via `useQuery` keyed on `['image', lightboxTarget.id]`. The query is disabled until a target is set. TanStack Query's default `staleTime: 0` means a refetch on every open — appropriate since presigned URLs expire.

**Alternative considered**: Fetch imperatively on click (no query cache). Rejected — misses deduplication if the user double-clicks, and doesn't benefit from the cache on rapid re-opens within the same session.

### 3. Dialog opens immediately; spinner shown while loading

The Dialog is opened as soon as the user clicks (controlled by `!!lightboxTarget`). The image area shows a centered `Loader2` spinner while `isLoading` is true, then swaps to the `<img>` once the URL is available.

**Alternative considered**: Wait for the URL to resolve before opening the Dialog. Rejected — gives no immediate feedback to the user on slow connections.

### 4. `DialogTitle` rendered visually hidden

shadcn's `DialogContent` requires an accessible title. The image's `title` from the already-loaded `Image` object is passed as a visually hidden `DialogTitle` (via Radix `VisuallyHidden`). This satisfies the accessibility requirement without showing text in the lightbox UI.

### 5. No new UI dependencies

shadcn `Dialog` is already imported in `ImageGrid`. Radix `VisuallyHidden` is a transitive dependency of shadcn and available without an additional install.

## Risks / Trade-offs

- **Presigned URL expiry**: If a user has the gallery open for longer than the URL TTL (set in the usecase layer), re-opening a cached image detail could serve an expired URL. Mitigation: `staleTime: 0` ensures a fresh fetch on every open; the cache only deduplicates within a single click event.
- **ContextMenu + click conflict**: `ImageCard` is currently a `ContextMenuTrigger`. Left-click (lightbox) and right-click (context menu) are distinct pointer events — no conflict. Confirmed by shadcn ContextMenu behaviour.
