## Context

The current full-res image viewer is a shadcn `Dialog` component inside `RightPanel`. It opens when the user clicks the panel thumbnail, shows a spinner while fetching the presigned `image_url`, then renders `<img>` constrained to `max-h-[90vh] max-w-[90vw]`. It has no controls, no zoom, no pan. The Dialog is a stopgap — it gives up most of the screen to the backdrop and doesn't belong in the panel.

The replacement is a dedicated `ImageViewer` component that occupies the full main content area, swapping out `ImageGrid` when active. The right panel stays visible alongside it.

## Goals / Non-Goals

**Goals:**
- Replace the Dialog lightbox with a full-panel viewer component
- Trigger viewer on double-click of any gallery image card
- Display the full-res image centered in the main area with correct aspect ratio
- Render the complete toolbar chrome (back, zoom slider/%, flip, rotate, 1:1 buttons) as a static shell
- Support Esc and back button to return to the gallery
- Remove the Dialog lightbox and thumbnail click handler from `RightPanel`

**Non-Goals:**
- Pan, zoom, rotate, flip interactions (deferred to follow-up change)
- Prev/next image navigation inside the viewer
- Any keyboard shortcut beyond Esc

## Decisions

### Viewer replaces gallery in `<main>`, not an overlay

The viewer lives in the same slot as `ImageGrid`. `AppLayout` renders either `<ImageGrid>` or `<ImageViewer>` depending on `viewerOpen` state:

```
<main>
  {viewerOpen ? <ImageViewer ... /> : <ImageGrid ... />}
</main>
```

**Why**: The right panel should remain visible alongside the viewer — metadata is useful while inspecting an image. An overlay that covers the panel would require the user to close the viewer just to edit a tag. Swapping in `<main>` keeps the layout stable and is architecturally simple: no wrapper div changes, no z-index layering.

### State owned by `AppLayout`

`AppLayout` gains two new state fields: `viewerOpen: boolean` and a viewer is opened with the existing `selectedImage` (already in `AppLayout`). No new image-identity state is needed — when the user double-clicks, the card is also selected (the double-click should select + open viewer together).

The `viewerOpen` flag is set to `false` when the user closes the viewer, selects a different image, or navigates away.

### `ImageViewer` fetches full-res URL itself via React Query

`ImageViewer` receives the base `Image` prop (same shape `RightPanel` receives) and calls `useQuery(['image', image.id])` to get `image_url`. Since `RightPanel` makes the same query when the image is selected, the React Query cache will typically serve it immediately — no double-fetch.

While the URL is loading, the viewer renders the thumbnail as a placeholder (same pattern as `RightPanel`). This avoids a blank frame on open.

### Toolbar is rendered as static chrome in this phase

All toolbar buttons are rendered with correct icons and layout but no interaction logic beyond the back button and Esc. The zoom slider is rendered but does nothing. This establishes the final visual structure so the follow-up change only needs to wire state — no layout work.

### No background color on viewer container

The viewer container inherits `background` from the page rather than specifying a color. This ensures it respects any future theme implementation without requiring changes to the component.

### `onImageDoubleClick` as a separate prop on `ImageGrid`

`ImageGrid` receives a new `onImageDoubleClick?: (image: Image) => void` callback alongside the existing `onImageSelect`. Double-clicking a card calls both: `onImageSelect` (to select the image, opening the panel) and `onImageDoubleClick` (to open the viewer). Keeping them separate avoids coupling selection logic to viewer logic and preserves the existing single-click behavior unchanged.

## Risks / Trade-offs

- **Full-res fetch on open**: `useQuery(['image', image.id])` fires when `ImageViewer` mounts. If the right panel was never opened for this image before double-clicking, there is no cached URL and a fetch is triggered. The thumbnail is shown immediately as a placeholder, so the viewer never appears blank. → Acceptable; matches existing lightbox behavior.

- **Viewer state lost on image switch**: If the user selects a different image from the gallery (which they can't reach while the viewer is open), the viewer closes. This is intentional — the viewer is tied to `selectedImage` in `AppLayout`.

- **Static toolbar may mislead**: Rendering non-functional controls could confuse users who try to interact with them. → Acceptable for an in-progress implementation; the follow-up change adds interactions immediately.

## Open Questions

None outstanding. Resolved decisions:
- Double-clicking an already-selected image opens the viewer.
- Closing the viewer sets `viewerOpen = false` only. `selectedImage` in `AppLayout` is untouched, so the right panel remains open until the user explicitly dismisses it.
