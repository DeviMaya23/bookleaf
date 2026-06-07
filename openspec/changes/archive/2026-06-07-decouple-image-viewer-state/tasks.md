## 1. Introduce independent viewer state

- [x] 1.1 In `AppLayout.tsx`, replace the `viewerOpen: boolean` state with `viewerImage: Image | null`
- [x] 1.2 Update `handleImageDoubleClick` to set `viewerImage` to the double-clicked image (alongside the existing `setSelectedImage`)
- [x] 1.3 Update the viewer's render condition (`viewerOpen && selectedImage`) to `viewerImage !== null`, passing `viewerImage` as the `image` prop and `() => setViewerImage(null)` as `onClose`

## 2. Remove the cross-coupling effect

- [x] 2.1 Remove the `useEffect(() => { if (!selectedImage) setViewerOpen(false) }, [selectedImage])` effect
- [x] 2.2 Verify the right panel's `onClose` (`setSelectedImage(null)`) no longer affects `viewerImage`/the viewer, and that the viewer's containing area widens via the existing flex layout once the panel unmounts

## 3. Dismiss viewer and right panel on folder navigation

- [x] 3.1 Add a `useEffect` keyed on the active view/folder identifier (derived from `useAppView()`) that resets `viewerImage`, `selectedImage`, and `autoFocusTitle` to their closed/null states
- [x] 3.2 Verify navigating to a different folder while the viewer and/or right panel is open returns to that folder's gallery grid with both closed

## 4. Close viewer when its image is deleted

- [x] 4.1 Extend the `onImageDeleted` handler to also check `viewerImage?.id === id` and reset `viewerImage` to `null` when it matches, independent of the existing `selectedImage` check
- [x] 4.2 Verify deleting the image open in the viewer closes the viewer even when the right panel is showing a different image

## 5. Verification

- [x] 5.1 Manually verify all scenarios in the `fe-image-viewer` delta spec (double-click open, back/Esc dismissal, right-panel-close survival + widening, folder-navigation dismissal, deletion handling)
- [x] 5.2 Run `npm run build` and fix any issues that arise
