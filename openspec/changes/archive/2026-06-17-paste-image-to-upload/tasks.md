## 1. UploadModal — initialFile prop and auto-focus

- [x] 1.1 Add `initialFile?: File` to `UploadModalProps` in `UploadModal.tsx`
- [x] 1.2 Add a `useEffect` that calls `handleFile(initialFile)` when `initialFile` changes and the modal is open
- [x] 1.3 Add a `ref` for the title `<input>` and attach it
- [x] 1.4 Auto-focus the title input when `initialFile` is set: inside the same `useEffect` (or a separate one keyed on `open`), call `titleInputRef.current?.focus()` after staging the file

## 2. AppLayout — paste listener and state

- [x] 2.1 Add `uploadInitialFile` state (`File | null`, default `null`) to `AppLayout`
- [x] 2.2 Add a `useEffect` that attaches a `paste` handler to `document` and removes it on cleanup
- [x] 2.3 Implement the paste handler: guard on `INPUT`/`TEXTAREA` target, guard if `uploadOpen` is already true, extract the first image item from `event.clipboardData.items`, call `item.getAsFile()`
- [x] 2.4 On a valid image: set `uploadInitialFile` to the file and set `uploadOpen` to `true`
- [x] 2.5 Pass `initialFile={uploadInitialFile}` to `<UploadModal />`
- [x] 2.6 Reset `uploadInitialFile` to `null` when `uploadOpen` changes to `false` (in the `onOpenChange` handler)

## 3. Unit tests

- [x] 3.1 Test `UploadModal` with `initialFile` set: assert file is staged and title input is focused on open
- [x] 3.2 Test `UploadModal` without `initialFile`: assert title input is not focused on open

## 4. Lint and build

- [x] 4.1 Run `npm run lint` in `frontend/` and fix any issues
- [x] 4.2 Run `npm run build` in `frontend/` and fix any issues
