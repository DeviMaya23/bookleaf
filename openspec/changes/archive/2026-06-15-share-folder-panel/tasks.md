## 1. Share API wrapper

- [x] 1.1 Create `frontend/src/lib/share.ts` with `getFolderShare(getToken, folderId)` calling `GET /folders/:id/share`, returning `{ token: string } | null` (404 → `null`, other non-ok → throw)
- [x] 1.2 Add `createFolderShare(getToken, folderId)` calling `POST /folders/:id/share`, returning `{ token: string }`
- [x] 1.3 Add `deleteFolderShare(getToken, folderId)` calling `DELETE /folders/:id/share`, returning `void`

## 2. Disable-share confirm dialog

- [x] 2.1 Create `frontend/src/features/right-panel/components/DisableShareDialog.tsx`, modeled on `DeleteFolderDialog`, with `open`, `onCancel`, `onConfirm` props and copy warning that the existing link will stop working and re-enabling generates a new one

## 3. FolderPanelContent integration

- [x] 3.1 Add `useQuery(['folder-share', folder.id], () => getFolderShare(getToken, folder.id))` to `FolderPanelContent`
- [x] 3.2 Add "Share folder" section between the Notes section and the Export footer, using the same bordered-row layout as Notes, containing a `Switch` labeled "Share folder" with `checked = !!shareData`
- [x] 3.3 Add `createShareMutation` (calls `createFolderShare`, invalidates `['folder-share', folder.id]` on success) and wire switch-on (while off) to trigger it directly
- [x] 3.4 Add `deleteShareMutation` (calls `deleteFolderShare`, invalidates `['folder-share', folder.id]` on success); wire switch-off (while on) to open `DisableShareDialog` instead of mutating directly, and trigger the mutation only on dialog confirm
- [x] 3.5 When `shareData` is present, render a read-only truncated input showing `${window.location.origin}/share/${shareData.token}` plus a copy icon button
- [x] 3.6 Implement copy button: `navigator.clipboard.writeText(url)` with `toast.success('Link copied')` on success and `toast.error('Failed to copy link')` on failure
- [x] 3.7 Disable the switch while the share query is loading or either mutation is pending

## 4. Tests

- [x] 4.1 In `FolderPanelContent.test.tsx`, mock `@/lib/share` and add scenarios: switch off when `getFolderShare` resolves `null`; switch on + link field shown when it resolves `{ token }`
- [x] 4.2 Add scenario: toggling the switch on calls `createFolderShare` and the link field appears on success
- [x] 4.3 Add scenario: toggling the switch off opens the confirm dialog; confirming calls `deleteFolderShare` and hides the link field; cancelling makes no call and leaves the switch on
- [x] 4.4 Add scenario: clicking the copy button calls `navigator.clipboard.writeText` with the full share URL

## 5. Verification

- [x] 5.1 Run `npm run build` and `npm run lint` in `frontend/` and fix any issues
