## 1. Backend: Migrations

- [x] 1.1 Create migration `000017_add_folder_icon` (`folders.icon TEXT`, nullable) with up/down SQL
- [x] 1.2 Create migration `000018_add_folder_icons_enabled_to_users` (`users.folder_icons_enabled BOOLEAN NOT NULL DEFAULT true`) with up/down SQL
- [x] 1.3 Run migrations locally and confirm both columns exist with correct defaults

## 2. Backend: Icon allowlist

- [x] 2.1 Define the 55-key icon allowlist as a Go constant (`map[string]struct{}` or equivalent) in the folder package, per `specs/folder-icon-customization/spec.md`
- [x] 2.2 Add a validation helper (e.g. `IsValidFolderIcon(key string) bool`) used by the usecase layer

## 3. Backend: Domain & repository

- [x] 3.1 Add `Icon *string` field (`gorm:"column:icon"`) to `Folder` struct in `internal/domain/folder.go`
- [x] 3.2 Add `FolderIconsEnabled bool` field (`gorm:"column:folder_icons_enabled;default:true"`) to `User` struct in `internal/domain/user.go`
- [x] 3.3 Confirm `folderRepository.Update`'s existing `map[string]any` plumbing requires no changes to support an `icon` key (verify via integration test in 3.4)
- [x] 3.4 Add/extend `folder_repository_integration_test.go` case(s) covering `Update` writing the `icon` column

## 4. Backend: Usecase

- [x] 4.1 Add `Icon *string` param to `FolderUsecase.Create` and `Icon **string` to `UpdateFolderParams`
- [x] 4.2 In `Create` and `Update`, validate any provided `Icon` value against the allowlist; return an invalid-icon error if not present
- [x] 4.3 In `Update`, include `icon` in the selective fields map only when `params.Icon` is non-nil
- [x] 4.4 Add `UpdateMeParams`-equivalent support (or extend existing `User` update path) for `FolderIconsEnabled`, mirroring how `VisionEnabled` is currently updated via `PATCH /me`
- [x] 4.5 Unit tests: `folder_usecase_test.go` — icon allowlist rejection on `Create` and `Update` (assert specific error type), icon passthrough when valid, icon omitted from fields map when not provided
- [x] 4.6 Unit tests: user/me usecase — `folder_icons_enabled` toggled true/false persists and is returned in result; non-boolean/empty-body rejection if validated at this layer

## 5. Backend: Handlers

- [x] 5.1 Add `Icon *string` (or `json.RawMessage` for partial-update presence detection) to `folderRequest`, `updateFolderRequest`, `folderResponse`, and `folderDetailResponse` in `internal/handler/folder.go`
- [x] 5.2 Wire `icon` parsing in `CreateFolder` and `UpdateFolder`, following the existing `description` presence-detection pattern; return `400` on allowlist validation failure
- [x] 5.3 Add `FolderIconsEnabled` to the `PATCH /me` request struct and `GET /me`/`PATCH /me` response struct in the me handler; update request validation so at least one of `vision_enabled`/`folder_icons_enabled` must be present
- [x] 5.4 Unit tests: `folder_handler_test.go` — request/response include `icon`; `400` returned for non-allowlisted icon
- [x] 5.5 Unit tests: me handler — `folder_icons_enabled` round-trips through `GET`/`PATCH /me`; `400` on empty body or non-boolean value

## 6. Backend: API contracts (bruno)

- [x] 6.1 Update `bruno/folders/create-folder.bru` to include an example `icon` field
- [x] 6.2 Update `bruno/folders/update-folder.bru` to include an example `icon` field (and a request setting it to `null`)
- [x] 6.3 Update `bruno/folders/list-folders.bru` / `bruno/folders/get-folder.bru` expected response examples to include `icon`
- [x] 6.4 Update `bruno/update-me.bru` to include an example `folder_icons_enabled` field
- [x] 6.5 Update `bruno/me.bru` expected response example to include `folder_icons_enabled`

## 7. Backend: Lint

- [x] 7.1 Run `golangci-lint run` and fix any issues introduced by the above changes

## 8. Frontend: Types & API client

- [x] 8.1 Add `icon: string | null` to the `Folder` interface and `UpdateFolderParams` in `frontend/src/lib/folders.ts`; add `icon?: string | null` to the create-folder params
- [x] 8.2 Add `folder_icons_enabled: boolean` to the `Me`/user type and update params in the me API client module
- [x] 8.3 Create a `FOLDER_ICONS` module: icon-key-to-lucide-component map covering the same 55 keys as the backend allowlist, plus the default (`folder`) and fixed system-entry icons (`file-stack`, `file-question-mark`, `trash-2`)

## 9. Frontend: Sidebar rendering

- [x] 9.1 In `FolderItem.tsx`, render the folder's icon (looked up via `FOLDER_ICONS`, falling back to default) between the expand arrow and the folder name, gated on `folder_icons_enabled`
- [x] 9.2 In `UnsortedEntry.tsx` and `TrashEntry.tsx`, render their fixed icons, gated on `folder_icons_enabled`
- [x] 9.3 In `FolderSidebar.tsx`, render the fixed icon for the "All" entry, gated on `folder_icons_enabled`
- [x] 9.4 Source `folder_icons_enabled` from the current user query (`GET /me`) at the point where the sidebar/folder list is rendered

## 10. Frontend: Change-icon context menu

- [x] 10.1 Add a "Change icon" `ContextMenuSub` item to `FolderItem.tsx`'s existing `ContextMenuContent`, listing all 55 icons via `ContextMenuSubContent`
- [x] 10.2 Wire icon selection to call the existing folder update mutation with `{ icon: <key> }`, following the same optimistic/refetch pattern used for rename
- [x] 10.3 Confirm no "Change icon" option is exposed on the All/Unsorted/Trash context menus (Trash already has one for "Empty trash" — must not be conflated)

## 11. Frontend: Settings toggle

- [x] 11.1 Add a `Switch` for `folder_icons_enabled` to `AppSection.tsx`, mirroring the `useMutation`/`updateMe` pattern in `AdvancedSection.tsx`'s vision toggle
- [x] 11.2 Verify the switch disables while the mutation is in flight and reflects the persisted value from the response without a separate refetch

## 12. Frontend: Build & lint

- [x] 12.1 Run `npm run build` and fix any type errors
- [x] 12.2 Run `npm run lint` and fix any issues introduced by the above changes

## 13. Manual verification

- [x] 13.1 Start the app, change a folder's icon via right-click, confirm it persists across reload
- [x] 13.2 Confirm All/Unsorted/Trash show their fixed icons and offer no "Change icon" option
- [x] 13.3 Toggle "Folder icons" off in settings, confirm all icons disappear and names shift left with no layout gap; toggle back on
