## REMOVED Requirements

### Requirement: Clicking the right panel thumbnail opens the lightbox overlay
**Reason**: The Dialog-based lightbox is replaced by the `fe-image-viewer` capability. The viewer is triggered by double-clicking an image card in the gallery, not from the right panel thumbnail.
**Migration**: Double-click an image card in the gallery to open the full-res viewer.

### Requirement: Lightbox fetches the presigned URL on open
**Reason**: URL fetching is now handled by `ImageViewer` via the existing React Query key `['image', image.id]`.
**Migration**: No migration needed; behavior is preserved in the new viewer component.

### Requirement: Lightbox is dismissable via multiple interactions
**Reason**: The Dialog-based lightbox is removed. The new viewer is dismissed via the back button in the toolbar or the Esc key.
**Migration**: Use the back button or Esc key to close the image viewer.

### Requirement: Lightbox shows image only, no metadata
**Reason**: Superseded by the new image viewer, which shows the image in the main content area alongside the metadata right panel.
**Migration**: No migration needed.
