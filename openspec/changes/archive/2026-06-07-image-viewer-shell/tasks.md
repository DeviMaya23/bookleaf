## 1. ImageViewer component

- [x] 1.1 Create `frontend/src/components/ImageViewer.tsx` — accepts `image: Image` and `onClose: () => void` props; render a flex-column container filling its parent with no background color specified
- [x] 1.2 Add the toolbar: 44px fixed-height row containing back button (chevron-left icon), zoom range slider (min=5, max=800), zoom percentage label (static "100%" for now), separator, flip horizontal button, rotate 90° CW button, 1:1 button, flex spacer — all buttons non-functional except back
- [x] 1.3 Add the canvas area: flex-1, overflow hidden, image centered with `object-contain`; show `thumbnail_url` as placeholder while full-res loads
- [x] 1.4 Fetch full-res URL via `useQuery(['image', image.id], () => getImage(getToken, image.id))`; replace thumbnail with `imageDetail.image_url` once resolved
- [x] 1.5 Add filename + dimensions badge: absolutely positioned bottom-center overlay showing `{image.title} · {image.width} × {image.height}`
- [x] 1.6 Wire Esc key to `onClose` via `useEffect` keydown listener
- [x] 1.7 Wire back button `onClick` to `onClose`

## 2. ImageGrid — add onImageDoubleClick

- [x] 2.1 Add `onImageDoubleClick?: (image: Image) => void` to `ImageGridProps` interface
- [x] 2.2 Pass `onImageDoubleClick` down to `ImageCard` as a prop
- [x] 2.3 Add `onDoubleClick` handler on the card's root element that calls both `onSelect` and `onDoubleClick` prop

## 3. AppLayout wiring

- [x] 3.1 Add `viewerOpen` boolean state to `AppLayout`, initialised to `false`
- [x] 3.2 Add `useEffect` to reset `viewerOpen` to `false` when `selectedImage` becomes `null` (covers the case where the user closes the right panel while the viewer is open)
- [x] 3.3 Add `handleImageDoubleClick` handler: sets `selectedImage` to the double-clicked image and sets `viewerOpen` to `true`
- [x] 3.4 Pass `onImageDoubleClick={handleImageDoubleClick}` to `ImageGrid`
- [x] 3.5 In `<main>`, conditionally render `<ImageViewer image={selectedImage} onClose={() => setViewerOpen(false)} />` when `viewerOpen && selectedImage`, otherwise render the existing `<ScrollArea>` + `<ImageGrid>` block

## 4. Remove Dialog lightbox from RightPanel

- [x] 4.1 Remove `lightboxOpen` state and the `setLightboxOpen(false)` call in the `image.id` reset effect
- [x] 4.2 Remove `onClick={() => setLightboxOpen(true)}` from the thumbnail wrapper div; keep the div and thumbnail `<img>` as non-interactive
- [x] 4.3 Remove the `Dialog`, `DialogContent`, `DialogTitle` JSX block and unused imports (`Dialog`, `DialogContent`, `DialogTitle` from `@/components/ui/dialog`)

## 5. Unit tests for ImageViewer

- [x] 5.1 Create `frontend/src/components/ImageViewer.test.tsx` with mocks for `@kinde-oss/kinde-auth-react` and `@/lib/images`
- [x] 5.2 Test: thumbnail is shown while full-res URL is loading (`getImage` pending — assert `thumbnail_url` src is rendered)
- [x] 5.3 Test: full-res image is shown once `getImage` resolves (`image_url` src is rendered, thumbnail replaced)
- [x] 5.4 Test: pressing Esc calls `onClose`
- [x] 5.5 Test: clicking the back button calls `onClose`
- [x] 5.6 Test: badge renders image title and dimensions

## 6. Build check

- [x] 6.1 Run `npm run build` from `frontend/` and fix any TypeScript or lint errors that arise
