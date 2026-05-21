## 1. Manifest

- [ ] 1.1 Add `"contextMenus"` and `"notifications"` to `permissions` in `extensions/manifest.json`

## 2. API Client

- [ ] 2.1 Create `src/lib/api.ts` with `apiFetch(path, options?)` — reads token from `getAuth()`, attaches `Authorization: Bearer`, prepends `VITE_API_BASE_URL`

## 3. Background Service Worker

- [ ] 3.1 Register context menu item on `chrome.runtime.onInstalled` in `src/background/index.ts`: `{ id: "save-to-bookleaf", title: "Save to Bookleaf", contexts: ["image"] }`
- [ ] 3.2 Add `chrome.contextMenus.onClicked` listener that extracts `srcUrl`, `pageUrl`, and `tab.title`
- [ ] 3.3 Implement `isTokenValid(auth)` guard — returns false if null or `Date.now() > auth.expiresAt`
- [ ] 3.4 Implement `fetchImageBlob(url)` — fetches image URL, returns `{ blob, mimeType }` or throws
- [ ] 3.5 Implement `saveImage({ blob, mimeType, title, pageUrl })` — drives the 3-step upload: `POST /images` → `PUT` blob to presigned URL → `POST /images/:id/complete`
- [ ] 3.6 Wire the click handler: check token → fetch blob → saveImage → show success/failure notification

## 4. Notifications

- [ ] 4.1 Implement `notify(title, message)` helper in `src/background/index.ts` using `chrome.notifications.create`
- [ ] 4.2 Show "Saved to Bookleaf!" notification on successful upload
- [ ] 4.3 Show "Please log in first" notification when token is missing or expired
- [ ] 4.4 Show "Save failed. Please try again." notification on any fetch or upload error

## 5. Verification

- [ ] 5.1 Run `npm run type-check` and confirm zero TypeScript errors
- [ ] 5.2 Run `npm run build` and confirm `dist/manifest.json` includes `contextMenus` and `notifications`
- [ ] 5.3 Load unpacked extension in Chrome, right-click an image, confirm "Save to Bookleaf" appears
- [ ] 5.4 Save an image while logged in and confirm it appears in Bookleaf with correct title and source URL
- [ ] 5.5 Log out via popup, attempt save, confirm "Please log in first" notification appears
- [ ] 5.6 Attempt save on a broken/unreachable image URL and confirm "Save failed" notification appears
