## 1. Backend — PATCH /images/:id: add source_url

- [x] 1.1 Add `SourceURL **string` to `UpdateImageParams` in `backend/internal/usecase/image_usecase.go`
- [x] 1.2 Add `source_url *string` field to `updateImageRequest` struct in `backend/internal/handler/image.go` using pointer-of-pointer decoding pattern (consistent with `folder_id`)
- [x] 1.3 Map `SourceURL` from request to usecase params in `UpdateImage` handler
- [x] 1.4 Update usecase `UpdateImage` to include `source_url` in the fields map when `SourceURL` param is non-nil
- [x] 1.5 Update handler unit tests: add success scenario for source_url update, verify existing failure scenario still passes
- [x] 1.6 Update usecase unit tests: add success scenario for source_url update
- [x] 1.7 Update Bruno file for `PATCH /images/:id` to include `source_url` in the request body example

## 2. Frontend — lib/images.ts: add updateImage and downloadImage

- [x] 2.1 Add `updateImage(getToken, id, params: { title?: string, description?: string | null, source_url?: string | null })` function to `src/lib/images.ts` calling `PATCH /images/:id`
- [x] 2.2 Add `downloadImage(getToken, id)` function to `src/lib/images.ts` calling `GET /images/:id/download` and returning the `download_url`

## 3. Frontend — AppLayout: lift state and add panel slot

- [x] 3.1 Add `selectedImage: Image | null` state to `AppLayout`
- [x] 3.2 Pass `onImageSelect` callback to `ImageGrid` and remove the card-click-to-lightbox logic from `ImageGrid`
- [x] 3.3 Render `<RightPanel>` in the layout flex row when `selectedImage` is non-null, passing `image`, `onClose`

## 4. Frontend — ImageGrid: masonry layout

- [x] 4.1 Replace the CSS grid (`grid grid-cols-2 lg:grid-cols-6`) with CSS column-count masonry (`columns-2 md:columns-3 lg:columns-4`)
- [x] 4.2 Update `ImageCard` to remove the fixed `aspect-square` wrapper; thumbnail should render at natural aspect ratio using `w-full h-auto`
- [x] 4.3 Remove border from `ImageCard` (remove `border` class)
- [x] 4.4 Add `break-inside-avoid mb-3` to each card wrapper
- [x] 4.5 Update `ImageCard` `onClick` to call `onImageSelect(image)` instead of opening lightbox
- [x] 4.6 Remove lightbox `Dialog` and its associated state (`lightboxTarget`, `imageDetail` query) from `ImageGrid`

## 5. Frontend — RightPanel component

- [x] 5.1 Create `src/components/RightPanel.tsx` with props `{ image: Image, onClose: () => void }`
- [x] 5.2 Render thumbnail at top using `image.thumbnail_url` with `w-full h-auto object-cover`, with ✕ close button overlaid (top-right, `position: absolute`)
- [x] 5.3 Clicking the thumbnail sets local `lightboxOpen` state to `true`; render lightbox `Dialog` (with `GET /images/:id` query) inside `RightPanel` when `lightboxOpen` is true
- [x] 5.4 Render title as an editable ghost input below thumbnail; auto-save via `updateImage` on blur when value changed and non-empty; revert to previous value if user clears the field; show success/error toast
- [x] 5.5 Render description as an editable textarea; auto-save via `updateImage` on blur when value changed; show success/error toast
- [x] 5.6 Render source URL input with "Open ↗" button; auto-save via `updateImage` on blur when value changed; show success/error toast
- [x] 5.7 "Open ↗" button opens `source_url` in new tab (`target="_blank" rel="noreferrer"`); visually disabled when empty
- [x] 5.8 Render 2-column details grid: formatted file size, dimensions (width × height), folder name (or "Unsorted"), added date
- [x] 5.9 Render sticky "Download image" footer button; on click, call `downloadImage`, then trigger browser download via temporary anchor; show disabled/loading state during fetch
- [x] 5.10 Panel is `w-80 flex-shrink-0 border-l h-screen overflow-y-auto`

## 6. Frontend — unit tests

- [x] 6.1 Write unit test for `RightPanel`: success scenario — renders title, notes, source URL, details, and download button for a given image
- [x] 6.2 Write unit test for `RightPanel`: failure scenario — download button shows error state when `downloadImage` rejects
