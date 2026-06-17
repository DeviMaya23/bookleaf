## Why

Folders are currently all identically unlabeled in the sidebar, making it hard to visually scan a long folder list. Users want to assign a recognizable icon to their own folders, and the three pinned system entries (All, Unsorted, Trash) should also show a fixed icon for visual consistency.

## What Changes

- Add an `icon` column to the `folders` table (nullable string; falls back to a default `folder` icon when unset).
- Add a server-side enum/allowlist of 55 valid icon keys; `PATCH /folders/:id` validates the `icon` field against this allowlist and rejects unknown values.
- Add a "Change icon" submenu to the folder right-click context menu, listing the allowlisted icons; selecting one updates the folder's icon.
- Render the folder's icon (or the default) to the left of the folder name in the sidebar tree.
- Render fixed, non-editable icons next to the "All" (`file-stack`), "Unsorted" (`file-question-mark`), and "Trash" (`trash-2`) system entries.
- Add a `folder_icons_enabled` boolean column to the `users` table (default `true`), exposed via `GET /me` / `PATCH /me`.
- Add a `Switch` in the SettingsModal's App section to toggle `folder_icons_enabled`. When disabled, no folder/system entry renders an icon (name shifts left; no layout side effects).

## Capabilities

### New Capabilities
- `folder-icon-customization`: Lets users pick an icon for their own folders from a curated allowlist via the right-click context menu; persists to the folder record; renders in the sidebar; includes fixed icons for system entries and the user-level enable/disable toggle.

### Modified Capabilities
- `folder-domain`: `Folder` struct gains a nullable `Icon` field.
- `folder-endpoints`: `POST /folders` and `PATCH /folders/:id` request/response bodies gain an `icon` field; update validates against the server-side allowlist.
- `fe-sidebar-nav`: Folder tree items and the three system entries (All, Unsorted, Trash) render an icon to the left of their label, controlled by the `folder_icons_enabled` preference.
- `user-domain`: `User` struct gains a `FolderIconsEnabled` boolean field, default `true`.
- `me-endpoint`: `GET /me` / `PATCH /me` response and update body gain `folder_icons_enabled`.

## Impact

- Backend: new migration (`folders.icon`, `users.folder_icons_enabled`), `internal/domain/folder.go`, `internal/domain/user.go`, `internal/handler/folder.go`, `internal/handler/me.go` (or equivalent), `internal/usecase/folder_usecase.go`, a new allowlist constant/validator.
- Frontend: `lib/folders.ts` (types), `FolderItem.tsx` (icon render + context submenu), `UnsortedEntry.tsx`, `TrashEntry.tsx`, `FolderSidebar.tsx` (All entry), new icon-key → lucide-component map, `AppSection.tsx` (toggle, following the `fe-vision-toggle` pattern), `lib/me.ts`-equivalent API client.
- No breaking changes; both new fields are additive and nullable/defaulted.
