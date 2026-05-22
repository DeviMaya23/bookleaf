## 1. Storage helpers

- [ ] 1.1 Add `RecentSave` interface and `bookleaf_recent_saves` key to `storage.ts`
- [ ] 1.2 Implement `getRecentSaves(): Promise<RecentSave[]>` — returns stored array or `[]`
- [ ] 1.3 Implement `addRecentSave(entry: RecentSave): Promise<void>` — prepend + slice to 5
- [ ] 1.4 Implement `getDarkMode(): Promise<boolean>` — returns stored value or `false`
- [ ] 1.5 Implement `setDarkMode(value: boolean): Promise<void>`
- [ ] 1.6 Add `bookleaf_username` key; implement `getUsername(): Promise<string | null>` and `setUsername(name: string): Promise<void>`
- [ ] 1.7 Add `bookleaf_avatar` key; implement `getAvatar(): Promise<string | null>` and `setAvatar(url: string): Promise<void>`

## 2. Auth — store username at login

- [ ] 2.1 Capture `id_token` from the token exchange response in `exchangeCodeForTokens`
- [ ] 2.2 Add a `decodeJwtPayload(token: string): Record<string, unknown>` utility in `auth.ts` that base64url-decodes the middle segment of the JWT
- [ ] 2.3 After `setAuth(auth)` in `login()`, decode the `id_token` payload; extract `given_name` → `name` → `email` and call `setUsername()`; extract `picture` and call `setAvatar()` if present

## 3. Background worker — thumbnail generation

- [ ] 3.1 Add `generateThumbnail(blob: Blob): Promise<string>` in `background/index.ts` — uses `createImageBitmap`, 60×60 `OffscreenCanvas` with cover crop, `convertToBlob(jpeg, 0.7)`, chunked `btoa` conversion
- [ ] 3.2 After the success notification in `handleSave`, call `generateThumbnail` and `addRecentSave` wrapped in try/catch — failures are logged only and do not affect the save result

## 4. Popup — redesign App.tsx

- [ ] 4.1 Update `index.html` body width from 280px to 320px
- [ ] 4.2 On mount, read `auth`, `username`, `avatar`, `recentSaves`, and `darkMode` in a single `Promise.all`; show blank container until resolved
- [ ] 4.3 Implement logged-out layout: header (icon + wordmark), centered body (dimmed icon, tagline, full-width "Log in to Bookleaf" CTA, "New here? Sign up free" footer line)
- [ ] 4.4 Implement logged-in layout: header with "Open ↗" button (`browser.tabs.create` with `VITE_APP_URL`), user row (`<img>` avatar with gradient fallback, username, dark mode toggle)
- [ ] 4.5 Implement recently saved thumbnail strip: horizontal strip of up to 5 `<img>` elements using stored `dataUrl`
- [ ] 4.6 Implement empty state: dimmed icon + "Nothing saved yet." + "Right-click any image to save it."
- [ ] 4.7 Implement dark/light mode palette swap; wire toggle to `setDarkMode` and local state update
- [ ] 4.8 Wire "Log out" footer button to `clearAuth()` + transition to logged-out state

## 5. Bruno file

- [ ] 5.1 No new API endpoints are introduced — no Bruno file needed for this change
