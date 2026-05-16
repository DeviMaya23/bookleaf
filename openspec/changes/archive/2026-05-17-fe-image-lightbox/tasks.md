## 1. Image API

- [x] 1.1 Add `ImageDetail` interface to `src/lib/images.ts` extending `Image` with `image_url: string`
- [x] 1.2 Add `getImage(getToken, id: string): Promise<ImageDetail>` to `src/lib/images.ts` that calls `GET /images/:id`

## 2. ImageCard

- [x] 2.1 Add `onOpen: (image: Image) => void` prop to `ImageCardProps`
- [x] 2.2 Add `onClick={() => onOpen(image)}` to the inner card `<div>` in `ImageCard`

## 3. Lightbox

- [x] 3.1 Add `lightboxTarget` state (`Image | null`, default `null`) to `ImageGrid`
- [x] 3.2 Add `useQuery(['image', lightboxTarget?.id], ...)` in `ImageGrid` with `enabled: !!lightboxTarget` that calls `getImage`
- [x] 3.3 Pass `onOpen={setLightboxTarget}` to each `ImageCard` in the grid
- [x] 3.4 Add lightbox `Dialog` to `ImageGrid`, controlled by `open={!!lightboxTarget}` and `onOpenChange` that clears `lightboxTarget`
- [x] 3.5 Inside `DialogContent`: render a visually hidden `DialogTitle` with the image title
- [x] 3.6 Inside `DialogContent`: render centered `Loader2` spinner when `isLoading` is true; render `<img>` with `src={imageDetail.image_url}` and `className="max-h-[90vh] max-w-[90vw] object-contain"` when data is available

## 4. Unit Tests

- [x] 4.1 Test `getImage` in `src/lib/images.ts`: success scenario (returns `ImageDetail` with `image_url`) and failure scenario (throws on non-ok response)
- [x] 4.2 Test lightbox in `ImageGrid`: success scenario (clicking a card opens the Dialog and renders the image after fetch resolves) and failure scenario (Dialog closes when `onOpenChange` fires with `false`)
