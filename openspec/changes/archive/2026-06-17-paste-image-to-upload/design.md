## Context

`AppLayout` already owns all upload modal state (`uploadOpen`, `batchUploadOpen`, `batchInitialFiles`) and has a drag-and-drop handler for file drops on the main area. The single-file drop path auto-uploads silently; the paste path will instead open the modal pre-filled so the user can title the image first.

`UploadModal` currently has no way to receive a pre-selected file from outside — it manages file state internally. It also has no auto-focus on the title field.

OS screenshots arrive in the clipboard as a `File` with a generic name like `image.png`, so `fileBaseName` would produce `"image"` — not a useful default title.

## Goals / Non-Goals

**Goals:**
- Intercept CTRL+V / CMD+V globally when an image is in the clipboard
- Open `UploadModal` with the image pre-staged and the title field blank and focused
- Do nothing when the paste target is a text input or textarea (preserve normal text paste)
- Do nothing when the clipboard has no image

**Non-Goals:**
- Auto-uploading on paste (intentionally choosing the modal path for discoverability and titling)
- Handling multiple images from clipboard
- Any backend or extension changes

## Decisions

### 1. Where to attach the paste listener

**Decision:** `useEffect` on `document` inside `AppLayout`, cleaned up on unmount.

A document-level listener captures paste regardless of which element is focused (except guarded inputs — see below). This mirrors how the drag-and-drop overlay works: `AppLayout` is the single owner of upload state, so it is the right place to open the modal.

Alternative considered: a custom hook (`useClipboardPaste`) — not needed for one call site; adding indirection without a second consumer would be premature.

### 2. Guard against text input targets

**Decision:** Check `event.target` before acting. If `target` is an `<input>` or `<textarea>`, return early and let the browser handle text paste normally.

```
const tag = (event.target as HTMLElement).tagName
if (tag === 'INPUT' || tag === 'TEXTAREA') return
```

This covers the title field, notes textarea, and search input inside the open app. `contenteditable` is not used in the codebase so no additional guard is needed.

### 3. How to pass the file into UploadModal

**Decision:** Add an `initialFile?: File` prop to `UploadModal`. When the modal opens with `initialFile` set, a `useEffect` calls the existing `handleFile(initialFile)` to stage it. `AppLayout` holds an `uploadInitialFile` state (`File | null`) alongside the existing `uploadOpen` boolean.

Alternative considered: thread the file through `batchInitialFiles` and open `BatchUploadModal` — rejected because this is a single-image paste path and the single-file modal has the title/notes/source-url form the user needs.

### 4. Title field behaviour on paste-open

**Decision:** Leave the title blank (do not pre-fill with the generic filename). Auto-focus the title `<input>` via the Base UI `Dialog.Popup`'s `initialFocus` prop when `initialFile` is provided.

`fileBaseName("image.png")` → `"image"` is noise the user would have to delete. A blank focused field communicates clearly that naming is expected. The title field's `placeholder` is made dynamic (`file ? fileBaseName(file.name) : 'Title'`) so the user still sees the fallback that will be used if they submit without typing.

A manually-called `titleInputRef.current?.focus()` inside the staging `useEffect` was tried first but did not stick: Base UI's `Dialog.Popup` runs its own focus management after the re-render triggered by staging the file, which moved focus to the first focusable element in the newly-rendered file preview (the "Remove file" button), overriding the manual call. The fix is to pass `initialFocus={initialFile ? titleInputRef : undefined}` to `DialogContent`, which lets Base UI itself focus the title input as part of its managed open-focus behaviour instead of fighting it.

### 5. Clipboard image extraction

**Decision:** Iterate `event.clipboardData.items`, find the first item where `kind === 'file'` and `type.startsWith('image/')`, then call `item.getAsFile()`. If no such item exists, return early with no action.

No need to call `navigator.clipboard.read()` (the async Clipboard API) — the synchronous `clipboardData` on the `paste` event is sufficient and avoids the permission prompt that the async API triggers in some browsers.

## Risks / Trade-offs

- **UploadModal already open** → if the user presses CTRL+V while the modal is already open, the guard `if (uploadOpen) return` prevents re-opening. The image in the clipboard is ignored. This is the least surprising behaviour; the user can drag the image into the modal's drop zone instead.
- **Non-image clipboard** → silently ignored (no toast). Adding a toast for "no image in clipboard" would be noisy; paste is a habitual gesture.
- **Very large clipboard images** → no size limit is checked at the paste stage. The existing `validateImageFile` inside `handleFile` handles type validation; size limits are enforced server-side on upload.
