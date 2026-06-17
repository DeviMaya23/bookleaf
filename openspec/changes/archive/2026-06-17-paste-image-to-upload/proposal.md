## Why

Users often copy images from other apps or take OS screenshots and want to save them to Bookleaf quickly. There is no keyboard-driven path today — every upload requires reaching for the mouse to use the button or drag-and-drop.

## What Changes

- A global `paste` event listener is added to `AppLayout` that fires when the user presses CTRL+V / CMD+V anywhere in the app
- If the clipboard contains an image file and no text input or textarea is focused, the single-file upload modal opens with the image pre-loaded
- The `UploadModal` gains an `initialFile` prop; when set, the file is staged immediately and the title field is left blank and auto-focused so the user is prompted to name it
- If the clipboard image has a generic name (e.g. `image.png` — typical for OS screenshots), the title field shows an empty value rather than "image"

## Capabilities

### New Capabilities
- `fe-paste-image-upload`: Global clipboard paste handler that detects image payloads and opens the upload modal pre-filled, with title blank and auto-focused

### Modified Capabilities
- `fe-image-upload-flow`: `UploadModal` gains an `initialFile` prop and title auto-focus behaviour when opened via paste

## Impact

- `frontend/src/app-shell/AppLayout.tsx` — adds `useEffect` paste listener, threads `initialFile` and `autoFocusTitle` state into `UploadModal`
- `frontend/src/features/upload/components/UploadModal.tsx` — new `initialFile?: File` prop, auto-focus title input on open when initialFile is set
- No backend changes
- No extension changes
