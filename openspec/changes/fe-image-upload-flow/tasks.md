## 1. Setup

- [ ] 1.1 Install `sonner` package
- [ ] 1.2 Add `<Toaster />` from `sonner` to `main.tsx` (or `App.tsx`) at app root

## 2. API Layer

- [ ] 2.1 Add `initiateUpload(getToken, params)` to `lib/images.ts` — calls `POST /images`, returns `{ id, upload_url, r2_path }`
- [ ] 2.2 Add `putToR2(uploadUrl, file)` to `lib/images.ts` — PUT file bytes to presigned URL with correct `Content-Type`
- [ ] 2.3 Add `completeUpload(getToken, id)` to `lib/images.ts` — calls `POST /images/:id/complete`, returns `{ image_id, suggested_folder_name, warning }`
- [ ] 2.4 Add `acceptSuggestion(getToken, id, suggestedFolderName)` to `lib/images.ts` — calls `POST /images/:id/accept-suggestion`

## 3. UploadModal Component

- [ ] 3.1 Create `src/components/UploadModal.tsx` with props: `open`, `onOpenChange`, `folderId: string | null`
- [ ] 3.2 Implement drop zone UI — drag-and-drop area that accepts file drop; shows filename when a file is staged; shows inline error for invalid types
- [ ] 3.3 Implement file picker — clicking the drop zone triggers a hidden `<input type="file" accept="image/jpeg,image/png,image/gif,image/webp" />` 
- [ ] 3.4 Implement file type validation on both drop and picker selection — reject non JPEG/PNG/GIF/WEBP with an inline error
- [ ] 3.5 Implement title input field — optional, placeholder shows selected filename without extension, value is not auto-filled
- [ ] 3.6 Implement submit handler using `useMutation` — chains `initiateUpload → putToR2 → completeUpload`; title falls back to filename (no extension) if blank; passes `folderId` as `folder_id` when non-null
- [ ] 3.7 Implement submit button loading and disabled state during mutation
- [ ] 3.8 Implement success path (no suggestion) — close modal, `toast.success(...)`, invalidate `['images']` query
- [ ] 3.9 Implement error path — `toast.error(...)`, keep modal open
- [ ] 3.10 Implement suggestion view — when `completeUpload` returns a non-null `suggested_folder_name`, replace form body with suggestion view showing the name and Accept / Ignore buttons
- [ ] 3.11 Implement Accept action — call `acceptSuggestion`, close modal, `toast.success(...)`
- [ ] 3.12 Implement Ignore action — close modal, `toast.success(...)` (no API call)

## 4. AppLayout Wiring

- [ ] 4.1 Add `+ Image` button to top-right of the main content area in `AppLayout.tsx`
- [ ] 4.2 Add `uploadOpen` state to `AppLayout.tsx` and wire it to the button and `<UploadModal />`
- [ ] 4.3 Pass `folderId` (from `useParams`, null-coerced when undefined) as prop to `<UploadModal />`
