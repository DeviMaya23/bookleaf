## 1. Implementation

- [x] 1.1 Add a small helper in `FolderPanelContent.tsx` to format the image count label ("1 image" / "{n} images", including "0 images")
- [x] 1.2 Render the label as a subtitle below the folder name input in the panel header, only when `folderDetail` has loaded

## 2. Testing

- [x] 2.1 Update `FolderPanelContent.test.tsx`: subtitle shows "{n} images" for counts other than 1, "1 image" for a count of 1, "0 images" for an empty folder, and is absent while the folder detail query is loading

## 3. Verification

- [x] 3.1 Run `npm run build` and `npm run lint` in `frontend/` and fix any issues
